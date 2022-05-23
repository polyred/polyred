// Copyright 2022 The Polyred Authors. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

// Package model offers a set of pre-defined 3D models.
package model

import (
	"log"
	"path"
	"runtime"

	"poly.red/scene"
)

// StanfordBunny returns a mesh of stanford bunny.
func StanfordBunny() *scene.Group {
	return MustLoad(fix("../internal/testdata/bunny.obj"))
}

// ChineseDragon returns a mesh of chinese dragon.
func ChineseDragon() *scene.Group {
	return MustLoad(fix("../internal/testdata/dragon.obj"))
}

// fix returns the absolute path of a given relative path
func fix(p string) string {
	_, filename, _, ok := runtime.Caller(1)
	if !ok {
		log.Fatalf("cannot get runtime caller")
	}
	return path.Join(path.Dir(filename), p)
}
