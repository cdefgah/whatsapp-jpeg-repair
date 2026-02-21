// SPDX-License-Identifier: GPL-3.0-only
// Copyright (c) 2021 by Rafael Osipov <rafael.osipov@outlook.com>

package testutil

// This package is not referenced from production code, so it should not be included to the executable file.
// It is referenced only as a helper from tests.

import (
	"bytes"
	"image"
	"image/color"
	"image/jpeg"

	"github.com/spf13/afero"
)

// CreateJpegBytesBuffer creates byte buffer with JPEG image for testing purposes.
func CreateJpegBytesBuffer() ([]byte, error) {
	img := image.NewRGBA(image.Rect(0, 0, 10, 10))
	var buf bytes.Buffer
	err := jpeg.Encode(&buf, img, nil)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// CreateJpegFile creates dummy jpeg file for testing purposes.
func CreateJpegFile(fs afero.Fs, path string) error {
	img := image.NewRGBA(image.Rect(0, 0, 1, 1))
	img.Set(0, 0, color.White)

	f, err := fs.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return jpeg.Encode(f, img, nil)
}
