// Package model offers a set of pre-defined 3D models.
package model

import (
	"log"
	"path"
	"runtime"

	"poly.red/geometry/mesh"
)

// TODO: use embed? But there is a limitation that the embed can only embed
// files less than 2GB. See https://golang.org/issue/47627

// StanfordBunny returns the path to locate stanford bunny.
func StanfordBunnyAs[T mesh.Mesh]() T {
	return mesh.MustLoadAs[T](fix("../internal/testdata/bunny.obj"))
}

// fix returns the absolute path of a given relative path
func fix(p string) string {
	_, filename, _, ok := runtime.Caller(1)
	if !ok {
		log.Fatalf("cannot get runtime caller")
	}
	return path.Join(path.Dir(filename), p)
}
