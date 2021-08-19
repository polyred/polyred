// Copyright 2021 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package render

import (
	"poly.red/color"
	"poly.red/shader"
	"poly.red/texture/buffer"
)

// DrawFragments is a concurrent executor of the given shader that travel
// through all fragments. Each fragment executes the given shaders exactly once.
//
// One should not manipulate the given image buffer in the shader.
// Instead, return the resulting color in the shader can avoid data race.
func (r *Renderer) DrawFragments(buf *buffer.Buffer, funcs ...shader.FragmentProgram) {
	if funcs == nil {
		return
	}

	// Because the shader executes exactly on each pixel once, there is
	// no need to lock the pixel while reading or writing.

	w := buf.Bounds().Dx()
	h := buf.Bounds().Dy()
	n := w * h

	batchSize := int(r.batchSize)
	wsteps := w / batchSize
	hsteps := h / batchSize

	defer r.sched.Wait()

	if wsteps == 0 && hsteps == 0 {
		r.sched.Add(1)

		// Note: sadly that the executing function will escape to the
		// heap which increases the memory allocation. No workaround.
		r.sched.Run(func() {
			for i := 0; i < n; i++ {
				r.DrawFragment(buf, i%w, i/w, funcs...)
			}
		})
		return
	}

	numTasks := n / batchSize
	r.sched.Add(uint64(numTasks))
	for i := 0; i < numTasks; i++ {
		ii := i
		r.sched.Run(func() {
			x0 := ii * batchSize
			x1 := x0 + batchSize
			for j := x0; j < x1; j++ {
				x, y := j%w, j/w
				r.DrawFragment(buf, x, y, funcs...)
			}
		})
	}

	if n%batchSize != 0 {
		r.sched.Add(1)
		r.sched.Run(func() {
			for j := numTasks * batchSize; j < n; j++ {
				x, y := j%w, j/w
				r.DrawFragment(buf, x, y, funcs...)
			}
		})
	}
}

// DrawFragment executes the given shaders on a specific fragment.
//
// Note that it is caller's responsibility to protect the safty of fragment
// coordinates, as well as data race of the given buffer.
func (r *Renderer) DrawFragment(buf *buffer.Buffer, x, y int, shaders ...shader.FragmentProgram) {
	info := buf.UnsafeAt(x, y)
	old := info.Col

	for i := 0; i < len(shaders); i++ {
		info.Col = shaders[i](info.Fragment)
		if info.Col == color.Discard {
			return
		}
	}

	if r.blendFunc != nil {
		info.Col = r.blendFunc(old, info.Col)
	}
	buf.UnsafeSet(x, y, info)
}
