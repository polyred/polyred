# ddd

A performant 3D renderer written in pure Go.

```go
import "changkun.de/x/ddd"
```

## Features

- OBJ file format
- rasterization
- backface, viewfrustum, and occlusion culling
- texture and mipmap
- deferred shading
- anti-aliasing

## Usage

```go
width, height := 1920, 1080

// Create scene graph
s := rend.NewScene()

// Specify camera settings
c := camera.NewPerspectiveCamera(
    math.NewVector(-0.5, 0.5, 0.5, 1),
    math.NewVector(0, 0, -0.5, 1),
    math.NewVector(0, 1, 0, 0),
    45,
    float64(width)/float64(height),
    -0.1,
    -3,
)
s.UseCamera(c)

// Add light sources
l := light.NewPointLight(color.RGBA{0, 0, 0, 255}, math.NewVector(-200, 250, 600, 1))
s.AddLight(l)

// Load assets
m := loadMesh("../testdata/bunny.obj")
tex := loadTexture("../testdata/bunny.png")
mat := material.NewBlinnPhongMaterial(tex, color.RGBA{0, 125, 255, 255}, 0.6, 1, 0.5, 150)
m.UseMaterial(mat)
m.Rotate(math.NewVector(0, 1, 0, 0), -math.Pi/6)
m.Translate(0, -0, -0.4)
s.AddMesh(m)

// Load another assets
m = loadMesh("../testdata/ground.obj")
tex = loadTexture("../testdata/ground.png")
mat = material.NewBlinnPhongMaterial(tex, color.RGBA{0, 125, 255, 255}, 0.6, 1, 0.5, 150)
m.UseMaterial(mat)
m.Rotate(math.NewVector(0, 1, 0, 0), -math.Pi/6)
m.Translate(0, -0, -0.4)
s.AddMesh(m)

// Create the rasterizer and render it.
r := rend.NewRasterizer(width, height, 1)
r.Save(r.Render(s), "./render.jpg")
```

![](./bench/render.jpg)

See complete example [here](./bench/main.go).

## Benchmark

```sh
$ cd bench && go build
$ for i in {1..10}; do perflock -governor 80% ./bench; done
BenchmarkRasterizer     40      28539371 ns/op  35.039314636612 fps
BenchmarkRasterizer     42      29070529 ns/op  34.39909882616859 fps
BenchmarkRasterizer     42      30322510 ns/op  32.97880023784311 fps
BenchmarkRasterizer     38      29042832 ns/op  34.43190388595713 fps
BenchmarkRasterizer     42      28728505 ns/op  34.8086334461191 fps
BenchmarkRasterizer     42      28635385 ns/op  34.921828360261266 fps
BenchmarkRasterizer     39      28476114 ns/op  35.117151167466176 fps
BenchmarkRasterizer     42      28535627 ns/op  35.04391194908736 fps
BenchmarkRasterizer     42      29167186 ns/op  34.2851038149515 fps
BenchmarkRasterizer     42      28533852 ns/op  35.04609191917026 fps

$ inxi -C
CPU: Topology: 8-Core model: Intel Core i9-9900K bits: 64 type: MT MCP L2 cache: 16.0 MiB Speed: 800 MHz min/max: 800/5000 MHz Core speeds (MHz): 1: 800 2: 800 3: 800 4: 800 5: 800 6: 800 7: 800 8: 801 9: 800 10: 800 11: 800 12: 800 13: 800 14: 800 15: 800 16: 800
```

## License

Copyright &copy; 2020-2021 [Changkun Ou](https://changkun.de). All rights reserved.