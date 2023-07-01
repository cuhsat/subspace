# ![](assets/inline.png) Subspace [![Go Reference](https://pkg.go.dev/badge/github.com/cuhsat/subspace.svg)](https://pkg.go.dev/github.com/cuhsat/subspace) [![Go Report Card](https://goreportcard.com/badge/github.com/cuhsat/subspace?style=flat-square)](https://goreportcard.com/report/github.com/cuhsat/subspace) [![Release](https://img.shields.io/github/release/cuhsat/subspace.svg?style=flat-square)](https://github.com/cuhsat/subspace/releases/latest)
Subspace is a fast, memory only signal broker using atomic network operations.

## How to
Start a subspace:
```sh
$ subspace
```

Send a signal:
```sh
$ echo foo | ss
```

Scan for signals:
```sh
$ ss
```

## License
Released under the [MIT License](LICENSE).
