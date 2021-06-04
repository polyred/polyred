# thread [![PkgGoDev](https://pkg.go.dev/badge/golang.design/x/thread)](https://pkg.go.dev/golang.design/x/thread) [![Go Report Card](https://goreportcard.com/badge/golang.design/x/thread)](https://goreportcard.com/report/golang.design/x/thread) ![thread](https://github.com/golang-design/thread/workflows/thread/badge.svg?branch=main)

Package thread provides threading facilities, such as scheduling
calls on a specific thread, local storage, etc.

```go
import "golang.design/x/thread"
```

## Quick Start

```go
th := thread.New()

th.Call(func() {
    // call on the created thread
})
```

## License

MIT &copy; 2021 The golang.design Initiative