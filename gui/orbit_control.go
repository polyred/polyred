// Copyright 2021 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package gui

import (
	"poly.red/camera"
	"poly.red/math"
)

// FIXME: work with orthographic camera

// OrbitEnabled specifies which control types are enabled.
type OrbitEnabled int

// The possible control types.
const (
	OrbitNone OrbitEnabled = 0x00
	OrbitRot  OrbitEnabled = 0x01
	OrbitZoom OrbitEnabled = 0x02
	OrbitPan  OrbitEnabled = 0x04
	OrbitKeys OrbitEnabled = 0x08
	OrbitAll  OrbitEnabled = 0xFF
)

// orbitState bitmask
type orbitState int

const (
	stateNone = orbitState(iota)
	stateRotate
	stateZoom
	statePan
)

// OrbitControl is a camera controller that allows orbiting a target
// point while looking at it. It allows the user to rotate, zoom, and
// pan a 3D scene using the mouse.
type OrbitControl struct {
	cam     camera.Interface // Controlled camera
	target  math.Vec3        // Camera target, around which the camera orbits
	up      math.Vec3        // The orbit axis (Y+)
	enabled OrbitEnabled     // Which controls are enabled
	state   orbitState       // Current control state

	MinDistance     float32 // Minimum distance from target (default is 1)
	MaxDistance     float32 // Maximum distance from target (default is infinity)
	MinPolarAngle   float32 // Minimum polar angle in radians (default is 0)
	MaxPolarAngle   float32 // Maximum polar angle in radians (default is Pi)
	MinAzimuthAngle float32 // Minimum azimuthal angle in radians (default is negative infinity)
	MaxAzimuthAngle float32 // Maximum azimuthal angle in radians (default is infinity)
	RotSpeed        float32 // Rotation speed factor (default is 1)
	ZoomSpeed       float32 // Zoom speed factor (default is 0.1)

	// Internal
	rotStart  math.Vec2
	panStart  math.Vec2
	zoomStart float32

	win *Window
}

// NewOrbitControl creates and returns a pointer to a new orbit control
// for the specified camera.
func NewOrbitControl(win *Window, cam camera.Interface) *OrbitControl {

	oc := &OrbitControl{
		cam:     cam,
		target:  math.NewVec3(0, 0, 0),
		up:      math.NewVec3(0, 1, 0),
		enabled: OrbitAll,

		MinDistance:     1.0,
		MaxDistance:     math.Inf(1),
		MinPolarAngle:   0,
		MaxPolarAngle:   math.Pi, // 180 degrees as radians
		MinAzimuthAngle: math.Inf(-1),
		MaxAzimuthAngle: math.Inf(1),
		RotSpeed:        1.0,
		ZoomSpeed:       0.1,

		win: win,
	}
	return oc
}

// Reset resets the orbit control.
func (oc *OrbitControl) Reset() {
	oc.target = math.NewVec3(0, 0, 0)
}

// Target returns the current orbit target.
func (oc *OrbitControl) Target() math.Vec3 {
	return oc.target
}

// Set camera orbit target Vec4
func (oc *OrbitControl) SetTarget(v math.Vec3) {
	oc.target = v
}

// Enabled returns the current OrbitEnabled bitmask.
func (oc *OrbitControl) Enabled() OrbitEnabled {

	return oc.enabled
}

// SetEnabled sets the current OrbitEnabled bitmask.
func (oc *OrbitControl) SetEnabled(bitmask OrbitEnabled) {
	oc.enabled = bitmask
}

// Rotate rotates the camera around the target by the specified angles.
func (oc *OrbitControl) Rotate(thetaDelta, phiDelta float32) {
	// Compute direction vector from target to camera
	tcam := oc.cam.Position()
	tcam = tcam.Sub(oc.target)

	// Calculate angles based on current camera position plus deltas
	radius := tcam.Len()
	theta := math.Atan2(tcam.X, tcam.Z) + thetaDelta
	phi := math.Acos(tcam.Y/radius) + phiDelta

	// Restrict phi and theta to be between desired limits
	phi = math.Clamp(phi, oc.MinPolarAngle, oc.MaxPolarAngle)
	phi = math.Clamp(phi, math.Epsilon, math.Pi-math.Epsilon)
	theta = math.Clamp(theta, oc.MinAzimuthAngle, oc.MaxAzimuthAngle)

	// Calculate new cartesian coordinates
	tcam.X = radius * math.Sin(phi) * math.Sin(theta)
	tcam.Y = radius * math.Cos(phi)
	tcam.Z = radius * math.Sin(phi) * math.Cos(theta)

	// Update camera position and orientation
	oc.cam.SetPosition(oc.target.Add(tcam))
	oc.cam.SetLookAt(oc.target, oc.up)
}

// Zoom moves the camera closer or farther from the target the specified
// amount and also updates the camera's orthographic size to match.
func (oc *OrbitControl) Zoom(delta float32) {

	// Compute direction vector from target to camera
	tcam := oc.cam.Position()
	tcam = tcam.Sub(oc.target)

	// Calculate new distance from target and apply limits
	dist := tcam.Len() * (1 + delta/10)
	dist = math.Max(oc.MinDistance, math.Min(oc.MaxDistance, dist))
	oldLength := tcam.Len()
	if oldLength != 0 && dist != oldLength {
		tcam = tcam.Scale(dist/oldLength, dist/oldLength, dist/oldLength)
	}
	oc.cam.SetPosition(oc.target.Add(tcam))
}

// Pan pans the camera and target the specified amount on the plane
// perpendicular to the viewing direction.
func (oc *OrbitControl) Pan(deltaX, deltaY float32) {
	// Compute direction vector from camera to target
	pos := oc.cam.Position()
	target, up := oc.cam.LookAt()
	vdir := oc.target.Sub(pos)

	// Conversion constant between an on-screen cursor delta and its
	// projection on the target plane
	c := 2 * vdir.Len() * math.Tan(math.DegToRad(oc.cam.Fov()/2.0)) /
		math.Max(float32(oc.win.width), float32(oc.win.height))

	// Calculate pan components, scale by the converted offsets and
	// combine them
	var pan, panX, panY math.Vec3
	panX = oc.up.Cross(vdir).Unit()
	panY = vdir.Cross(panX).Unit()
	panX = panX.Scale(c*deltaX, c*deltaX, c*deltaX)
	panY = panY.Scale(c*deltaY, c*deltaY, c*deltaY)
	pan = panX.Add(panY)

	// Add pan offset to camera and target
	oc.cam.SetPosition(pos.Add(pan))
	oc.cam.SetLookAt(target.Add(pan), up)
	oc.target = oc.target.Add(pan)
}

// OnMouse is called when an OnMouseDown/OnMouseUp event is received.
func (oc *OrbitControl) OnMouse(name EventName, ev Event) {
	// If nothing enabled ignore event
	if oc.enabled == OrbitNone {
		return
	}

	switch name {
	case OnMouseDown:
		mev := ev.(*MouseEvent)
		switch mev.Button {
		case MouseButtonLeft: // Rotate
			if oc.enabled&OrbitRot != 0 {
				oc.state = stateRotate
				oc.rotStart.X = mev.Xpos
				oc.rotStart.Y = mev.Ypos
			}
		case MouseButtonMiddle: // Zoom
			if oc.enabled&OrbitZoom != 0 {
				oc.state = stateZoom
				oc.zoomStart = mev.Ypos
			}
		case MouseButtonRight: // Pan
			if oc.enabled&OrbitPan != 0 {
				oc.state = statePan
				oc.panStart.X = mev.Xpos
				oc.panStart.Y = mev.Ypos
			}
		}
	case OnMouseUp:
		oc.state = stateNone
	}
}

// OnCursor is called when an OnCursor event is received.
func (oc *OrbitControl) OnCursor(evname EventName, ev Event) {
	// If nothing enabled ignore event
	if oc.enabled == OrbitNone || oc.state == stateNone {
		return
	}

	mev := ev.(*CursorEvent)
	switch oc.state {
	case stateRotate:
		c := -2 * math.Pi * oc.RotSpeed / math.Max(float32(oc.win.width), float32(oc.win.height))
		oc.Rotate(c*(mev.Xpos-oc.rotStart.X), c*(mev.Ypos-oc.rotStart.Y))
		oc.rotStart.X = mev.Xpos
		oc.rotStart.Y = mev.Ypos
	case stateZoom:
		oc.Zoom(oc.ZoomSpeed * (mev.Ypos - oc.zoomStart))
		oc.zoomStart = mev.Ypos
	case statePan:
		oc.Pan(mev.Xpos-oc.panStart.X, mev.Ypos-oc.panStart.Y)
		oc.panStart.X = mev.Xpos
		oc.panStart.Y = mev.Ypos
	}
}

// onScroll is called when an OnScroll event is received.
func (oc *OrbitControl) OnScroll(evname EventName, ev Event) {
	if oc.enabled&OrbitZoom != 0 {
		sev := ev.(*ScrollEvent)
		oc.Zoom(-sev.Yoffset)
	}
}
