# errorwrapper-go

[![Go Reference](https://pkg.go.dev/badge/github.com/colduction/errorwrapper-go/v1.svg)](https://pkg.go.dev/github.com/colduction/errorwrapper-go/v1)
[![Go Report Card](https://goreportcard.com/badge/github.com/colduction/errorwrapper-go)](https://goreportcard.com/report/github.com/colduction/errorwrapper-go)
[![License: MIT](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)

A tiny Go library for building structured, prefix-chained error messages.

---

## Install

```sh
go get -u github.com/colduction/errorwrapper-go/v1
```

## Usage

```go
import errorwrapper "github.com/colduction/errorwrapper-go/v1"

e1 := errorwrapper.New('.', "database")
err1 := e1.NewErrorString("connection refused")
// → "database: connection refused"

e2 := errorwrapper.New('.', "repository")
err2 := e2.NewError(err1)
// → "repository.database: connection refused"

e3 := errorwrapper.New('.', "service")
err3 := e3.NewError(err2, "query failed")
// → "service.repository.database: [query failed] connection refused"
```

## API

| Function                                                  | Description                                                            |
| --------------------------------------------------------- | ---------------------------------------------------------------------- |
| `New(joiner byte, prefix ...string) ErrorWrapper`         | Creates a new wrapper with an optional prefix and joiner (default `.`) |
| `(ew) NewError(err error, msg ...string) error`           | Wraps an existing error, merging any prefix chain                      |
| `(ew) NewErrorString(errStr string, msg ...string) error` | Creates and wraps a new error from a string                            |

- Returns `nil` when given a `nil` error or an empty string.
- Prefix chains are flattened automatically when wrappers are nested.
