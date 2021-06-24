package win

import (
	"fmt"
	"image"
	"image/color"
	"time"
	"unsafe"

	"changkun.de/x/polyred/rend"
	"github.com/go-gl/gl/v2.1/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
	"golang.design/x/mainthread"
	"golang.design/x/thread"
	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"
)

// Option is a functional option to the window constructor New.
type Option func(*options)

type options struct {
	title         string
	width, height int
	resizable     bool
	showFPS       bool
}

// Title option sets the title (caption) of the window.
func Title(title string) Option {
	return func(o *options) {
		o.title = title
	}
}

// Size option sets the width and height of the window.
func Size(width, height int) Option {
	return func(o *options) {
		o.width = width
		o.height = height
	}
}

// ShowFPS sets the window to show FPS.
func ShowFPS() Option {
	return func(o *options) {
		o.showFPS = true
	}
}

// Resizable option makes the window resizable by the user.
func Resizable() Option {
	return func(o *options) {
		o.resizable = true
	}
}

var conv = map[bool]int{
	true:  glfw.True,
	false: glfw.False,
}

// Win represents a window.
type Win struct {
	win *glfw.Window
	th  thread.Thread // render thread

	resize chan image.Rectangle

	renderer *rend.Renderer

	img     *image.RGBA
	ratio   int // for retina display
	showFPS bool
}

// NewWindow constructs a new graphical window.
func NewWindow(opts ...Option) (*Win, error) {
	o := options{
		title:     "",
		width:     800,
		height:    600,
		resizable: false,
	}
	for _, opt := range opts {
		opt(&o)
	}

	var (
		w = &Win{
			th:     thread.New(),
			resize: make(chan image.Rectangle),
		}
		err error
	)
	defer func() {
		if err != nil {
			// This function must be called from the mainthread.
			mainthread.Call(w.win.Destroy)
			w.th.Terminate()
		}
	}()

	mainthread.Call(func() {
		glfw.WindowHint(glfw.ContextVersionMajor, 2)
		glfw.WindowHint(glfw.ContextVersionMinor, 1)
		glfw.WindowHint(glfw.DoubleBuffer, glfw.False)
		glfw.WindowHint(glfw.Resizable, conv[o.resizable])

		w.win, err = glfw.CreateWindow(o.width, o.height, o.title, nil, nil)
		if err != nil {
			return
		}

		// Ratio test. for high DPI, e.g. macOS Retina
		width, _ := w.win.GetFramebufferSize()
		w.ratio = width / o.width
		if w.ratio < 1 {
			w.ratio = 1
		}
		if w.ratio != 1 {
			o.width /= w.ratio
			o.height /= w.ratio
		}
		w.win.Destroy()
		w.win, err = glfw.CreateWindow(o.width, o.height, o.title, nil, nil)
	})
	if err != nil {
		return nil, err
	}
	w.img = image.NewRGBA(image.Rect(0, 0, o.width*w.ratio, o.height*w.ratio))
	w.th.Call(func() {
		w.win.MakeContextCurrent()
		err = gl.Init()
	})
	w.showFPS = o.showFPS
	if err != nil {
		return nil, err
	}
	return w, nil
}

// SetRenderer sets fn as the renderer callback
func (w *Win) SetRenderer(r *rend.Renderer) {
	w.th.Call(func() { w.renderer = r }) // for thread safety
}

// Run runs the given window and enters event handling loop
// and blocks until it should be close and destroyed.
func (w *Win) Run() {
	defer func() {
		// This function must be called from the mainthread.
		mainthread.Call(w.win.Destroy)
		w.th.Terminate()
	}()

	for !w.Closed() {
		select {
		// TODO: fix resize
		case r := <-w.resize:
			fmt.Println("resize:", r.Max)
			w.renderer.UpdateOptions(
				rend.WithSize(r.Max.X, r.Max.Y),
			)
		default:
			mainthread.Call(func() { glfw.WaitEventsTimeout(1.0 / 30) })
		}
		w.th.Call(func() { w.flush() })
	}
}

// Stop stops the given window.
func (w *Win) Stop() {
	// This function can be called from any threads.
	w.th.Call(func() { w.win.SetShouldClose(true) })
}

// Closed asserts if the given window is closed.
func (w *Win) Closed() bool {
	// This function can be called from any thread.
	var stop bool
	w.th.Call(func() { stop = w.win.ShouldClose() })
	return stop
}

// flush flushes rendered image to the window.
func (w *Win) flush() {
	if w.renderer == nil {
		return
	}
	var img *image.RGBA
	if w.showFPS {
		t := time.Now()
		img = w.renderer.Render()
		col := color.RGBA{200, 100, 0, 255}
		point := fixed.Point26_6{X: fixed.Int26_6(0 * 64), Y: fixed.Int26_6(13 * 64)}
		d := font.Drawer{
			Dst: img, Src: image.NewUniform(col),
			Face: basicfont.Face7x13, Dot: point,
		}
		d.DrawString(fmt.Sprintf("%d", time.Second/time.Since(t)))
	} else {
		img = w.renderer.Render()
	}

	dx, dy := img.Bounds().Dx(), img.Bounds().Dy()
	gl.DrawBuffer(gl.FRONT)
	gl.Viewport(0, 0, int32(dx), int32(dy))
	gl.RasterPos2d(-1, 1)
	gl.PixelZoom(1, -1)
	gl.DrawPixels(int32(dx), int32(dy),
		gl.RGBA, gl.UNSIGNED_BYTE, unsafe.Pointer(&img.Pix[0]))
	gl.Flush()
}
