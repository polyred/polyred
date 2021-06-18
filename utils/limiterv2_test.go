// Copyright 2021 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package utils_test

import (
	"fmt"
	"sync/atomic"
	"testing"

	"changkun.de/x/ddd/utils"
)

func TestLimiterV2(t *testing.T) {
	l := utils.NewWorkerPool(4)
	l.Add(10)
	sum := uint32(0)
	for i := 0; i < 10; i++ {
		fmt.Println(i)
		ii := uint32(i)
		l.Execute(func() {
			atomic.AddUint32(&sum, ii)
		})
	}
	l.Wait()
	if sum != 45 {
		t.Fatalf("wrong sum, expect: %d, want %d", 45, sum)
	}

	l.Add(10)
	sum = uint32(0)
	for i := 0; i < 10; i++ {
		ii := uint32(i)
		l.Execute(func() {
			atomic.AddUint32(&sum, ii)
		})
	}
	l.Wait()
	if sum != 45 {
		t.Fatalf("wrong sum, expect: %d, want %d", 45, sum)
	}
}
