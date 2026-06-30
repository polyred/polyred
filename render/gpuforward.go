// Copyright 2026 The Polyred Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package render

import (
	"errors"
	stdmath "math"

	"poly.red/buffer"
	"poly.red/geometry"
	"poly.red/geometry/primitive"
	"poly.red/gpu"
	"poly.red/material"
	"poly.red/math"
	"poly.red/scene"
)

// The GPU forward rasterizer. The vertex shader transforms model positions to clip
// space (gl_Position = -(trans*pos); the negation matches the renderer's projection
// whose w is negated, and lets glViewport reproduce ViewportMatrix). World position,
// world normal, vertex color and uv are computed/passed CPU-side (exactly as draw())
// and interpolated. The fragment writes a three-target G-buffer with depth testing
// and back-face culling (gl_FrontFacing) to match the CPU forward pass:
//
//	target 0 (RGBA32F): world position xyz, depth (remapped to the CPU's [-1,1])
//	target 1 (RGBA32F): unit world normal xyz, material id
//	target 2 (RGBA32F): u, v, du, dv (texture coords + squared screen-space uv
//	                    gradients via dFdx/dFdy, for the mipmap LOD the CPU derives)
//
// vertex color is not stored: the deferred pass takes basecol from the material
// (texture/diffuse), using the fragment color only for materialless passthrough,
// which the textured scenes this drives do not use.
const fwdGBufVert = `#version 310 es
layout(std430, binding = 0) readonly buffer _pos { float pos[]; };
layout(std430, binding = 1) readonly buffer _wp  { float wpos[]; };
layout(std430, binding = 2) readonly buffer _wn  { float wnor[]; };
layout(std430, binding = 3) readonly buffer _mid { float mid[]; };
layout(std430, binding = 4) readonly buffer _uv  { float uv[]; };
layout(std430, binding = 5) readonly buffer _m   { float m[]; };
out vec3 vWorld;
out vec3 vNormal;
out vec2 vUV;
flat out float vMat;
void main() {
	int i = gl_VertexID;
	vec4 p = vec4(pos[i*4], pos[i*4+1], pos[i*4+2], pos[i*4+3]);
	mat4 T = mat4(m[0],m[1],m[2],m[3], m[4],m[5],m[6],m[7],
	              m[8],m[9],m[10],m[11], m[12],m[13],m[14],m[15]);
	gl_Position = -(T * p);
	vWorld  = vec3(wpos[i*4], wpos[i*4+1], wpos[i*4+2]);
	vNormal = vec3(wnor[i*4], wnor[i*4+1], wnor[i*4+2]);
	vUV     = vec2(uv[i*2], uv[i*2+1]);
	vMat    = mid[i];
}`

const fwdGBufFrag = `#version 310 es
precision highp float;
in vec3 vWorld;
in vec3 vNormal;
in vec2 vUV;
flat in float vMat;
layout(location = 0) out vec4 outWP; // xyz world position, w depth (CPU [-1,1])
layout(location = 1) out vec4 outN;  // xyz unit world normal, w material id
layout(location = 2) out vec4 outUV; // u, v, du, dv
void main() {
	if (!gl_FrontFacing) discard;
	outWP = vec4(vWorld, gl_FragCoord.z * 2.0 - 1.0);
	outN  = vec4(normalize(vNormal), vMat);
	vec2 dx = dFdx(vUV);
	vec2 dy = dFdy(vUV);
	outUV = vec4(vUV, dot(dx, dx), dot(dy, dy));
}`

const noFragment = -2.0

var errGPUForwardUnavailable = errors.New("render: no GPU device for the forward pass")

// gpuForwardPass rasterizes the scene's forward G-buffer on the GPU and fills the
// renderer's FragmentBuffer, the same buffer the deferred pass consumes. It also
// builds r.matTable (as the CPU pass does) since the deferred pass needs it.
//
// It returns an error -- and runPass falls back to the CPU forward pass -- when no
// device is present or the device cannot run the GLSL G-buffer pipeline (the Metal
// backend errors on missing MSL), so darwin and the golden tests are unchanged.
func (r *Renderer) gpuForwardPass() error {
	dev := r.cfg.GPUDevice
	if dev == nil {
		return errGPUForwardUnavailable
	}
	buf := r.CurrBuffer()
	w, h := buf.Bounds().Dx(), buf.Bounds().Dy()
	objs := r.buildForwardObjects()

	vmod, err := dev.NewShaderModule(gpu.ShaderSource{GLSL: fwdGBufVert})
	if err != nil {
		return err
	}
	fmod, err := dev.NewShaderModule(gpu.ShaderSource{GLSL: fwdGBufFrag})
	if err != nil {
		return err
	}
	pipe, err := dev.NewRenderPipeline(gpu.RenderPipelineDescriptor{
		VertexModule: vmod, VertexEntry: "main",
		FragmentModule: fmod, FragmentEntry: "main",
		ColorFormat:       gpu.RGBA32Float,
		ExtraColorFormats: []gpu.TextureFormat{gpu.RGBA32Float, gpu.RGBA32Float},
		DepthFormat:       gpu.Depth32Float,
	})
	if err != nil {
		return err
	}
	mkF32 := func() (*gpu.Texture, error) {
		return dev.NewTexture(gpu.TextureDescriptor{Format: gpu.RGBA32Float, Width: w, Height: h, RenderTarget: true})
	}
	wt, err := mkF32()
	if err != nil {
		return err
	}
	nt, err := mkF32()
	if err != nil {
		return err
	}
	ut, err := mkF32()
	if err != nil {
		return err
	}
	depth, err := dev.NewTexture(gpu.TextureDescriptor{Format: gpu.Depth32Float, Width: w, Height: h, RenderTarget: true})
	if err != nil {
		return err
	}

	enc := dev.NewCommandEncoder()
	rp := enc.BeginRenderPass(gpu.RenderPassDescriptor{
		ColorTexture: wt, Load: gpu.LoadClear, ClearColor: [4]float64{0, 0, 0, 0},
		ExtraColorTargets: []gpu.ColorTarget{
			{Texture: nt, ClearColor: [4]float64{0, 0, 0, noFragment}},
			{Texture: ut, ClearColor: [4]float64{0, 0, 0, 0}},
		},
		DepthTexture: depth, ClearDepth: 1,
	})
	rp.SetPipeline(pipe)
	for _, o := range objs {
		b0, err := newF32Buffer(dev, o.pos)
		if err != nil {
			return err
		}
		b1, _ := newF32Buffer(dev, o.wpos)
		b2, _ := newF32Buffer(dev, o.wnor)
		b3, _ := newF32Buffer(dev, o.mid)
		b4, _ := newF32Buffer(dev, o.uv)
		b5, _ := newF32Buffer(dev, o.trans[:])
		rp.SetVertexBuffer(0, b0)
		rp.SetVertexBuffer(1, b1)
		rp.SetVertexBuffer(2, b2)
		rp.SetVertexBuffer(3, b3)
		rp.SetVertexBuffer(4, b4)
		rp.SetVertexBuffer(5, b5)
		rp.Draw(gpu.TriangleList, 0, len(o.pos)/4)
	}
	rp.End()
	dev.Queue().Submit(enc.Finish())
	dev.Queue().WaitIdle()

	wp := floats32(wt.ReadPixels())
	nr := floats32(nt.ReadPixels())
	uv := floats32(ut.ReadPixels())
	// Render-target texture readback follows GL's bottom-left origin: source row r is
	// screen row h-1-r. The FragmentBuffer (like the CPU pass) is top-down, so read
	// the mirrored row when writing each (x, y). (The deferred pass reads a compute
	// SSBO, which is not flipped, hence only the render path needs this.)
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			idx := ((h-1-y)*w + x) * 4
			if nr[idx+3] < noFragment+0.5 { // no fragment written
				continue
			}
			buf.Set(x, y, buffer.Fragment{
				Ok: true,
				Fragment: primitive.Fragment{
					X:          x,
					Y:          y,
					Depth:      wp[idx+3],
					U:          uv[idx],
					V:          uv[idx+1],
					Du:         uv[idx+2],
					Dv:         uv[idx+3],
					Nor:        math.Vec4[float32]{X: nr[idx], Y: nr[idx+1], Z: nr[idx+2], W: 0},
					WordPos:    math.Vec4[float32]{X: wp[idx], Y: wp[idx+1], Z: wp[idx+2], W: 1},
					MaterialID: int64(stdmath.Round(float64(nr[idx+3]))),
				},
			})
		}
	}
	return nil
}

// forwardObject is one scene object's GPU forward-raster input.
type forwardObject struct {
	pos, wpos, wnor, mid, uv []float32 // model pos; world pos; world normal; flat matid; uv
	trans                    [16]float32
}

// buildForwardObjects tabulates materials into r.matTable (so the deferred pass can
// read them) and produces the per-object vertex streams, mirroring cpuForwardPass.
func (r *Renderer) buildForwardObjects() []forwardObject {
	cam := r.cfg.Camera
	view, proj := cam.ViewMatrix(), cam.ProjMatrix()
	r.matTable = r.matTable[:0]
	var objs []forwardObject
	scene.IterObjects(r.cfg.Scene, func(g *geometry.Geometry, model math.Mat4[float32]) bool {
		world := model.MulM(g.ModelMatrix())
		normalMat := world.Inv().T()
		trans := proj.MulM(view).MulM(world)

		base := int64(len(r.matTable))
		for _, m := range g.Materials() {
			bp, _ := m.(*material.BlinnPhong)
			r.matTable = append(r.matTable, bp)
		}

		o := forwardObject{trans: colMajorMat4(trans)}
		for _, tri := range g.Triangles() {
			if !tri.IsValid() {
				continue
			}
			flatMatID := tri.MaterialID
			if flatMatID >= 0 {
				flatMatID += base
			}
			for _, v := range []*primitive.Vertex{tri.V1, tri.V2, tri.V3} {
				wp := world.MulV(v.Pos)
				wn := v.Nor.Apply(normalMat)
				o.pos = append(o.pos, v.Pos.X, v.Pos.Y, v.Pos.Z, v.Pos.W)
				o.wpos = append(o.wpos, wp.X, wp.Y, wp.Z, 1)
				o.wnor = append(o.wnor, wn.X, wn.Y, wn.Z, 0)
				o.uv = append(o.uv, v.UV.X, v.UV.Y)
				o.mid = append(o.mid, float32(flatMatID))
			}
		}
		objs = append(objs, o)
		return true
	})
	return objs
}

func colMajorMat4(m math.Mat4[float32]) [16]float32 {
	var a [16]float32
	for col := 0; col < 4; col++ {
		for row := 0; row < 4; row++ {
			a[col*4+row] = m.Get(row, col)
		}
	}
	return a
}

func newF32Buffer(dev *gpu.Device, d []float32) (*gpu.Buffer, error) {
	b := make([]byte, len(d)*4)
	for i, f := range d {
		u := stdmath.Float32bits(f)
		b[i*4], b[i*4+1], b[i*4+2], b[i*4+3] = byte(u), byte(u>>8), byte(u>>16), byte(u>>24)
	}
	return dev.NewBuffer(gpu.BufferDescriptor{Data: b, Usage: gpu.BufferStorage})
}

func floats32(b []byte) []float32 {
	out := make([]float32, len(b)/4)
	for i := range out {
		out[i] = stdmath.Float32frombits(uint32(b[i*4]) | uint32(b[i*4+1])<<8 | uint32(b[i*4+2])<<16 | uint32(b[i*4+3])<<24)
	}
	return out
}
