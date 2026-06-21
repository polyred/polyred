package material

import (
	"sync"

	"poly.red/buffer"
	"poly.red/color"
)

// ID identifies a material in the process-wide pool.
//
// ID semantics:
//   - 0 is always the default material (seeded in init, cannot be deleted).
//   - Positive IDs are assigned by Put, incrementally.
//   - A negative ID is intentionally never in the pool: Get returns nil for it,
//     which the renderer reads as "use vertex color directly". Producers set it
//     explicitly (e.g. model/plane.go and geometry/primitive/polygon.go use -1);
//     consumers are render/raster.go and render/gpudeferred.go via Get.
type ID int64

// pool is the process-wide material registry. It is global mutable state:
// material IDs are shared across all renderers and scenes, and a material lives
// until Del removes it (creation is at asset-load time, not per frame, so it does
// not grow per frame, but it is never auto-freed). The GPU deferred path
// additionally builds its own per-frame materials table keyed by *BlinnPhong
// (render/gpudeferred.go), so two indexing schemes coexist. De-globalizing this
// into scene-owned materials is a design change with real blast radius (every
// material.Get in render); see specs/foundations/material-ownership.md.
var pool struct {
	mu      sync.RWMutex
	allocID int64 // incremental
	idToMat map[ID]Material
	matToId map[Material]ID
}

func init() {
	pool.idToMat = make(map[ID]Material)
	pool.matToId = make(map[Material]ID)

	// Put the first material as the default material, and its material ID is always 0.
	Put(defaultMaterial)
}

// Get returns the associated material of the given ID.
func Get(id ID) Material {
	pool.mu.RLock()
	defer pool.mu.RUnlock()

	m, ok := pool.idToMat[id]
	if !ok {
		return nil
	}
	return m
}

// Resolve is the single material-resolution path for the renderer (CPU and GPU).
// It returns the BlinnPhong material for id, or nil when id resolves to no usable
// material: a negative ID (the "use vertex color" hint), an absent ID, or a
// non-BlinnPhong material. Callers treat nil as "no material" (vertex color).
func Resolve(id ID) *BlinnPhong {
	m := Get(id)
	if m == nil {
		return nil
	}
	bp, _ := m.(*BlinnPhong)
	return bp
}

// Put puts the given material to the centralized material pool.
func Put(m Material) (id ID, ok bool) {
	pool.mu.Lock()
	defer pool.mu.Unlock()

	id, ok = pool.matToId[m]
	if ok {
		return id, false
	}

	newID := ID(pool.allocID)
	pool.idToMat[newID] = m
	pool.matToId[m] = newID
	pool.allocID++

	return newID, true
}

// Del deletes the associated material of the given material ID.
// The function returns true if deleted.
func Del(id ID) bool {
	if id == 0 {
		panic("material: default material cannot be deleted!")
	}

	pool.mu.Lock()
	defer pool.mu.Unlock()

	mat, ok := pool.idToMat[id]
	if !ok {
		return false
	}
	delete(pool.matToId, mat)
	delete(pool.idToMat, id)
	return true
}

var defaultMaterial = &BlinnPhong{
	Standard: Standard{
		FlatShading:      false,
		AmbientOcclusion: false,
		ReceiveShadow:    false,
		Texture:          buffer.NewUniformTexture(color.Blue),
		name:             "default",
	},
	Ambient:   color.FromValue(0.7, 0.7, 0.7, 1.0),
	Diffuse:   color.FromValue(0.7, 0.7, 0.7, 1.0),
	Specular:  color.FromValue(0.5, 0.5, 0.5, 1.0),
	Shininess: 30.0,
}
