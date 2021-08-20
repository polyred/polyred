package bbb

import (
	"poly.red/texture/buffer"
)

func index(buf []buffer.Fragment, idx int) *buffer.Fragment {
	return &buf[idx]
}
