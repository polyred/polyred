package material

import (
	"sync"

	"poly.red/buffer"
	"poly.red/color"
)

// ID represents the ID of a material.
//
// Design Decision (at the moment): This ID may be negative to hint
// the renderer to use vertex color directly.
type ID int64

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
