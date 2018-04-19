# Go licenser

Small zero dependency license header checker for Go source files.

## Installing

```
go get -u github.com/elastic/go-licenser
```

## Usage

```
Usage: go-licenser [flags] [path]

  go-licenser walks the specified path recursiely and appends a license Header if the current
  header doesn't match the one found in the file.

Options:

  -d	skips rewriting files and returns exitcode 1 if any discrepancies are found.
  -ext string
    	sets the file extension to scan for. (default ".go")
```

## Contributing

See [CONTRIBUTING.md](./CONTRIBUTING.md).

