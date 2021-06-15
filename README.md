# ddd

A software hybrid renderer written in pure Go.

```go
import "changkun.de/x/ddd"
```

_Caution: experiment, expect it to break at any time. Use it at your own risk._

## Features

- OBJ file format
- hybrid rendering: rasterization + ray tracing
- backface, viewfrustum, and occlusion culling
- isotropic texturing
- deferred shading
- shadow mapping
- anti-aliasing
- concurrent processing
- built-in geometries (plane)

<!-- ## Performance

ddd renders in parallel and utilizes all the CPUs.

| model | resolution | msaa | rendering (ms) |
|:------|:-----------|:-----|:---------------|
| stanford bunny |  -->

![](./examples/teaser.png)


## More Examples

| Example | Code |
|:-------:|:-----:|
|<img src="./examples/bunny/bunny.png" width="300px"/>|[bunny](./examples/bunny/bunny.go)|
|<img src="./examples/dragon/dragon.png" width="300px"/>|[dragon](./examples/dragon/dragon.go)|

## License

Copyright &copy; 2020-2021 [Changkun Ou](https://changkun.de). All rights reserved.