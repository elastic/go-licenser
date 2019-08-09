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
	"gopkg.in/src-d/go-license-detector.v2/licensedb"
)

const (
	defaultExt      = ".go"
	defaultPath     = "."
	defaultLicense  = "ASL2"
	defaultLicensor = "Elasticsearch B.V."
	defaultFormat   = "%s: is missing the license header\n"
)

const (
	exitDefault = iota
	exitSourceNeedsToBeRewritten
	exitFailedToStatTree
	exitFailedToStatFile
	exitFailedToWalkPath
	exitFailedToOpenWalkFile
	errFailedRewrittingFile
	errUnknownLicense
	errOpenFileFailed
	errGenerateNoticeFailed
)

var usageText = `
Usage: go-licenser [flags] [path]

  go-licenser walks the specified path recursiely and appends a license Header if the current
  header doesn't match the one found in the file.

  Using the -notice flag a compiled list of the project's dependencies and licenses is compiled
  after the "go.mod" file is inspected. If the dependencies aren't found locally, it will fail.

Options:

`[1:]

// Headers is the map of supported licenses
var Headers = map[string][]string{
	"ASL2": {
		`// Licensed to %s under one or more contributor`,
		`// license agreements. See the NOTICE file distributed with`,
		`// this work for additional information regarding copyright`,
		`// ownership. %s licenses this file to you under`,
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
	},
	"Elastic": {
		`// Copyright Elasticsearch B.V. and/or licensed to Elasticsearch B.V. under one`,
		`// or more contributor license agreements. Licensed under the Elastic License;`,
		`// you may not use this file except in compliance with the Elastic License.`,
	},
	"Cloud": {
		`// ELASTICSEARCH CONFIDENTIAL`,
		`// __________________`,
		`//`,
		`//  Copyright Elasticsearch B.V. All rights reserved.`,
		`//`,
		`// NOTICE:  All information contained herein is, and remains`,
		`// the property of Elasticsearch B.V. and its suppliers, if any.`,
		`// The intellectual and technical concepts contained herein`,
		`// are proprietary to Elasticsearch B.V. and its suppliers and`,
		`// may be covered by U.S. and Foreign Patents, patents in`,
		`// process, and are protected by trade secret or copyright`,
		`// law.  Dissemination of this information or reproduction of`,
		`// this material is strictly forbidden unless prior written`,
		`// permission is obtained from Elasticsearch B.V.`,
	},
}

var (
	dryRun             bool
	showVersion        bool
	generateNotice     bool
	noticeYear         string
	noticeFile         string
	noticeHeader       string
	noticeProject      string
	extension          string
	args               []string
	license            string
	licensor           string
	exclude            sliceFlag
	defaultExludedDirs = []string{"vendor", ".git"}
)

type runParams struct {
	args          []string
	license       string
	licensor      string
	exclude       []string
	ext           string
	dry           bool
	notice        bool
	noticeYear    string
	noticeFile    string
	noticeHeader  string
	noticeProject string
	out           io.Writer
	analyseFunc   func(args ...string) []licensedb.Result
}

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
	flag.BoolVar(&showVersion, "version", false, `prints out the binary version.`)
	flag.BoolVar(&generateNotice, "notice", false, `generates a NOTICE (use -notice-file to change it) file on the folder where it's being run.`)
	flag.StringVar(&noticeYear, "notice-year", "", `specifies the start of the project so the notice file reflects it.`)
	flag.StringVar(&noticeFile, "notice-file", "NOTICE", `specifies the file where to write the license notice.`)
	flag.StringVar(&noticeHeader, "notice-header", "", `specifies the notice header Go Template.`)
	flag.StringVar(&noticeProject, "notice-project-name", "", `specifies the notice project name at the top of the Go Template (defaults to folder name).`)
	flag.StringVar(&extension, "ext", defaultExt, "sets the file extension to scan for.")
	flag.StringVar(&license, "license", defaultLicense, "sets the license type to check: ASL2, Elastic, Cloud")
	flag.StringVar(&licensor, "licensor", defaultLicensor, "sets the name of the licensor")
	flag.Usage = usageFlag
	flag.Parse()
	args = flag.Args()
}

func main() {
	if showVersion {
		fmt.Printf("go-licenser %s (%s)\n", version, commit)
		return
	}

	err := run(runParams{
		args:          args,
		license:       license,
		licensor:      licensor,
		exclude:       exclude,
		ext:           extension,
		dry:           dryRun,
		notice:        generateNotice,
		noticeYear:    noticeYear,
		noticeFile:    noticeFile,
		noticeHeader:  noticeHeader,
		noticeProject: noticeProject,
		out:           os.Stdout,
		analyseFunc:   licensedb.Analyse,
	})
	if err != nil && err.Error() != "<nil>" {
		fmt.Fprint(os.Stderr, err)
	}

	os.Exit(Code(err))
}

func run(params runParams) error {
	header, ok := Headers[params.license]
	if !ok {
		return &Error{err: fmt.Errorf("unknown license: %s", params.license), code: errUnknownLicense}
	}

	var headerBytes []byte
	for i, line := range header {
		if strings.Contains(line, "%s") {
			header[i] = fmt.Sprintf(line, params.licensor)
		}
		headerBytes = append(headerBytes, []byte(header[i])...)
		headerBytes = append(headerBytes, []byte("\n")...)
	}

	var path = defaultPath
	if len(params.args) > 0 {
		path = params.args[0]
	}

	if _, err := os.Stat(path); err != nil {
		return &Error{err: err, code: exitFailedToStatTree}
	}

	var walkErr error
	if err := walk(path, params.ext, params.license, headerBytes,
		params.exclude, params.dry, params.out,
	); err != nil {
		walkErr = err
	}

	if !params.notice {
		return walkErr
	}

	if err := doNotice(path, params); err != nil {
		return err
	}

	return walkErr
}

func reportFile(out io.Writer, f string) {
	cwd, _ := filepath.Abs(filepath.Dir(os.Args[0]))
	rel, err := filepath.Rel(cwd, f)
	if err != nil {
		rel = f
	}
	fmt.Fprintf(out, defaultFormat, rel)
}

func walk(p, ext, license string, headerBytes []byte, exclude []string, dry bool, out io.Writer) error {
	var err error
	filepath.Walk(p, func(path string, info os.FileInfo, walkErr error) error {
		if walkErr != nil {
			err = &Error{err: walkErr, code: exitFailedToWalkPath}
			return walkErr
		}

		var currentPath = cleanPathPrefixes(
			strings.Replace(path, p, "", 1),
			[]string{string(os.PathSeparator)},
		)

		var excludedDir = info.IsDir() && stringInSlice(info.Name(), defaultExludedDirs)
		if needsExclusion(currentPath, exclude) || excludedDir {
			return filepath.SkipDir
		}

		if e := addOrCheckLicense(path, ext, license, headerBytes, info, dry, out); e != nil {
			err = e
		}

		return nil
	})

	return err
}

func addOrCheckLicense(path, ext, license string, headerBytes []byte, info os.FileInfo, dry bool, out io.Writer) error {
	if info.IsDir() || filepath.Ext(path) != ext {
		return nil
	}

	f, e := os.Open(path)
	if e != nil {
		return &Error{err: e, code: exitFailedToOpenWalkFile}
	}
	defer f.Close()

	if licensing.ContainsHeader(f, Headers[license]) {
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

func openTruncateFile(p string) (*os.File, error) {
	f, err := os.OpenFile(p, os.O_CREATE|os.O_WRONLY|os.O_APPEND, os.ModePerm)
	if err != nil {
		return nil, err
	}

	f.Truncate(0)
	f.Seek(0, 0)
	return f, nil
}
