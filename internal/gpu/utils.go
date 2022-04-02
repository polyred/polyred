// Copyright 2022 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package gpu

import "errors"

func try[T any](v T, err error) T {
	if err != nil {
		panic(err)
	}
	return v
}

func handle(f func(err error)) {
	if r := recover(); r != nil {
		var err error
		switch x := r.(type) {
		case string:
			err = errors.New(x)
		case error:
			err = x
		default:
			err = errors.New("unknown panic")
		}
		f(err)
	}
	return
}
