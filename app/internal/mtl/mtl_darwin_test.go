package mtl_test

import (
	"testing"

	"poly.red/app/internal/mtl"
)

func TestDevice(t *testing.T) {
	_, err := mtl.CreateSystemDefaultDevice()
	if err != nil {
		t.Fatal(err)
	}
}
