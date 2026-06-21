// Copyright 2022 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

// Darwin windowing, cgo-free via purego/objc (companion to the cgo-free Metal in
// gpu/mtl and the CAMetalLayer in gpu/ctx/ca). It creates an NSApplication,
// an NSWindow, and a Metal-backed NSView subclass (PolyredView) entirely through
// the Objective-C runtime, and pumps main-thread work via libdispatch's
// dispatch_async_f. No cgo; builds with CGO_ENABLED=0.
//
// Milestone status (specs/foundations/cgo-free-windowed-present.md, brick 2):
// window creation + Metal present are ported; mouse/keyboard event delivery
// (the NSView event-method IMPs) is stubbed and lands in a follow-up (M3).
package app

import (
	"fmt"
	"image"
	"os"
	"runtime"
	"sync"
	"time"
	"unsafe"

	"github.com/ebitengine/purego"
	"github.com/ebitengine/purego/objc"

	"poly.red/gpu/ctx/ca"
	"poly.red/gpu/mtl"
	"poly.red/math"
)

func init() {
	runtime.LockOSThread()
}

// init loads the Cocoa frameworks so their classes (NSApplication, NSWindow,
// NSView, CAMetalLayer) are registered with the Objective-C runtime. With cgo
// gone, nothing else links them, so objc.GetClass would return a nil class and
// subclassing (RegisterClass) would produce a class with no real superclass.
func init() {
	for _, fw := range []string{
		"/System/Library/Frameworks/AppKit.framework/AppKit",
		"/System/Library/Frameworks/QuartzCore.framework/QuartzCore",
		"/System/Library/Frameworks/Metal.framework/Metal",
	} {
		if _, err := purego.Dlopen(fw, purego.RTLD_NOW|purego.RTLD_GLOBAL); err != nil {
			panic(fmt.Errorf("app: dlopen %s: %w", fw, err))
		}
	}
}

// --- libdispatch (main-queue pump) --------------------------------------------

var (
	dispatchGetMainQueue func() uintptr
	dispatchAsyncF       func(queue uintptr, ctx uintptr, work uintptr)
)

func init() {
	lib, err := purego.Dlopen("/usr/lib/libSystem.dylib", purego.RTLD_NOW|purego.RTLD_GLOBAL)
	if err != nil {
		panic(fmt.Errorf("app: dlopen libSystem: %w", err))
	}
	purego.RegisterLibFunc(&dispatchAsyncF, lib, "dispatch_async_f")
	// dispatch_get_main_queue is a static inline in C; the underlying symbol is
	// the global _dispatch_main_q. Take its address.
	q, err := purego.Dlsym(lib, "_dispatch_main_q")
	if err != nil {
		panic(fmt.Errorf("app: dlsym _dispatch_main_q: %w", err))
	}
	mainQ := q
	dispatchGetMainQueue = func() uintptr { return mainQ }
}

// dispatchWork is the C callback invoked on the main thread by dispatch_async_f.
var dispatchWork = purego.NewCallback(func(_ uintptr) uintptr {
	polyredDispatchMainFuncs()
	return 0
})

func wakeupMainThread() {
	dispatchAsyncF(dispatchGetMainQueue(), 0, dispatchWork)
}

// --- selectors ----------------------------------------------------------------

var (
	selAlloc                 = objc.RegisterName("alloc")
	selInit                  = objc.RegisterName("init")
	selNew                   = objc.RegisterName("new")
	selRelease               = objc.RegisterName("release")
	selSharedApplication     = objc.RegisterName("sharedApplication")
	selSetActivationPolicy   = objc.RegisterName("setActivationPolicy:")
	selActivateIgnoringOther = objc.RegisterName("activateIgnoringOtherApps:")
	selSetDelegate           = objc.RegisterName("setDelegate:")
	selRun                   = objc.RegisterName("run")
	selFinishLaunching       = objc.RegisterName("finishLaunching")
	selNextEvent             = objc.RegisterName("nextEventMatchingMask:untilDate:inMode:dequeue:")
	selSendEvent             = objc.RegisterName("sendEvent:")
	selDateWithInterval      = objc.RegisterName("dateWithTimeIntervalSinceNow:")
	selStringWithUTF8        = objc.RegisterName("stringWithUTF8String:")
	selInitWithContentRect   = objc.RegisterName("initWithContentRect:styleMask:backing:defer:")
	selSetContentView        = objc.RegisterName("setContentView:")
	selMakeFirstResponder    = objc.RegisterName("makeFirstResponder:")
	selMakeKeyAndOrderFront  = objc.RegisterName("makeKeyAndOrderFront:")
	selSetTitle              = objc.RegisterName("setTitle:")
	selSetContentSize        = objc.RegisterName("setContentSize:")
	selSetContentMinSize     = objc.RegisterName("setContentMinSize:")
	selSetContentMaxSize     = objc.RegisterName("setContentMaxSize:")
	selSetAcceptsMouseMoved  = objc.RegisterName("setAcceptsMouseMovedEvents:")
	selSetReleasedWhenClosed = objc.RegisterName("setReleasedWhenClosed:")
	selInitWithFrame         = objc.RegisterName("initWithFrame:")
	selSetWantsLayer         = objc.RegisterName("setWantsLayer:")
	selLayer                 = objc.RegisterName("layer")
	selSetContentsScale      = objc.RegisterName("setContentsScale:")
	selSetLayerDelegate      = objc.RegisterName("setDelegate:")
	selBackingScaleFactor    = objc.RegisterName("backingScaleFactor")
	selWindow                = objc.RegisterName("window")
	selObject                = objc.RegisterName("object")
	selContentView           = objc.RegisterName("contentView")
)

// Objective-C struct types passed by value (HFA: doubles in FP registers).
type cgPoint struct{ x, y float64 }
type cgSize struct{ width, height float64 }
type cgRect struct {
	origin cgPoint
	size   cgSize
}

const (
	nsWindowStyleTitled         = 1 << 0
	nsWindowStyleClosable       = 1 << 1
	nsWindowStyleMiniaturizable = 1 << 2
	nsWindowStyleResizable      = 1 << 3
	nsBackingStoreBuffered      = 2
	nsActivationPolicyRegular   = 0
)

// --- Objective-C classes (created once) ---------------------------------------

var (
	classPolyredView    objc.Class
	classAppDelegate    objc.Class
	classWindowDelegate objc.Class
)

func registerClasses() {
	var err error

	// PolyredView: a Metal-backed NSView. makeBackingLayer returns a CAMetalLayer;
	// displayLayer drives a redraw. (Event methods are added in M3.)
	classPolyredView, err = objc.RegisterClass(
		"PolyredView", objc.GetClass("NSView"), nil, nil,
		[]objc.MethodDef{
			{Cmd: objc.RegisterName("makeBackingLayer"), Fn: func(self objc.ID, _ objc.SEL) objc.ID {
				layer := objc.ID(objc.GetClass("CAMetalLayer")).Send(selAlloc).Send(selInit)
				layer.Send(selSetLayerDelegate, self)
				return layer
			}},
			{Cmd: objc.RegisterName("displayLayer:"), Fn: func(self objc.ID, _ objc.SEL, layer objc.ID) {
				scale := objc.Send[float64](objc.ID(self).Send(selWindow), selBackingScaleFactor)
				if scale == 0 {
					scale = 1
				}
				layer.Send(selSetContentsScale, scale)
				viewOnDraw(self)
			}},
		},
	)
	if err != nil {
		panic(fmt.Errorf("app: register PolyredView: %w", err))
	}

	classAppDelegate, err = objc.RegisterClass(
		"PolyredAppDelegate", objc.GetClass("NSObject"), nil, nil,
		[]objc.MethodDef{
			// No-op: the launch handshake + activation is driven explicitly from
			// polyredMain (finishLaunching), so this only exists for the
			// NSApplicationDelegate protocol.
			{Cmd: objc.RegisterName("applicationDidFinishLaunching:"), Fn: func(_ objc.ID, _ objc.SEL, _ objc.ID) {
			}},
		},
	)
	if err != nil {
		panic(fmt.Errorf("app: register PolyredAppDelegate: %w", err))
	}

	classWindowDelegate, err = objc.RegisterClass(
		"PolyredWindowDelegate", objc.GetClass("NSObject"), nil, nil,
		[]objc.MethodDef{
			{Cmd: objc.RegisterName("windowWillClose:"), Fn: func(_ objc.ID, _ objc.SEL, notification objc.ID) {
				window := notification.Send(selObject)
				view := window.Send(selContentView)
				viewOnClose(view)
			}},
		},
	)
	if err != nil {
		panic(fmt.Errorf("app: register PolyredWindowDelegate: %w", err))
	}
}

// --- event callbacks (from the Obj-C runtime, on the main thread) -------------

func viewOnDraw(view objc.ID) {
	mu.Lock()
	w := windows[view]
	mu.Unlock()
	if w == nil {
		return
	}
	// Non-blocking: this runs on the main thread; never block it on the draw
	// goroutine. A skipped resize just defers the size update one frame.
	select {
	case w.resize <- resizeEvent{w.win.config.size.X, w.win.config.size.Y}:
	default:
	}
}

func viewOnClose(view objc.ID) {
	mu.Lock()
	w := windows[view]
	delete(windows, view)
	mu.Unlock()
	if w != nil {
		view.Send(selRelease)
		w.win.window.Send(selRelease)
		w.win.ctx.Release()
	}
	os.Exit(0)
}

var launched = make(chan struct{})

// --- window state -------------------------------------------------------------

var (
	nsApp               objc.ID
	globalWindowDel     objc.ID
	registerClassesOnce sync.Once
)

var (
	mu      sync.Mutex
	windows = map[objc.ID]*window{} // view -> window
)

type osWindow struct {
	view   objc.ID // PolyredView*
	window objc.ID // NSWindow*
	ctx    *mtlContext

	viewScale   float32
	screenScale int
	config      *config
}

func (w *window) run(app Window, cfg config, opts ...Option) {
	<-launched

	w.win = &osWindow{config: &cfg}
	runOnMainSync(func() {
		w.win.view = createView()
		if w.win.view == 0 {
			panic("app: failed to create view")
		}

		mu.Lock()
		windows[w.win.view] = w
		mu.Unlock()

		w.win.viewScale = 1
		w.win.screenScale = 1

		canResize := false
		if _, ok := app.(ResizeHandler); ok {
			canResize = true
		}
		w.win.window = createWindow(w.win.view, canResize)
		w.configs(opts...)
		w.win.window.Send(selMakeKeyAndOrderFront, objc.ID(0))

		layerPtr := unsafe.Pointer(w.win.view.Send(selLayer))
		var err error
		w.win.ctx, err = newMtlContext(w.win.config, ca.NewMetalLayer(layerPtr))
		if err != nil {
			panic(fmt.Errorf("app: failed to use Metal: %w", err))
		}
		close(w.ready)
	})
}

func createView() objc.ID {
	view := objc.ID(classPolyredView).Send(selAlloc).Send(selInitWithFrame, cgRect{})
	view.Send(selSetWantsLayer, uint64(1))
	return view
}

func createWindow(view objc.ID, canResize bool) objc.ID {
	style := uint64(nsWindowStyleTitled | nsWindowStyleMiniaturizable | nsWindowStyleClosable)
	if canResize {
		style |= nsWindowStyleResizable
	}
	rect := cgRect{size: cgSize{640, 480}}
	window := objc.ID(objc.GetClass("NSWindow")).Send(selAlloc).
		Send(selInitWithContentRect, rect, style, uint64(nsBackingStoreBuffered), uint64(0))
	window.Send(selSetAcceptsMouseMoved, uint64(1))
	window.Send(selSetContentView, view)
	window.Send(selMakeFirstResponder, view)
	window.Send(selSetReleasedWhenClosed, uint64(0))
	window.Send(selSetDelegate, globalWindowDel)
	return window
}

func nsString(s string) objc.ID {
	b := append([]byte(s), 0)
	return objc.ID(objc.GetClass("NSString")).Send(selStringWithUTF8, &b[0])
}

func (w *window) configs(opts ...Option) {
	cfg := w.win.config
	for _, o := range opts {
		o(cfg)
	}
	w.win.window.Send(selSetContentSize, cgSize{float64(cfg.size.X), float64(cfg.size.Y)})
	if cfg.minSize.X > 0 || cfg.minSize.Y > 0 {
		w.win.window.Send(selSetContentMinSize, cgSize{float64(cfg.minSize.X), float64(cfg.minSize.Y)})
	}
	if cfg.maxSize.X > 0 || cfg.maxSize.Y > 0 {
		w.win.window.Send(selSetContentMaxSize, cgSize{float64(cfg.maxSize.X), float64(cfg.maxSize.Y)})
	}
	w.win.config.title = cfg.title
	w.win.window.Send(selSetTitle, nsString(cfg.title))
}

func (w *window) main(app Window) {
	go w.event(app)
	go w.draw(app)
	polyredMain()
}

func polyredMain() {
	registerClassesOnce.Do(registerClasses)
	nsApp = objc.ID(objc.GetClass("NSApplication")).Send(selSharedApplication)
	nsApp.Send(selSetActivationPolicy, uint64(nsActivationPolicyRegular))
	del := objc.ID(classAppDelegate).Send(selNew)
	nsApp.Send(selSetDelegate, del)
	globalWindowDel = objc.ID(classWindowDelegate).Send(selNew)
	nsApp.Send(selFinishLaunching)
	nsApp.Send(selActivateIgnoringOther, uint64(1))
	close(launched)

	// A manual Cocoa event loop. [NSApp run] does not block reliably when driven
	// from Go via purego (it returns immediately). We pump events ourselves and,
	// each pass, drain funcQ (the main-thread work queue that window creation and
	// present are scheduled on) so we do not depend on the dispatch wake. The
	// short timeout keeps it polling (~500 Hz). This is the on-screen-load-bearing
	// piece.
	const nsEventMaskAny = ^uint64(0)
	defaultMode := nsString("kCFRunLoopDefaultMode")
	nsDate := objc.GetClass("NSDate")
	for {
		polyredDispatchMainFuncs()
		until := objc.ID(nsDate).Send(selDateWithInterval, float64(0.002))
		ev := nsApp.Send(selNextEvent, nsEventMaskAny, until, defaultMode, uint64(1))
		if ev != 0 {
			nsApp.Send(selSendEvent, ev)
		}
	}
}

// The Event Thread
//
// The ticker ticks every ~1ms which permits a maximum of 960 fps for input
// events handling, keeping the window responsive (especially on macOS).
func (w *window) event(app Window) {
	<-w.ready

	tk := time.NewTicker(time.Second / 960)
	for range tk.C {
		select {
		case key := <-w.keyboard:
			a, ok := app.(KeyboardHanlder)
			if !ok {
				continue
			}
			a.OnKey(key)
		case mo := <-w.mouse:
			a, ok := app.(MouseHandler)
			if !ok {
				continue
			}
			a.OnMouse(mo)
		}
	}
}

type frame struct {
	img  *image.RGBA
	done chan event
}

// The Draw Thread: triple-buffered drawing.
func (w *window) draw(app Window) {
	<-w.ready

	frames := [3]frame{
		{done: make(chan event, 1)},
		{done: make(chan event, 1)},
		{done: make(chan event, 1)},
	}
	for i := 0; i < len(frames); i++ {
		frames[i].done <- event{}
	}
	frameIdx := 0

	last := time.Now()
	tPerFrame := time.Second / 240
	tk := time.NewTicker(tPerFrame)
	for {
		select {
		case siz := <-w.resize:
			w.win.ctx.Resize(siz.w, siz.h)
			w.win.config.size.X = siz.w
			w.win.config.size.Y = siz.h
			if a, ok := app.(ResizeHandler); ok {
				a.OnResize(w.win.config.size.X, w.win.config.size.Y)
				continue
			}
		case <-tk.C:
			appdraw, ok := app.(DrawHandler)
			if !ok {
				continue
			}

			img, redraw := appdraw.Draw()
			if !redraw {
				continue
			}

			c := time.Now()
			t := c.Sub(last)
			last = c
			if t < tPerFrame {
				continue
			}
			if w.win.config.fps {
				w.fontDrawer.Dot = math.P(5, 15)
				w.fontDrawer.Dst = img
				w.fontDrawer.DrawString(fmt.Sprintf("%d", time.Second/t))
			}

			f := frames[frameIdx]
			f.img = img
			w.flush(f)
			frameIdx = (frameIdx + 1) % 3
		}
	}
}

func (w *window) flush(f frame) {
	<-f.done

	dx, dy := f.img.Bounds().Dx(), f.img.Bounds().Dy()
	drawable, err := w.win.ctx.layer.NextDrawable()
	if err != nil {
		panic(fmt.Errorf("app: couldn't get the next drawable: %w", err))
	}

	tex := w.win.ctx.device.MakeTexture(mtl.TextureDescriptor{
		PixelFormat: mtl.PixelFormatBGRA8UNorm,
		Width:       dx,
		Height:      dy,
		StorageMode: mtl.StorageModeManaged,
	})

	region := mtl.RegionMake2D(0, 0, dx, dy)
	tex.ReplaceRegion(region, 0, f.img.Pix, uintptr(4*dx))
	cb := w.win.ctx.queue.MakeCommandBuffer()
	bce := cb.MakeBlitCommandEncoder()
	drawTex := drawable.Texture()
	bce.CopyFromTexture(tex, 0, 0, mtl.Origin{},
		mtl.Size{Width: dx, Height: dy, Depth: 1},
		drawTex, 0, 0, mtl.Origin{})
	bce.EndEncoding()
	cb.PresentDrawable(drawable)

	cb.AddCompletedHandler(func() {
		f.done <- event{}
		bce.Release()
		cb.Release()
		tex.Release()
		drawTex.Release()
	})

	cb.Commit()
}

type funcData struct {
	fn   func()
	done chan event
}

var (
	funcQ    = make(chan funcData, runtime.GOMAXPROCS(0))
	donePool = sync.Pool{New: func() any { return make(chan event) }}
)

func isMainThread() bool {
	return objc.Send[bool](objc.ID(objc.GetClass("NSThread")), objc.RegisterName("isMainThread"))
}

func runOnMainAsync(f func()) {
	if isMainThread() {
		f()
		return
	}
	funcQ <- funcData{fn: f}
	wakeupMainThread()
}

func runOnMainSync(f func()) {
	if isMainThread() {
		f()
		return
	}
	done := donePool.Get().(chan event)
	defer donePool.Put(done)

	funcQ <- funcData{fn: f, done: done}
	wakeupMainThread()
	<-done
}

func polyredDispatchMainFuncs() {
	for {
		select {
		case f := <-funcQ:
			f.fn()
			if f.done != nil {
				f.done <- event{}
			}
		default:
			return
		}
	}
}
