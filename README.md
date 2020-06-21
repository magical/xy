This is a collection of tools for working with file formats
found in Pokémon X and Y.  They are written in Go and expect
to be located at `xy` in the root of your your GOPATH.

    git clone https://github.com/magical/xy $GOPATH/xy

The `*.go` files in the repo root (this directory) are scripts
to extract various bits of data. Each file is a separate tool.
Run one with `go run whatever.go`.

The directories (garc, lz, text, etc) are packages that can be
imported. `go doc ./whatever` for the API.

