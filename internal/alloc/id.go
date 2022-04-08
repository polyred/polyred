// Copyright 2022 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

// Package alloc allocates globally unique objects.
package alloc

import "sync/atomic"

var objectID uint64

func ID() uint64 {
	return atomic.AddUint64(&objectID, 1)
}
