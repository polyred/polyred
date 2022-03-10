// Package model offers a set of pre-defined 3D models.
package model

import (
	"log"
	"path"
	"runtime"

	"poly.red/geometry/mesh"
)

// StanfordBunny returns a mesh of stanford bunny.
func StanfordBunnyAs[T mesh.Mesh[float32]]() T {
	return mesh.MustLoadAs[T](fix("../internal/testdata/bunny.obj"))
}

// ChineseDragon returns a mesh of chinese dragon.
func ChineseDragonAs[T mesh.Mesh[float32]]() T {
	return mesh.MustLoadAs[T](fix("../internal/testdata/dragon.obj"))
}

// fix returns the absolute path of a given relative path
func fix(p string) string {
	_, filename, _, ok := runtime.Caller(1)
	if !ok {
		log.Fatalf("cannot get runtime caller")
	}
	return path.Join(path.Dir(filename), p)
}
