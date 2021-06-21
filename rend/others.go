// Copyright 2021 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package rend

import (
	"runtime"
	"sync/atomic"

	"changkun.de/x/ddd/geometry/primitive"
	"changkun.de/x/ddd/math"
)

func (r *Renderer) resetGBuf() {
	for i := range r.gBuf {
		r.gBuf[i] = gInfo{z: -1}
	}
}

func (r *Renderer) resetFrameBuf() {
	for i := range r.frameBuf.Pix {
		r.frameBuf.Pix[i] = 0
	}
}

// wait waits the current rendering terminates
func (r *Renderer) wait() {
	atomic.StoreUint32(&r.stop, 1)
	for r.isRunning() {
		runtime.Gosched()
	}
	atomic.StoreUint32(&r.stop, 0)
}

func (r *Renderer) startRunning() {
	atomic.StoreUint32(&r.running, 1)
}

func (r *Renderer) isRunning() bool {
	return atomic.LoadUint32(&r.running) == 1
}

func (r *Renderer) stopRunning() {
	atomic.StoreUint32(&r.running, 0)
}

func (r *Renderer) shouldStop() bool {
	return atomic.LoadUint32(&r.stop) == 1
}

func defaultVertexShader(v primitive.Vertex, uniforms map[string]interface{}) primitive.Vertex {
	matModel := uniforms["matModel"].(math.Matrix)
	matView := uniforms["matView"].(math.Matrix)
	matProj := uniforms["matProj"].(math.Matrix)
	matVP := uniforms["matVP"].(math.Matrix)
	matNormal := uniforms["matNormal"].(math.Matrix)

	pos := v.Pos.Apply(matModel).Apply(matView).Apply(matProj).Apply(matVP)
	return primitive.Vertex{
		Pos: pos.Scale(1/pos.W, 1/pos.W, 1/pos.W, 1/pos.W),
		Col: v.Col,
		UV:  v.UV,
		Nor: v.Nor.Apply(matNormal),
	}
}
