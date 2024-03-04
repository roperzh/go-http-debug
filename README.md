# go-http-debug

`go-http-debug` is a set of middlewares that can be used to print information
about requests and responses of a Go web server during development.

It's meant to be used as a toy/experiment, and as an aid to develop side projects.

## Usage

For detailed usage read the Go documentation, as a summary, this library
exposes a few middlewares:

### stdout

Prints debugging information to stdout, supports two formats:

```go
// RawStdout produces pretty printed output
handler = httpdebug.RawStdout(handler)

// JSONStdout prints a JSON structure representing the transaction
handler = httpdebug.JSONStdout(handler)
```

### web

The web middleware starts a web server that serves a web UI.

A web interface with default values can be started with:

```go
handler = httpdebug.WebUI(handler)
```

If you want to customize some of the defaults:

```go
debug := httpdebug.NewWebUIHandler(
    httpdebug.WithAddress(":1234"),
    httpdebug.WithoutMessage(),
)

handler = debug.WebUI(handler)
```

## License

[GPL-3.0](./COPYING)
