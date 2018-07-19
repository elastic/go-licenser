// Licensed to Elasticsearch B.V. under one or more contributor
// license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright
// ownership. Elasticsearch B.V. licenses this file to you under
// the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/elastic/go-licenser/licensing"
)

const (
	defaultExt    = ".go"
	defaultPath   = "."
	defaultFormat = "%s: is missing the license header\n"
)

const (
	exitDefault = iota
	exitSourceNeedsToBeRewritten
	exitFailedToStatTree
	exitFailedToStatFile
	exitFailedToWalkPath
	exitFailedToOpenWalkFile
	errFailedRewrittingFile
)

var usageText = `
Usage: go-licenser [flags] [path]

  go-licenser walks the specified path recursiely and appends a license Header if the current
  header doesn't match the one found in the file.

Options:

`[1:]

// Header is the licenser that all of the files in the repository must have.
var Header = []string{
	`// Licensed to Elasticsearch B.V. under one or more contributor`,
	`// license agreements. See the NOTICE file distributed with`,
	`// this work for additional information regarding copyright`,
	`// ownership. Elasticsearch B.V. licenses this file to you under`,
	`// the Apache License, Version 2.0 (the "License"); you may`,
	`// not use this file except in compliance with the License.`,
	`// You may obtain a copy of the License at`,
	`//`,
	`//     http://www.apache.org/licenses/LICENSE-2.0`,
	`//`,
	`// Unless required by applicable law or agreed to in writing,`,
	`// software distributed under the License is distributed on an`,
	`// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY`,
	`// KIND, either express or implied.  See the License for the`,
	`// specific language governing permissions and limitations`,
	`// under the License.`,
}

var (
	dryRun             bool
	extension          string
	args               []string
	headerBytes        []byte
	exclude            sliceFlag
	defaultExludedDirs = []string{"vendor", ".git"}
)

type sliceFlag []string

func (f *sliceFlag) String() string {
	var s string
	for _, i := range *f {
		s += i + " "
	}
	return s
}

func (f *sliceFlag) Set(value string) error {
	*f = append(*f, value)
	return nil
}

func init() {
	flag.Var(&exclude, "exclude", `path to exclude (can be specified multiple times).`)
	flag.BoolVar(&dryRun, "d", false, `skips rewriting files and returns exitcode 1 if any discrepancies are found.`)
	flag.StringVar(&extension, "ext", defaultExt, "sets the file extension to scan for.")
	flag.Usage = usageFlag
	flag.Parse()
	args = flag.Args()
	for i := range Header {
		headerBytes = append(headerBytes, []byte(Header[i])...)
		headerBytes = append(headerBytes, []byte("\n")...)
	}
}

func main() {
	err := run(args, exclude, extension, dryRun, os.Stdout)
	if err != nil && err.Error() != "<nil>" {
		fmt.Fprint(os.Stderr, err)
	}

	os.Exit(Code(err))
}

func run(args, exclude []string, ext string, dry bool, out io.Writer) error {
	var path = defaultPath
	if len(args) > 0 {
		path = args[0]
	}

	if _, err := os.Stat(path); err != nil {
		return &Error{err: err, code: exitFailedToStatTree}
	}

	return walk(path, ext, exclude, dry, out)
}

func reportFile(out io.Writer, f string) {
	cwd, _ := filepath.Abs(filepath.Dir(os.Args[0]))
	rel, err := filepath.Rel(cwd, f)
	if err != nil {
		rel = f
	}
	fmt.Fprintf(out, defaultFormat, rel)
}

func walk(p, ext string, exclude []string, dry bool, out io.Writer) error {
	var err error
	filepath.Walk(p, func(path string, info os.FileInfo, walkErr error) error {
		if walkErr != nil {
			err = &Error{err: walkErr, code: exitFailedToWalkPath}
			return walkErr
		}

		var currentPath = cleanPathPrefixes(
			strings.Replace(path, p, "", 1),
			[]string{"/"},
		)

		var excludedDir = info.IsDir() && stringInSlice(info.Name(), defaultExludedDirs)
		if needsExclusion(currentPath, exclude) || excludedDir {
			return filepath.SkipDir
		}

		if e := addOrCheckLicense(path, ext, info, dry, out); e != nil {
			err = e
		}

		return nil
	})

	return err
}

func addOrCheckLicense(path, ext string, info os.FileInfo, dry bool, out io.Writer) error {
	if info.IsDir() || filepath.Ext(path) != ext {
		return nil
	}

	f, e := os.Open(path)
	if e != nil {
		return &Error{err: e, code: exitFailedToOpenWalkFile}
	}
	defer f.Close()

	if licensing.ContainsHeader(f, Header) {
		return nil
	}

	if dry {
		reportFile(out, path)
		return &Error{code: exitSourceNeedsToBeRewritten}
	}

	if err := licensing.RewriteFileWithHeader(path, headerBytes); err != nil {
		return &Error{err: err, code: errFailedRewrittingFile}
	}

	return nil
}

func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

func usageFlag() {
	fmt.Fprintf(os.Stderr, usageText)
	flag.PrintDefaults()
}
