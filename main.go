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
	exitFailedToAbstractPath
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
	defaultExludedDirs = []string{"vendor", ".git"}
)

func init() {
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
	code, err := run(args, defaultExludedDirs, extension, dryRun, os.Stdout)
	if err != nil {
		fmt.Println(err)
	}

	os.Exit(code)
}

func run(args, exclDirs []string, ext string, dry bool, out io.Writer) (int, error) {
	var path = defaultPath
	if len(args) > 0 {
		path = args[0]
	}

	if !filepath.IsAbs(path) {
		var err error
		if path, err = filepath.Abs(path); err != nil {
			return exitFailedToAbstractPath, err
		}
	}

	return walk(path, ext, defaultExludedDirs, dry, out)
}

func reportFile(out io.Writer, f string) {
	cwd, _ := filepath.Abs(filepath.Dir(os.Args[0]))
	rel, err := filepath.Rel(cwd, f)
	if err != nil {
		rel = f
	}
	fmt.Fprintf(out, defaultFormat, rel)
}

func walk(p, ext string, exclude []string, dry bool, out io.Writer) (int, error) {
	var code int
	return code, filepath.Walk(p, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			code = exitFailedToWalkPath
			return err
		}

		if info.IsDir() && stringInSlice(info.Name(), exclude) {
			return filepath.SkipDir
		}

		code, err = addOrCheckLicense(path, ext, info, dry, out)
		return err
	})
}

func addOrCheckLicense(path, ext string, info os.FileInfo, dry bool, out io.Writer) (int, error) {
	if info.IsDir() || filepath.Ext(path) != ext {
		return exitDefault, nil
	}

	f, e := os.Open(path)
	if e != nil {
		return exitFailedToOpenWalkFile, e
	}
	defer f.Close()

	if licensing.ContainsHeader(f, Header) {
		return exitDefault, nil
	}

	if dry {
		reportFile(out, path)
		return exitSourceNeedsToBeRewritten, nil
	}

	if err := licensing.RewriteFileWithHeader(path, headerBytes); err != nil {
		return errFailedRewrittingFile, err
	}

	return exitDefault, nil
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
