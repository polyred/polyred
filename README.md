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

The CPU rasterizer can render the bench example ~200fps:

```sh

$ cd bench && go build
$ for i in {1..10}; do perflock -governor 80% ./bench; done
BenchmarkRasterizer     243     4915913 ns/op   203.42101253622673 fps
BenchmarkRasterizer     244     4891938 ns/op   204.41796277875966 fps
BenchmarkRasterizer     242     4884135 ns/op   204.74454534938118 fps
BenchmarkRasterizer     237     4897988 ns/op   204.16546549317803 fps
BenchmarkRasterizer     243     4886564 ns/op   204.64277148523993 fps
BenchmarkRasterizer     243     4912620 ns/op   203.55736857318496 fps
BenchmarkRasterizer     244     4950226 ns/op   202.01097889268084 fps
BenchmarkRasterizer     242     4871713 ns/op   205.26660745409265 fps
BenchmarkRasterizer     244     4926549 ns/op   202.9818438830102 fps
BenchmarkRasterizer     244     4880202 ns/op   204.9095508751482 fps
$ inxi -C
CPU: Topology: 8-Core model: Intel Core i9-9900K bits: 64 type: MT MCP L2 cache: 16.0 MiB Speed: 800 MHz min/max: 800/5000 MHz Core speeds (MHz): 1: 800 2: 800 3: 800 4: 800 5: 800 6: 800 7: 800 8: 801 9: 800 10: 800 11: 800 12: 800 13: 800 14: 800 15: 800 16: 800
```

## License

GNU GPLv3 &copy; [Changkun Ou](https://changkun.de)