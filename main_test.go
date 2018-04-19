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
	"flag"
	"fmt"
	"go/build"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"
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

func Test_run(t *testing.T) {
	type args struct {
		args     []string
		exclDirs []string
		ext      string
		dry      bool
	}
	tests := []struct {
		name       string
		args       args
		want       int
		wantErr    bool
		wantOutput string
		wantGolden bool
	}{
		{
			name: "Run a diff prints a list of files that need the license header",
			args: args{
				args:     []string{"testdata"},
				exclDirs: defaultExludedDirs,
				ext:      defaultExt,
				dry:      true,
			},
			want: 1,
			wantOutput: `
src/github.com/elastic/go-licenser/testdata/multilevel/doc.go: is missing the license header
src/github.com/elastic/go-licenser/testdata/multilevel/main.go: is missing the license header
src/github.com/elastic/go-licenser/testdata/multilevel/sublevel/autogen.go: is missing the license header
src/github.com/elastic/go-licenser/testdata/multilevel/sublevel/doc.go: is missing the license header
src/github.com/elastic/go-licenser/testdata/multilevel/sublevel/partial.go: is missing the license header
src/github.com/elastic/go-licenser/testdata/singlelevel/doc.go: is missing the license header
src/github.com/elastic/go-licenser/testdata/singlelevel/main.go: is missing the license header
src/github.com/elastic/go-licenser/testdata/singlelevel/wrapper.go: is missing the license header
`[1:],
		},
		{
			name: "Run against an unexisting dir fails",
			args: args{
				args:     []string{"ignore"},
				exclDirs: defaultExludedDirs,
				ext:      defaultExt,
				dry:      false,
			},
			want:    4,
			wantErr: true,
		},
		{
			name: "Run with default mode rewrites the source files",
			args: args{
				args:     []string{"testdata"},
				exclDirs: defaultExludedDirs,
				ext:      defaultExt,
				dry:      false,
			},
			want:       0,
			wantErr:    false,
			wantGolden: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.args.args[0] != "ignore" {
				defer copyFixtures(t, tt.args.args[0])()
			}

			var buf = new(bytes.Buffer)
			got, err := run(tt.args.args, tt.args.exclDirs, tt.args.ext, tt.args.dry, buf)
			if (err != nil) != tt.wantErr {
				t.Errorf("run() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("run() = %v, want %v", got, tt.want)
			}

			gopath := os.Getenv("GOPATH")
			if gopath == "" {
				gopath = build.Default.GOPATH
			}
			gotOutput := strings.Replace(buf.String(), gopath, "", -1)
			gotOutput = strings.Replace(gotOutput, "../", "", -1)
			if gotOutput != tt.wantOutput {
				t.Errorf("Output = \n%v\n want \n%v", gotOutput, tt.wantOutput)
			}

			if tt.wantGolden {
				if *update {
					copyFixtures(t, "golden")
					if _, err := run([]string{"golden"}, tt.args.exclDirs, tt.args.ext, tt.args.dry, buf); err != nil {
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
