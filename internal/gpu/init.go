// Copyright 2022 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package gpu

// Function is a compute function that runs on GPU.
type Function struct {
	shaderFn
}

// Device is an abstraction for GPU device.
type Device interface {
	Available() bool
}

// Driver returns an avaliable GPU device.
func Driver() Device { return device }
