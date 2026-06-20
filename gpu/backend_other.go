// Copyright 2026 The Polyred Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build !darwin

package gpu

// openBackend has no GPU driver wired on non-darwin platforms yet (GL/Vulkan
// land in later phases).
func openBackend(d Driver) (backend, Driver, error) {
	return nil, DriverAuto, ErrUnsupported
}
