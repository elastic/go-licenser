# Go Licenser [![Build Status](https://travis-ci.org/elastic/go-licenser.svg?branch=master)](https://travis-ci.org/elastic/go-licenser)

Small license header checker for source files, mainly thought and tested for Go source files although it might work for other ones. The aim of this project is to provide a common
binary that can be used to ensure that code source files contain a license header.

Additionally, when the `-notice` flag is set, generates a `NOTICE` file at the root folder of the project when used with or stdout when using `-d` for dry runs.

## Supported Licenses

* Apache 2.0
* Elastic
* Elastic Cloud

## Supported languages

* Go

## Installing

```
go get -u github.com/elastic/go-licenser
```

## Usage

```
Usage: go-licenser [flags] [path]

  go-licenser walks the specified path recursiely and appends a license Header if the current
  header doesn't match the one found in the file.

  Using the -notice flag a compiled list of the project's dependencies and licenses is compiled
  after the "go.mod" file is inspected. If the dependencies aren't found locally, it will fail.

Options:

  -d	skips rewriting files and returns exitcode 1 if any discrepancies are found.
  -exclude value
    	path to exclude (can be specified multiple times).
  -ext string
    	sets the file extension to scan for. (default ".go")
  -license string
    	sets the license type to check: ASL2, Elastic, Cloud (default "ASL2")
  -licensor string
        sets the name of the licensor (default "Elasticsearch B.V.")
  -notice
    	generates a NOTICE (use -notice-file to change it) file on the folder where it's being run.
  -notice-file string
    	specifies the file where to write the license notice. (default "NOTICE")
  -notice-header string
    	specifies the notice header Go Template.
  -notice-project-name string
    	specifies the notice project name at the top of the Go Template (defaults to folder name).
  -notice-year string
    	specifies the start of the project so the notice file reflects it.
  -version
    	prints out the binary version.
```

## Contributing

See [CONTRIBUTING.md](./CONTRIBUTING.md).

