// Copyright 2022 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package imageutil

import (
	"fmt"
	"image"
	"image/png"
	"os"
)

// Save stores the current frame buffer to a newly created file.
func Save(buf *image.RGBA, dst string) error {
	err := flushBuf(buf, dst)
	if err != nil {
		return fmt.Errorf("cannot save the given buffer to a file, err: %w", err)
	}
	return nil
}

func flushBuf(buf *image.RGBA, dst string) error {
	f, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer f.Close()

	err = png.Encode(f, buf)
	if err != nil {
		return err
	}

	return nil
}
