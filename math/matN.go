// Copyright 2022 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package math

import (
	"fmt"
	"math/rand"
	"reflect"
	"strings"
)

// Type defines all supported data types.
type Type interface {
	~uint8 | ~uint32 | ~int32 | ~float32
}

// TypeSize returns the corresponding type size.
func TypeSize[T Type]() int {
	var v T
	switch any(v).(type) {
	case uint8:
		return 1
	case int32:
		return 4
	case uint32:
		return 4
	case float32:
		return 4
	}
	panic("unknown data size for type")
}

// FIXME: decide of column major and row major.

// Mat represents a WxH matrix.
type Mat[T Type] struct {
	Row  int
	Col  int
	Data []T
}

// NewMat creates a new WxH matrix.
func NewMat[T Type](row, col int, xn ...T) Mat[T] {
	switch len(xn) {
	case row * col:
		return Mat[T]{
			Row:  row,
			Col:  col,
			Data: append([]T{}, xn...),
		}
	case 0:
		return Mat[T]{
			Row:  row,
			Col:  col,
			Data: make([]T, row*col),
		}
	}
	panic("math: incorrect matrix dimension")
}

// NewRandMat returns a random matrix.
func NewRandMat[T Type](row, col int) Mat[T] {
	m := Mat[T]{
		Row:  row,
		Col:  col,
		Data: make([]T, row*col),
	}

	for i := range m.Data {
		m.Data[i] = T(rand.Float64())
	}
	return m
}

// String returns a string format of the given Mat4.
func (m Mat[T]) String() string {
	ss := "[\n"
	for i := 0; i < m.Row; i++ {
		s := "["
		for j := 0; j < m.Col; j++ {
			s += fmt.Sprintf("%v, ", m.Get(i, j))
		}
		ss += "	" + strings.TrimSuffix(s, ", ") + "],\n"
	}
	ss += "]"
	return ss
}

// Index returns the element index of Data at (i, j)
func (m Mat[T]) Index(i, j int) int {
	return i*m.Col + j
}

// Get gets the corresponding element at (i, j)
func (m Mat[T]) Get(i, j int) T {
	return m.Data[m.Index(i, j)]
}

// Set sets the given value to matrix at (i, j)
func (m Mat[T]) Set(i, j int, v T) {
	m.Data[m.Index(i, j)] = v
}

// Eq returns true if two matrices are equal.
func (m Mat[T]) Eq(n Mat[T]) bool {
	if m.Row != n.Row || m.Col != n.Col {
		return false
	}

	return reflect.DeepEqual(m.Data, n.Data)
}

// Add adds two given matrix: m+n
func (m Mat[T]) Add(n Mat[T]) Mat[T] {
	if m.Row == n.Row && m.Col == n.Col {
		r := Mat[T]{
			Row:  m.Row,
			Col:  m.Col,
			Data: make([]T, len(m.Data)),
		}
		for i := range m.Data {
			r.Data[i] = m.Data[i] + n.Data[i]
		}
		return r
	}
	panic(fmt.Sprintf("math: mismatched matrix dimension: A(%v, %v) != B(%v, %v)", m.Row, m.Col, n.Row, n.Col))
}

// Add subtracts two given matrix: m-n
func (m Mat[T]) Sub(n Mat[T]) Mat[T] {
	if m.Row == n.Row && m.Col == n.Col {
		r := Mat[T]{
			Row:  m.Row,
			Col:  m.Col,
			Data: make([]T, len(m.Data)),
		}
		for i := range m.Data {
			r.Data[i] = m.Data[i] - n.Data[i]
		}
		return r
	}
	panic("math: mismatched matrix dimension")
}

// Mul applies matrix multiplication of two given matrix, and returns
// the resulting matrix: r = m*n
func (m Mat[T]) Mul(n Mat[T]) Mat[T] {
	if m.Col == n.Row {
		blockSize := 4
		return m.blockMul(blockSize, n)
	}
	panic("math: mismatched matrix dimension")
}

// blockMul is a blocking version of matrix multiplication in jki order.
func (m Mat[T]) blockMul(blockSize int, n Mat[T]) (r Mat[T]) {
	if m.Col != n.Row {
		panic("math: mismatched matrix dimension")
	}

	r = Mat[T]{
		Row:  m.Row,
		Col:  n.Col,
		Data: make([]T, m.Row*n.Col),
	}

	min := m.Row
	if m.Col < min {
		min = m.Col
	}
	if n.Col < min {
		min = n.Col
	}
	var (
		kk, jj, i, j, k int
		rr              T
		en              = blockSize * (min / blockSize)
	)

	for kk = 0; kk < en; kk += blockSize {
		for jj = 0; jj < en; jj += blockSize {
			for k = kk; k < kk+blockSize; k++ {
				for j = jj; j < jj+blockSize; j++ {
					rr = n.Get(k, j)
					for i = 0; i < m.Row; i++ {
						r.Set(i, j, r.Get(i, j)+rr*m.Get(i, k))
					}
				}
			}
		}
		for k = kk; k < kk+blockSize; k++ {
			for j = en; j < n.Col; j++ {
				rr = n.Get(k, j)
				for i = 0; i < m.Row; i++ {
					r.Set(i, j, r.Get(i, j)+rr*m.Get(i, k))
				}
			}
		}
	}

	for jj = 0; jj < en; jj += blockSize {
		for k = en; k < m.Col; k++ {
			for j = jj; j < jj+blockSize; j++ {
				rr = n.Get(k, j)
				for i = 0; i < m.Row; i++ {
					r.Set(i, j, r.Get(i, j)+rr*m.Get(i, k))
				}
			}
		}
	}

	for k = en; k < m.Col; k++ {
		for j = en; j < n.Col; j++ {
			rr = n.Get(k, j)
			for i = 0; i < m.Row; i++ {
				r.Set(i, j, r.Get(i, j)+rr*m.Get(i, k))
			}
		}
	}
	return
}

// T returns the transpose of a given matrix.
func (m Mat[T]) T() Mat[T] {
	r := Mat[T]{
		Row:  m.Col,
		Col:  m.Row,
		Data: make([]T, len(m.Data)),
	}

	for j := 0; j < m.Col; j++ {
		for i := 0; i < m.Row; i++ {
			r.Set(i, j, m.Get(j, i))
		}
	}
	return r
}
