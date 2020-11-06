# ddd

ddd is a 3d rasterizer written in pure Go.

## Usage

```go
width, height := 800, 500

// create rasterizer
r := ddd.NewRasterizer(width, height)

// load obj model
m, err := ddd.LoadOBJ("path/to/bunny.obj")
if err != nil {
    panic(fmt.Sprintf("cannot load obj model, err: %v", err))
}

// set model matrix
m.SetScale(ddd.Vector{1500, 1500, 1500, 0})
m.SetTranslate(ddd.Vector{-700, -5, 350, 1})

// set texture
err = m.SetTexture("path/to/texture.jpg", 150)
if err != nil {
    panic(fmt.Sprintf("cannot load model texture, err: %v", err))
}

// set the camera
r.SetCamera(ddd.NewPerspectiveCamera(
    ddd.Vector{-550, 194, 734, 1},
    ddd.Vector{-1000, 0, 0, 1},
    ddd.Vector{0, 1, 1, 0},
    float64(width)/float64(height),
    100, 600, 45,
))
r.SetCamera(ddd.NewOrthographicCamera(
    ddd.Vector{-550, 194, 734, 1},
    ddd.Vector{-1000, 0, 0, 1},
    ddd.Vector{0, 1, 1, 0},
    -float64(width)/2, float64(width)/2,
    float64(height)/2, -float64(height)/2,
    200, -200,
))

// construct scene graph
s := ddd.NewScene()
r.SetScene(s)
s.AddMesh(m)
l := ddd.NewPointLight(
    color.RGBA{255, 255, 255, 255},
    ddd.Vector{-200, 250, 600, 1}, 0.5, 0.6, 1,
)
s.AddLight(l)

// render!
r.Render()

// save result
r.Save("path/to/render.jpg")
```

![](./bench/render.jpg)

See complete example [here](./bench/main.go).

## Benchmark

The CPU rasterizer can render the bench example >300fps:

```sh

$ cd bench && go build
$ for i in {1..10}; do perflock -governor 80% ./bench; done
BenchmarkRasterizer     343     3498860 ns/op   285.80737726002184 fps
BenchmarkRasterizer     344     3459285 ns/op   289.0770780667103 fps
BenchmarkRasterizer     345     3472535 ns/op   287.9740592967385 fps
BenchmarkRasterizer     345     3473249 ns/op   287.91486012088393 fps
BenchmarkRasterizer     337     3456515 ns/op   289.3087401616947 fps
BenchmarkRasterizer     345     3494367 ns/op   286.1748637163755 fps
BenchmarkRasterizer     345     3470063 ns/op   288.17920596830663 fps
BenchmarkRasterizer     345     3481662 ns/op   287.2191499347151 fps
BenchmarkRasterizer     345     3464998 ns/op   288.60045518063794 fps
BenchmarkRasterizer     344     3468320 ns/op   288.32403007796285 fps
$ inxi -C
CPU: Topology: 8-Core model: Intel Core i9-9900K bits: 64 type: MT MCP L2 cache: 16.0 MiB Speed: 800 MHz min/max: 800/5000 MHz Core speeds (MHz): 1: 800 2: 800 3: 800 4: 800 5: 800 6: 800 7: 800 8: 801 9: 800 10: 800 11: 800 12: 800 13: 800 14: 800 15: 800 16: 800
```

## License

GNU GPLv3 &copy; [Changkun Ou](https://changkun.de)