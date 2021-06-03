package material

// // Texture is a texture
// type Texture struct {
// 	Shininess float64
// 	data      image.Image
// 	width     int
// 	height    int
// }

// // SetTexture sets the texture of a given mesh.
// func (m *TriangleMesh) SetTexture(path string, shininess float64) error {
// 	tex, err := os.Open(path)
// 	if err != nil {
// 		return fmt.Errorf("cannot read texture: %v", err)
// 	}
// 	defer tex.Close()

// 	data, _, err := image.Decode(tex)
// 	if err != nil {
// 		return fmt.Errorf("decode texture error: %v", err)
// 	}

// 	m.texture = &Texture{
// 		Shininess: shininess,
// 		data:      data,
// 		width:     data.Bounds().Max.X,
// 		height:    data.Bounds().Max.Y,
// 	}

// 	return nil
// }
