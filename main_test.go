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
	"errors"
	"flag"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

var update = flag.Bool("update", false, "updates the golden files with the latest iteration of the code")

func Test_run(t *testing.T) {
	type args struct {
		args     []string
		license  string
		licensor string
		exclude  []string
		ext      string
		dry      bool
	}
	tests := []struct {
		name       string
		args       args
		want       int
		err        error
		wantOutput string
		wantGolden bool
	}{
		{
			name: "Run a diff prints a list of files that need the license header",
			args: args{
				args:     []string{"testdata"},
				license:  defaultLicense,
				licensor: defaultLicensor,
				exclude:  []string{"excludedpath", "x-pack", "x-pack-v2", "cloud"},
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
			name: "Run a diff prints a list of files that need the Elastic license header",
			args: args{
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
testdata/x-pack-v2/correct.go: is missing the license header
testdata/x-pack-v2/wrong.go: is missing the license header
`[1:],
		},
		{
			name: "Run a diff prints a list of files that need the Elastic license 2.0 header",
			args: args{
				args:     []string{"testdata"},
				license:  "Elasticv2",
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
testdata/x-pack/correct.go: is missing the license header
testdata/x-pack/wrong.go: is missing the license header
testdata/x-pack-v2/wrong.go: is missing the license header
`[1:],
		},
		{
			name: "Run a diff prints a list of files that need the Cloud license header",
			args: args{
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
testdata/x-pack/correct.go: is missing the license header
testdata/x-pack/wrong.go: is missing the license header
testdata/x-pack-v2/correct.go: is missing the license header
testdata/x-pack-v2/wrong.go: is missing the license header
`[1:],
		},
		{
			name: "Run against an unexisting dir fails",
			args: args{
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
			args: args{
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
			name: "Check ASL2 license rewrite",
			args: args{
				args:     []string{"testdata"},
				license:  defaultLicense,
				licensor: defaultLicensor,
				exclude:  []string{"excludedpath"},
				ext:      defaultExt,
				dry:      false,
			},
			want:       0,
			wantGolden: true,
		},
		{
			name: "Check ASL2-short license rewrite",
			args: args{
				args:     []string{"testdata"},
				license:  "ASL2-Short",
				licensor: defaultLicensor,
				exclude:  []string{"excludedpath"},
				ext:      defaultExt,
				dry:      false,
			},
			want:       0,
			wantGolden: true,
		},
		{
			name: "Check Cloud license rewrite",
			args: args{
				args:     []string{"testdata"},
				license:  "Cloud",
				licensor: defaultLicensor,
				exclude:  []string{"excludedpath"},
				ext:      defaultExt,
				dry:      false,
			},
			want:       0,
			wantGolden: true,
		},
		{
			name: "Check Elastic license rewrite",
			args: args{
				args:     []string{"testdata"},
				license:  "Elastic",
				licensor: defaultLicensor,
				exclude:  []string{"excludedpath"},
				ext:      defaultExt,
				dry:      false,
			},
			want:       0,
			wantGolden: true,
		},
		{
			name: "Check Elastic 2.0 license rewrite",
			args: args{
				args:     []string{"testdata"},
				license:  "Elasticv2",
				licensor: defaultLicensor,
				exclude:  []string{"excludedpath"},
				ext:      defaultExt,
				dry:      false,
			},
			want:       0,
			wantGolden: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.args.args[0] != "ignore" {
				defer copyFixtures(t, tt.args.args[0])()
			}

			var buf = new(bytes.Buffer)
			var err = run(tt.args.args, tt.args.license, tt.args.licensor, tt.args.exclude, tt.args.ext, tt.args.dry, buf)
			if !reflect.DeepEqual(err, tt.err) {
				t.Errorf("run() error = %v, wantErr %v", err, tt.err)
				return
			}

			var got = Code(err)
			if got != tt.want {
				t.Errorf("run() = %v, want %v", got, tt.want)
			}

			var gotOutput = buf.String()
			tt.wantOutput = filepath.FromSlash(tt.wantOutput)
			if gotOutput != tt.wantOutput {
				t.Errorf("Output = \n%v\n want \n%v", gotOutput, tt.wantOutput)
			}

			if tt.wantGolden {
				goldenDirectory := filepath.Join("golden", tt.args.license)
				if *update {
					copyFixtures(t, goldenDirectory)
					if err := run([]string{goldenDirectory}, tt.args.license, tt.args.licensor, tt.args.exclude, tt.args.ext, tt.args.dry, buf); err != nil {
						t.Fatal(err)
					}
				}
				hashDirectories(t, "testdata", goldenDirectory)
			}
		})
	}
}

func BenchmarkRun(b *testing.B) {
	args := []string{"."}
	excluded := append(defaultExludedDirs, "golden")

	b.Run("dot", func(b *testing.B) {
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			run(args, defaultLicense, defaultLicensor, excluded, defaultExt, false, os.Stdout)
		}
	})
}
