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
	"bytes"
	"crypto/sha1"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"syscall"
	"testing"

	"gopkg.in/src-d/go-license-detector.v2/licensedb"
)

const fixtures = "fixtures"

var update = flag.Bool("update", false, "updates the golden files with the latest iteration of the code")

func copyFixtures(t *testing.T, dest string) func() {
	if err := copy(fixtures, dest); err != nil {
		t.Fatal(err)
	}
	return func() {
		if err := os.RemoveAll(dest); err != nil {
			t.Fatal(err)
		}
	}
}

func copy(src, dest string) error {
	info, err := os.Stat(src)
	if err != nil {
		return err
	}
	if info.IsDir() {
		return dcopy(src, dest, info)
	}
	return fcopy(src, dest, info)
}

func fcopy(src, dest string, info os.FileInfo) error {
	f, err := os.Create(
		strings.Replace(dest, ".testdata", ".go", 1),
	)
	if err != nil {
		return err
	}
	defer f.Close()

	if err = os.Chmod(f.Name(), info.Mode()); err != nil {
		return err
	}

	s, err := os.Open(src)
	if err != nil {
		return err
	}
	defer s.Close()

	_, err = io.Copy(f, s)
	return err
}

func dcopy(src, dest string, info os.FileInfo) error {
	if err := os.MkdirAll(dest, info.Mode()); err != nil {
		return err
	}

	infs, err := ioutil.ReadDir(src)
	if err != nil {
		return err
	}

	for i := range infs {
		var source = filepath.Join(src, infs[i].Name())
		var destination = filepath.Join(dest, infs[i].Name())
		if err := copy(source, destination); err != nil {
			return err
		}
	}

	return nil
}

var goModAnalyseFunc = genAnalyseFunc([]licensedb.Result{
	{
		Arg:     "github.com/hashicorp/multierror-go",
		Matches: []licensedb.Match{{License: "MPL-2.0"}},
	},
	{
		Arg:     "github.com/sirkon/goproxy",
		Matches: []licensedb.Match{{License: "MIT"}},
	},
	{
		Arg:     "gopkg.in/src-d/go-license-detector.v2",
		Matches: []licensedb.Match{{License: "Apache-2.0"}},
	},
})

func genAnalyseFunc(r []licensedb.Result) func(args ...string) []licensedb.Result {
	return func(args ...string) []licensedb.Result { return r }
}

func Test_run(t *testing.T) {
	tests := []struct {
		name       string
		args       runParams
		want       int
		err        error
		wantOutput string
		wantGolden bool
	}{
		{
			name: "Run a diff prints a list of files that need the default license header",
			args: runParams{
				args:     []string{"testdata"},
				license:  defaultLicense,
				licensor: defaultLicensor,
				exclude:  []string{"excludedpath", "x-pack", "cloud"},
				ext:      defaultExt,
				dry:      true,
			},
			want: 1,
			err:  &Error{code: 1},
			wantOutput: `
testdata/multilevel/doc.go: is missing the license header
testdata/multilevel/main.go: is missing the license header
testdata/multilevel/sublevel/autogen.go: is missing the license header
testdata/multilevel/sublevel/doc.go: is missing the license header
testdata/multilevel/sublevel/partial.go: is missing the license header
testdata/singlelevel/doc.go: is missing the license header
testdata/singlelevel/main.go: is missing the license header
testdata/singlelevel/wrapper.go: is missing the license header
`[1:],
		},
		{
			name: "Run a diff prints a list of files that need the default license header and notice",
			args: runParams{
				args:          []string{"testdata"},
				license:       defaultLicense,
				licensor:      defaultLicensor,
				exclude:       []string{"excludedpath", "x-pack", "cloud"},
				ext:           defaultExt,
				dry:           true,
				notice:        true,
				noticeProject: "SomeProject",
				analyseFunc:   goModAnalyseFunc,
			},
			want: 1,
			err:  &Error{code: 1},
			wantOutput: `
testdata/multilevel/doc.go: is missing the license header
testdata/multilevel/main.go: is missing the license header
testdata/multilevel/sublevel/autogen.go: is missing the license header
testdata/multilevel/sublevel/doc.go: is missing the license header
testdata/multilevel/sublevel/partial.go: is missing the license header
testdata/singlelevel/doc.go: is missing the license header
testdata/singlelevel/main.go: is missing the license header
testdata/singlelevel/wrapper.go: is missing the license header
Dumping NOTICE to output...

SomeProject
Copyright 2019 Elasticsearch B.V.

This product includes software developed at Elasticsearch B.V. and
third-party software developed by the licenses listed below.

=========================================================================

gopkg.in/src-d/go-license-detector.v2    Apache-2.0
github.com/sirkon/goproxy                MIT
github.com/hashicorp/multierror-go       MPL-2.0

=========================================================================
`[1:],
		},
		{
			name: "Run a diff prints a list of files that need the Elastic license header",
			args: runParams{
				args:     []string{"testdata"},
				license:  "Elastic",
				licensor: defaultLicensor,
				ext:      defaultExt,
				dry:      true,
			},
			want: 1,
			err:  &Error{code: 1},
			wantOutput: `
testdata/cloud/doc.go: is missing the license header
testdata/cloud/wrong.go: is missing the license header
testdata/excludedpath/file.go: is missing the license header
testdata/multilevel/doc.go: is missing the license header
testdata/multilevel/main.go: is missing the license header
testdata/multilevel/sublevel/autogen.go: is missing the license header
testdata/multilevel/sublevel/doc.go: is missing the license header
testdata/multilevel/sublevel/partial.go: is missing the license header
testdata/singlelevel/doc.go: is missing the license header
testdata/singlelevel/main.go: is missing the license header
testdata/singlelevel/wrapper.go: is missing the license header
testdata/singlelevel/zrapper.go: is missing the license header
testdata/x-pack/wrong.go: is missing the license header
`[1:],
		},
		{
			name: "Run a diff prints a list of files that need the Cloud license header",
			args: runParams{
				args:     []string{"testdata"},
				license:  "Cloud",
				licensor: defaultLicensor,
				ext:      defaultExt,
				dry:      true,
			},
			want: 1,
			err:  &Error{code: 1},
			wantOutput: `
testdata/cloud/wrong.go: is missing the license header
testdata/excludedpath/file.go: is missing the license header
testdata/multilevel/doc.go: is missing the license header
testdata/multilevel/main.go: is missing the license header
testdata/multilevel/sublevel/autogen.go: is missing the license header
testdata/multilevel/sublevel/doc.go: is missing the license header
testdata/multilevel/sublevel/partial.go: is missing the license header
testdata/singlelevel/doc.go: is missing the license header
testdata/singlelevel/main.go: is missing the license header
testdata/singlelevel/wrapper.go: is missing the license header
testdata/singlelevel/zrapper.go: is missing the license header
testdata/x-pack/doc.go: is missing the license header
testdata/x-pack/wrong.go: is missing the license header
`[1:],
		},
		{
			name: "Run against an unexisting dir fails",
			args: runParams{
				args:     []string{"ignore"},
				license:  defaultLicense,
				licensor: defaultLicensor,
				ext:      defaultExt,
				dry:      false,
			},
			want: 2,
			err:  goosPathError(2, "ignore"),
		},
		{
			name: "Unknown license fails",
			args: runParams{
				args:     []string{"ignore"},
				license:  "foo",
				licensor: defaultLicensor,
				ext:      defaultExt,
				dry:      false,
			},
			want: 7,
			err:  &Error{err: errors.New("unknown license: foo"), code: 7},
		},
		{
			name: "Run with default mode rewrites the source files and notice",
			args: runParams{
				args:          []string{"testdata"},
				license:       defaultLicense,
				licensor:      defaultLicensor,
				exclude:       []string{"excludedpath", "x-pack", "cloud"},
				ext:           defaultExt,
				dry:           false,
				notice:        true,
				noticeProject: "SomeProject",
				analyseFunc:   goModAnalyseFunc,
			},
			want:       0,
			wantOutput: "Generating NOTICE file...\n\n",
			wantGolden: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.args.args[0] != "ignore" {
				defer copyFixtures(t, tt.args.args[0])()
			}

			var buf = new(bytes.Buffer)
			tt.args.out = buf
			var err = run(tt.args)
			if !reflect.DeepEqual(err, tt.err) {
				t.Errorf("run() error = %v, wantErr %v", err, tt.err)
				return
			}

			var got = Code(err)
			if got != tt.want {
				t.Errorf("run() = %v, want %v", got, tt.want)
			}

			var gotOutput = buf.String()
			if so := strings.Split(tt.wantOutput, "Dumping"); len(so) > 1 {
				tt.wantOutput = filepath.FromSlash(so[0]) + "Dumping" + so[1]
			} else {
				tt.wantOutput = filepath.FromSlash(tt.wantOutput)
			}

			if gotOutput != tt.wantOutput {
				t.Errorf("Output = \n%v\n want \n%v", gotOutput, tt.wantOutput)
			}

			if tt.wantGolden {
				if *update {
					copyFixtures(t, "golden")
					params := tt.args
					params.args = []string{"golden"}
					if err := run(params); err != nil {
						t.Fatal(err)
					}
				}
				hashDirectories(t, "testdata", "golden")
			}
		})
	}
}

func hashDirectories(t *testing.T, src, dest string) {
	var srcHash = sha1.New()
	var dstHash = sha1.New()
	t.Logf("===== Walking %s =====\n", src)
	if err := filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil || path == src {
			return nil
		}

		t.Log(fmt.Sprint(info.Name(), " => ", info.Size()))
		io.WriteString(srcHash, fmt.Sprint(info.Name(), info.Size()))
		return nil
	}); err != nil {
		t.Fatal(err)
	}

	t.Logf("===== Walking %s =====\n", dest)
	if err := filepath.Walk(dest, func(path string, info os.FileInfo, err error) error {
		if err != nil || path == dest {
			return nil
		}

		t.Log(fmt.Sprint(info.Name(), " => ", info.Size()))
		io.WriteString(dstHash, fmt.Sprint(info.Name(), info.Size()))
		return nil
	}); err != nil {
		t.Fatal(err)
	}

	t.Log("===========================")
	var srcSum, dstSum = srcHash.Sum(nil), dstHash.Sum(nil)
	if bytes.Compare(srcSum, dstSum) > 0 {
		t.Errorf("Contents of %s are not the same as %s", src, dest)
		t.Errorf("src folder hash: %x", srcSum)
		t.Errorf("dst folder hash: %x", dstSum)
	}
}

func goosPathError(code int, p string) error {
	var opName = "stat"
	if runtime.GOOS == "windows" {
		opName = "CreateFile"
	}

	return &Error{code: code, err: &os.PathError{
		Op:   opName,
		Path: p,
		Err:  syscall.ENOENT,
	}}
}
