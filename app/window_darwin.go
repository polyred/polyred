// Copyright 2022 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package app

/*
#cgo CFLAGS: -Werror -fmodules -fobjc-arc -x objective-c

#include <AppKit/AppKit.h>

#define MOUSE_MOVE 0
#define MOUSE_UP 1
#define MOUSE_DOWN 2
#define MOUSE_SCROLL 3

__attribute__ ((visibility ("hidden"))) void polyred_main(void);
__attribute__ ((visibility ("hidden"))) CFTypeRef polyred_createView(void);
__attribute__ ((visibility ("hidden"))) CFTypeRef polyred_createWindow(
	CFTypeRef viewRef,
	const char *title,
	CGFloat width,
	CGFloat height,
	CGFloat minWidth,
	CGFloat minHeight,
	CGFloat maxWidth,
	CGFloat maxHeight,
	bool canResize
);

__attribute__ ((visibility ("hidden"))) void polyred_wakeupMainThread(void);

static bool isMainThread() {
	return [NSThread isMainThread];
}

static CGFloat getScreenBackingScale(void) {
	return [NSScreen.mainScreen backingScaleFactor];
}

static CGFloat getViewBackingScale(CFTypeRef viewRef) {
	NSView *view = (__bridge NSView *)viewRef;
	return [view.window backingScaleFactor];
}

static NSPoint cascadeTopLeftFromPoint(CFTypeRef windowRef, NSPoint topLeft) {
	NSWindow *window = (__bridge NSWindow *)windowRef;
	return [window cascadeTopLeftFromPoint:topLeft];
}

static void makeKeyAndOrderFront(CFTypeRef windowRef) {
	NSWindow *window = (__bridge NSWindow *)windowRef;
	[window makeKeyAndOrderFront:nil];
}

static void setSize(CFTypeRef windowRef, CGFloat width, CGFloat height) {
	NSWindow* window = (__bridge NSWindow *)windowRef;
	NSSize size = NSMakeSize(width, height);
	[window setContentSize:size];
}

static void setMinSize(CFTypeRef windowRef, CGFloat width, CGFloat height) {
	NSWindow* window = (__bridge NSWindow *)windowRef;
	window.contentMinSize = NSMakeSize(width, height);
}

static void setMaxSize(CFTypeRef windowRef, CGFloat width, CGFloat height) {
	NSWindow* window = (__bridge NSWindow *)windowRef;
	window.contentMaxSize = NSMakeSize(width, height);
}

static void setTitle(CFTypeRef windowRef, const char *title) {
	NSWindow* window = (__bridge NSWindow *)windowRef;
	window.title = [NSString stringWithUTF8String: title];
}

static CFTypeRef layerForView(CFTypeRef viewRef) {
	NSView *view = (__bridge NSView *)viewRef;
	return (__bridge CFTypeRef)view.layer;
}

static CGFloat viewHeight(CFTypeRef viewRef) {
	NSView *view = (__bridge NSView *)viewRef;
	return [view bounds].size.height;
}

static CGFloat viewWidth(CFTypeRef viewRef) {
	NSView *view = (__bridge NSView *)viewRef;
	return [view bounds].size.width;
}
*/
import "C"
import (
	"errors"
	"fmt"
	"image"
	"os"
	"runtime"
	"sync"
	"time"
	"unsafe"

	"poly.red/app/internal/mtl"
	"poly.red/math"
)

func init() {
	runtime.LockOSThread()
}

//export polyred_onHide
func polyred_onHide(view C.CFTypeRef) {}

//export polyred_onShow
func polyred_onShow(view C.CFTypeRef) {}

//export polyred_onChangeScreen
func polyred_onChangeScreen(view C.CFTypeRef, did uint64) {}

//export polyred_onFocus
func polyred_onFocus(view C.CFTypeRef, focus C.int) {}

//export polyred_onMouse
func polyred_onMouse(view C.CFTypeRef, act C.int, btn C.NSUInteger, x, y, dx, dy C.CGFloat, ti C.double, mods C.NSUInteger) {
	w := windows[view]
	w.mouse <- MouseEvent{
		Action:  MouseAction(act),
		Button:  MouseButton(btn),
		Mods:    ModifierKey(mods),
		Xpos:    float32(x),
		Ypos:    float32(y),
		Xoffset: float32(dx / 10),
		Yoffset: float32(dy / 10),
	}
}

//export polyred_onKeys
func polyred_onKeys(view C.CFTypeRef, code uint32, cstr *C.char, ti C.double, mods C.NSUInteger, keyDown C.bool) {
	w := windows[view]
	w.keyboard <- KeyEvent{
		Keycode: Key{code: code, char: C.GoString(cstr)},
		Mods:    ModifierKey(mods),
		Pressed: bool(keyDown),
	}
}

//export polyred_onText
func polyred_onText(view C.CFTypeRef, cstr *C.char) {}

//export polyred_onDraw
func polyred_onDraw(view C.CFTypeRef) {
	w := windows[view]
	w.resize <- resizeEvent{
		int(float32(C.viewWidth(view))),
		int(float32(C.viewHeight(view))),
	}
}

//export polyred_onClose
func polyred_onClose(view C.CFTypeRef) {
	w := windows[view]
	delete(windows, view)
	C.CFRelease(w.win.view)
	C.CFRelease(w.win.window)
	w.win.view = 0
	w.win.window = 0
	w.win.ctx.Release()
	os.Exit(0)
}

//export polyred_onAppHide
func polyred_onAppHide() {}

//export polyred_onAppShow
func polyred_onAppShow() {}

var launched = make(chan struct{})

//export polyred_onFinishLaunching
func polyred_onFinishLaunching() {
	close(launched)
}

// nextTopLeft is the offset to use for the next window's call to
// cascadeTopLeftFromPoint.
var nextTopLeft C.NSPoint

var (
	mu      sync.Mutex
	windows = map[C.CFTypeRef]*window{} // view -> window
)

type osWindow struct {
	view   C.CFTypeRef // PolyredView*
	window C.CFTypeRef // NSWindow*
	ctx    *mtlContext

	viewScale   float32
	screenScale int
	config      *config
}

func (w *window) run(app Window, cfg config, opts ...Opt) {
	<-launched

	w.win = &osWindow{config: &cfg}
	runOnMainSync(func() {

		w.win.view = C.polyred_createView()
		if w.win.view == 0 {
			panic(errors.New("CreateWindow: failed to create view"))
		}

		mu.Lock()
		windows[w.win.view] = w
		mu.Unlock()

		w.win.viewScale = float32(C.getViewBackingScale(w.win.view))
		w.win.screenScale = int(C.getScreenBackingScale())

		canResize := false
		if _, ok := app.(ResizeHandler); ok {
			canResize = true
		}
		w.win.window = C.polyred_createWindow(w.win.view, nil, 0, 0, 0, 0, 0, 0, C.bool(canResize))
		w.configs(opts...)

		if nextTopLeft.x == 0 && nextTopLeft.y == 0 {
			// cascadeTopLeftFromPoint treats (0, 0) as a no-op,
			// and just returns the offset we need for the first window.
			nextTopLeft = C.cascadeTopLeftFromPoint(w.win.window, nextTopLeft)
		}

		nextTopLeft = C.cascadeTopLeftFromPoint(w.win.window, nextTopLeft)
		C.makeKeyAndOrderFront(w.win.window)

		// initialize Metal driver
		var err error
		w.win.ctx, err = newMtlContext(w.win.config, newMetalLayer(C.layerForView(w.win.view)))
		if err != nil {
			panic(fmt.Errorf("app: failed to use Metal: %w", err))
		}
		close(w.ready)
	})
}

func (w *window) configs(opts ...Opt) {
	cfg := w.win.config
	for _, o := range opts {
		o(cfg)
	}
	C.setSize(w.win.window, C.CGFloat(cfg.size.X), C.CGFloat(cfg.size.Y))
	C.setMinSize(w.win.window, C.CGFloat(cfg.minSize.X), C.CGFloat(cfg.minSize.Y))
	C.setMaxSize(w.win.window, C.CGFloat(cfg.maxSize.X), C.CGFloat(cfg.maxSize.Y))

	w.win.config.title = cfg.title
	title := C.CString(cfg.title)
	defer C.free(unsafe.Pointer(title))
	C.setTitle(w.win.window, title)
}

func (w *window) main(app Window) {
	go w.event(app)
	go w.draw(app)
	C.polyred_main()
}

// The Event Thread
//
// The ticker ticks every ~1ms which permits a maximum of 960 fps
// (should large enough) for input events handling as the key to
// making sure the window being responsive (especially on macOS).
// Since we manage time event timeout ourselves using the ticker,
// the glfw.PollEvents is used.
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

// The Draw Thread
//
// We use tripple-buffering techniques for the drawing.
// The benefit is that this enables motion vectors between.
//
// TODO: Think about how to provide API to let people to access
// previous frame, such as frame.Prev()? But I don't have enough
// practice here yet. Maybe try something like TAA with the tripple
// buffering.
func (w *window) draw(app Window) {
	<-w.ready

	// Managing 3 drawable frames:
	// frames contain their own done indicator, make sure
	// that each frame is indeed drawed from the GPU level.
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
	tPerFrame := time.Second / 240 // 120 fps
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

	// We create a new texture for every draw call. A temporary texture
	// is needed since ReplaceRegion tries to sync the pixel data between
	// CPU and GPU, and doing it on the existing texture is inefficient.
	// The texture cannot be reused until sending the pixels finishes,
	// then create new ones for each call.
	tex := w.win.ctx.device.MakeTexture(mtl.TextureDescriptor{
		PixelFormat: mtl.PixelFormatBGRA8UNorm,
		Width:       dx,
		Height:      dy,
		StorageMode: mtl.StorageModeManaged,
	})

	region := mtl.RegionMake2D(0, 0, dx, dy)
	tex.ReplaceRegion(region, 0, &f.img.Pix[0], uintptr(4*dx))
	cb := w.win.ctx.queue.MakeCommandBuffer()
	bce := cb.MakeBlitCommandEncoder()
	drawTex := drawable.Texture()
	bce.CopyFromTexture(tex, 0, 0, mtl.Origin{},
		mtl.Size{Width: dx, Height: dy, Depth: 1},
		drawTex, 0, 0, mtl.Origin{})
	bce.EndEncoding()
	cb.PresentDrawable(drawable)

	// We need a synchornization to understand if the current frame is
	// completed. Similar to glFinish (instead of glFlush). A general
	// discussion about CPU, GPU and display synchornization can be
	// found here:
	// - Working with Metal: Fundamentals, 21:28
	// - https://developer.apple.com/videos/play/wwdc2014/604/
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

func runOnMainAsync(f func()) {
	if C.isMainThread() {
		f()
		return
	}

	funcQ <- funcData{fn: f}
	C.polyred_wakeupMainThread()
}

func runOnMainSync(f func()) {
	if C.isMainThread() {
		f()
		return
	}

	done := donePool.Get().(chan event)
	defer donePool.Put(done)

	funcQ <- funcData{fn: f, done: done}
	C.polyred_wakeupMainThread()
	<-done
}

//export polyred_dispatchMainFuncs
func polyred_dispatchMainFuncs() {
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
