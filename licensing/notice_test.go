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

package licensing

import (
	"bytes"
	"errors"
	"io/ioutil"
	"os"
	"reflect"
	"testing"

	"github.com/hashicorp/go-multierror"
	"gopkg.in/src-d/go-license-detector.v2/licensedb"
)

func genAnalyseFunc(r []licensedb.Result) func(args ...string) []licensedb.Result {
	return func(args ...string) []licensedb.Result { return r }
}

func createGoModFile(path string, contents string) func() {
	ioutil.WriteFile(path, []byte(contents), 0777)

	return func() { os.RemoveAll(path) }
}

const fooCompanyOut = `somedeps
Copyright 2012-2019 FooCompany L.T.D.

This product includes software developed at FooCompany L.T.D. and
third-party software developed by the licenses listed below.

=========================================================================

gopkg.in/src-d/go-license-detector.v2    Apache-2.0
github.com/hashicorp/multierror-go       MPL-2.0

=========================================================================
`

func TestGenerateNotice(t *testing.T) {
	type args struct {
		params     GenerateNoticeParams
		withBuffer bool
	}
	tests := []struct {
		name    string
		args    args
		want    Notice
		wantOut string
		err     error
	}{
		{
			name: "parses go.mod with two dependencies and returns the findings",
			args: args{params: GenerateNoticeParams{
				GoModFile: "testfiles/twodeps.mod",
				AnalyseFunc: genAnalyseFunc([]licensedb.Result{
					{
						Arg:     "github.com/hashicorp/multierror-go",
						Matches: []licensedb.Match{{License: "MPL-2.0"}},
					},
					{
						Arg:     "gopkg.in/src-d/go-license-detector.v2",
						Matches: []licensedb.Match{{License: "Apache-2.0"}},
					},
				}),
				Project: "somedeps",
			}},
			want: Notice{Project: "somedeps", ProjectYears: "2019", Dependencies: []Dependency{
				{Name: "gopkg.in/src-d/go-license-detector.v2", License: "Apache-2.0"},
				{Name: "github.com/hashicorp/multierror-go", License: "MPL-2.0"},
			}},
		},
		{
			name: "parses go.mod with two dependencies with a writer",
			args: args{withBuffer: true, params: GenerateNoticeParams{
				Licensor:  "FooCompany L.T.D.",
				StartYear: 2012,
				GoModFile: "testfiles/twodeps.mod",
				AnalyseFunc: genAnalyseFunc([]licensedb.Result{
					{
						Arg:     "github.com/hashicorp/multierror-go",
						Matches: []licensedb.Match{{License: "MPL-2.0"}},
					},
					{
						Arg:     "gopkg.in/src-d/go-license-detector.v2",
						Matches: []licensedb.Match{{License: "Apache-2.0"}},
					},
				}),
				Project: "somedeps",
			}},
			want: Notice{Project: "somedeps", ProjectYears: "2012-2019", Licensor: "FooCompany L.T.D.",
				Dependencies: []Dependency{
					{Name: "gopkg.in/src-d/go-license-detector.v2", License: "Apache-2.0"},
					{Name: "github.com/hashicorp/multierror-go", License: "MPL-2.0"},
				},
				DependencyBlob: "gopkg.in/src-d/go-license-detector.v2    Apache-2.0\ngithub.com/hashicorp/multierror-go       MPL-2.0\n",
			},
			wantOut: fooCompanyOut,
		},
		{
			name: "parses go.mod with two dependencies and one of them has no matches",
			args: args{params: GenerateNoticeParams{
				GoModFile: "testfiles/twodeps.mod",
				AnalyseFunc: genAnalyseFunc([]licensedb.Result{
					{
						Arg:    "github.com/hashicorp/multierror-go",
						ErrStr: "no license file was found",
					},
					{
						Arg:     "gopkg.in/src-d/go-license-detector.v2",
						Matches: []licensedb.Match{{License: "Apache-2.0"}},
					},
				}),
				Project: "somedeps",
			}},
			want: Notice{Project: "somedeps", ProjectYears: "2019", Dependencies: []Dependency{
				{Name: "gopkg.in/src-d/go-license-detector.v2", License: "Apache-2.0"},
				{Name: "github.com/hashicorp/multierror-go", License: "no license file was found"},
			}},
		},
		{
			name: "fails on invalid parameters",
			args: args{params: GenerateNoticeParams{}},
			err: &multierror.Error{Errors: []error{
				errors.New("notice: missing file path"),
				errors.New("notice: missing AnalyseFunc"),
				errors.New("notice: missing project name"),
			}},
		},
		{
			name: "fails parsing gomod when the format is wrong",
			args: args{params: GenerateNoticeParams{
				GoModFile:   "notice.go",
				AnalyseFunc: genAnalyseFunc(nil),
				Project:     "some",
			}},
			want: Notice{Project: "some", ProjectYears: "2019"},
			err:  errors.New("notice.go:41:42: unexpected newline in string"),
		},
		{
			name: "parses go.mod but has no dependencies",
			args: args{params: GenerateNoticeParams{
				GoModFile:   "testfiles/nodeps.mod",
				AnalyseFunc: genAnalyseFunc(nil),
				Project:     "nodeps",
			}},
			want: Notice{Project: "nodeps", ProjectYears: "2019"},
			err:  errors.New("modfile has no dependencies to generate notice"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf = new(bytes.Buffer)
			if tt.args.withBuffer {
				tt.args.params.Writer = buf
			}

			got, err := GenerateNotice(tt.args.params)
			if !reflect.DeepEqual(err, tt.err) {
				t.Errorf("GenerateNotice() error = %v, wantErr %v", err, tt.err)
				return
			}

			if buf.String() != tt.wantOut {
				t.Errorf("GenerateNotice() output = %v, wantOut %v", buf.String(), tt.wantOut)
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GenerateNotice() = %v, want %v", got, tt.want)
			}
		})
	}
}
