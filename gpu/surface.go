// Copyright 2026 The Polyred Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gpu

import (
	"errors"
	"image"
)

// Surface is a presentable swapchain: a ring of render-target textures the
// renderer draws into, one per frame. It is the API windowed present is built
// on. This implementation is backend-agnostic and headless — AcquireNextTexture
// rotates through the textures and Present completes the frame so its pixels can
// be read back (the render-to-image path). Attaching the swapchain to an on-screen
// window (a CAMetalLayer drawable on darwin, an EGL/WGL window surface elsewhere)
// is the one piece that needs a display and is layered on top of this API; see
// specs/foundations/gpu-windowed-present.md.
type Surface struct {
	d        *Device
	w, h     int
	format   TextureFormat
	textures []*Texture
	frame    int
	acquired bool
	bs       backendWindowSurface // nil for headless; set for an on-screen surface
}

// SurfaceDescriptor configures a swapchain.
type SurfaceDescriptor struct {
	Width  int
	Height int
	Format TextureFormat
	// Frames is the swapchain length (number of in-flight textures); defaults to
	// 2 (double buffering) when zero.
	Frames int
}

// CreateSurface creates a headless presentable swapchain. The textures are
// usable as render-pass color attachments.
func (d *Device) CreateSurface(desc SurfaceDescriptor) (*Surface, error) {
	if desc.Width <= 0 || desc.Height <= 0 {
		return nil, errors.New("gpu: surface size must be > 0")
	}
	frames := desc.Frames
	if frames <= 0 {
		frames = 2
	}
	s := &Surface{d: d, w: desc.Width, h: desc.Height, format: desc.Format}
	if err := s.allocate(frames); err != nil {
		return nil, err
	}
	return s, nil
}

// WindowSurfaceDescriptor configures an on-screen swapchain bound to a native window.
type WindowSurfaceDescriptor struct {
	Display uintptr // native display handle (e.g. X11 Display*); 0 if not applicable
	Window  uintptr // native window handle (e.g. X11 Window XID, Win32 HWND)
	Width   int
	Height  int
	Format  TextureFormat
}

// CreateWindowSurface creates an on-screen swapchain bound to a native window.
// The backend presents acquired frames to the window (the GL backend blits and
// swaps an EGL window surface). Not all backends support this; those that do not
// return ErrUnsupported.
func (d *Device) CreateWindowSurface(desc WindowSurfaceDescriptor) (*Surface, error) {
	if desc.Width <= 0 || desc.Height <= 0 {
		return nil, errors.New("gpu: surface size must be > 0")
	}
	bsurf, err := d.b.newWindowSurface(desc.Display, desc.Window, desc.Width, desc.Height)
	if err != nil {
		return nil, err
	}
	return &Surface{d: d, w: desc.Width, h: desc.Height, format: desc.Format, bs: bsurf}, nil
}

func (s *Surface) allocate(frames int) error {
	s.textures = s.textures[:0]
	for i := 0; i < frames; i++ {
		t, err := s.d.NewTexture(TextureDescriptor{
			Format: s.format, Width: s.w, Height: s.h, RenderTarget: true,
		})
		if err != nil {
			return err
		}
		s.textures = append(s.textures, t)
	}
	return nil
}

// AcquireNextTexture returns the swapchain texture to render the next frame into.
// Pair each acquire with a Present.
func (s *Surface) AcquireNextTexture() *Texture {
	if s.bs != nil {
		s.acquired = true
		return &Texture{b: s.bs.acquire(), w: s.w, h: s.h}
	}
	t := s.textures[s.frame%len(s.textures)]
	s.acquired = true
	return t
}

// Present finishes the acquired frame. Headless, this waits for the GPU so the
// frame's pixels are ready for ReadPixels; an on-screen surface would instead
// hand the drawable to the window server.
func (s *Surface) Present() error {
	if !s.acquired {
		return errors.New("gpu: Present without AcquireNextTexture")
	}
	if s.bs != nil {
		s.acquired = false
		s.frame++
		return s.bs.present()
	}
	s.d.Queue().WaitIdle()
	s.acquired = false
	s.frame++
	return nil
}

// PresentImage is a convenience for the windowed path: it acquires the next
// frame's texture, uploads img's pixels into it, and presents. img.Pix is
// assumed tightly packed (stride == width*4), matching the app's frame buffer.
func (s *Surface) PresentImage(img *image.RGBA) error {
	tex := s.AcquireNextTexture()
	tex.Write(img.Pix)
	return s.Present()
}

// Release frees the on-screen surface's backend resources. No-op for headless.
func (s *Surface) Release() {
	if s.bs != nil {
		s.bs.release()
	}
}

// PresentedPixels reads back the pixels the on-screen surface presents to the
// window, as top-down tightly-packed RGBA (width*height*4 bytes). It is meant for
// tests and screenshots; it returns nil for a headless surface.
func (s *Surface) PresentedPixels() []byte {
	if s.bs == nil {
		return nil
	}
	return s.bs.readback()
}

// Texture returns the most recently presented frame's texture (for read-back in
// the headless path). Valid after at least one Present.
func (s *Surface) Texture() *Texture {
	idx := (s.frame - 1) % len(s.textures)
	if idx < 0 {
		idx += len(s.textures)
	}
	return s.textures[idx]
}

// Resize reallocates the swapchain textures for a new size.
func (s *Surface) Resize(w, h int) error {
	if w <= 0 || h <= 0 {
		return errors.New("gpu: surface size must be > 0")
	}
	if s.bs != nil {
		s.w, s.h = w, h
		s.frame, s.acquired = 0, false
		return s.bs.resize(w, h)
	}
	s.w, s.h = w, h
	s.frame, s.acquired = 0, false
	return s.allocate(len(s.textures))
}

// Size reports the current swapchain dimensions.
func (s *Surface) Size() (int, int) { return s.w, s.h }
