// Copyright 2026 The Polyred Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build !darwin && !linux && !windows

package gpu

// openBackend has no GPU driver wired on platforms without Metal (darwin) or the
// GL backend (linux) yet; Windows/Vulkan/DX12 land in later phases.
func openBackend(c config) (backend, Driver, error) {
	return nil, DriverAuto, ErrUnsupported
}
