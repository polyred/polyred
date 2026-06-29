// Copyright 2026 The Polyred Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build !linux

package gpu

// openVKBackend has no Vulkan implementation off Linux yet; the GL backend's
// openBackend dispatches here for DriverVulkan on such platforms.
func openVKBackend(c config) (backend, Driver, error) {
	return nil, DriverAuto, ErrUnsupported
}
