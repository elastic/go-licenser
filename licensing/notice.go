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
	"fmt"
	"go/build"
	"io"
	"io/ioutil"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"text/tabwriter"
	"text/template"
	"time"

	"github.com/hashicorp/go-multierror"
	"github.com/sirkon/goproxy/gomod"
	"gopkg.in/src-d/go-license-detector.v2/licensedb"
)

// DefaultNoticeHeader is used as the default NOTICE header.
const DefaultNoticeHeader = `{{.Project}}
Copyright {{.ProjectYears}} {{.Licensor}}

This product includes software developed at {{.Licensor}} and
third-party software developed by the licenses listed below.
`

// NoticeFormat is the template format for the generated notice file.
const NoticeFormat = `%s
=========================================================================

{{.DependencyBlob}}
=========================================================================
`

// Notice contains all of the information
type Notice struct {
	// Licensor is the source code's owner / maintainer.
	Licensor string

	// Project is the source code's project or repository name.
	Project string

	// ProjectYears contains the years that the project has been running
	// in the {{start year}}-{{current year}} or {{year}} format depending
	// if it's a single or multi-year effort.
	ProjectYears string

	// Dependencies contains the dependency list that can be consumed in a
	// programatic way.
	Dependencies []Dependency

	// DependencyBlob is a stringified version of Dependencies that has been
	// processed through the tabwritter package so the columns are aligned.
	// Only populated when Writer is not nil.
	DependencyBlob string
}

// Dependency contains the dependency name and the license.
type Dependency struct {
	Name, License string
}

// GenerateNoticeParams is consumed by GenerateNotice.
type GenerateNoticeParams struct {
	// GoModFile is the location of the `go.mod` file to open.
	GoModFile string

	// Licensor of the codebase.
	Licensor string

	// Project name.
	Project string

	// Project's start year so it can be reflected in the NOTICE file.
	StartYear int

	// Writer where to write the NOTICE templated output. If not specified,
	// writing the template will be skipped, and Notice won't have a populated
	// DependencyBlob.
	Writer io.Writer

	// NoticeHeader overrides the DefaultNoticeHeader used in NOTICE. Only relevant
	// when Writer is not nil.
	NoticeHeader string

	// AnalyseFunc returns a list of results with their license.
	AnalyseFunc func(args ...string) []licensedb.Result
}

// Validate ensures that the structure is usable by its consumer.
func (params GenerateNoticeParams) Validate() error {
	var merr = new(multierror.Error)
	if params.GoModFile == "" {
		merr = multierror.Append(merr, errors.New("notice: missing file path"))
	}
	if params.AnalyseFunc == nil {
		merr = multierror.Append(merr, errors.New("notice: missing AnalyseFunc"))
	}
	if params.Project == "" {
		merr = multierror.Append(merr, errors.New("notice: missing project name"))
	}

	return merr.ErrorOrNil()
}

func (params GenerateNoticeParams) fillDefaults() GenerateNoticeParams {
	if params.NoticeHeader == "" {
		params.NoticeHeader = DefaultNoticeHeader
	}
	return params
}

func goPkgPath() string {
	return filepath.Join(build.Default.GOPATH, "pkg", "mod")
}

// GenerateNotice inspects the `go.mod` file and inspects the contents of the
// downloaded dependency to determine which license is most predominant in its
// source files. When a Writer is passed in the parameters, it also writes a
// templated output in the writer.
func GenerateNotice(params GenerateNoticeParams) (Notice, error) {
	var notice Notice
	if err := params.Validate(); err != nil {
		return notice, err
	}

	params = params.fillDefaults()
	notice = buildNotice(params)

	paths, err := getModulePaths(params.GoModFile)
	if err != nil {
		return notice, err
	}

	notice.Dependencies = getLicenses(params.AnalyseFunc, paths...)

	// Skip template execution.
	if params.Writer == nil {
		return notice, nil
	}

	templateFormat := fmt.Sprintf(NoticeFormat, params.NoticeHeader)
	if err := writeTemplate(&notice, templateFormat, params.Writer); err != nil {
		return notice, err
	}
	return notice, nil
}

// buildNotice constructs a Notice from the parameters.
func buildNotice(params GenerateNoticeParams) (notice Notice) {
	var currentYear = time.Now().Year()
	notice = Notice{
		Licensor:     params.Licensor,
		Project:      params.Project,
		ProjectYears: fmt.Sprint(params.StartYear, "-", currentYear),
	}
	if params.StartYear == currentYear || params.StartYear == 0 {
		notice.ProjectYears = fmt.Sprint(currentYear)
	}
	return notice
}

// getModulePaths opens and parses a `go.mod` file obtaining the list of
// dependencies and constructing a local filesystem path where they have
// been downloaded. The returning slice is sorted.
func getModulePaths(file string) ([]string, error) {
	contents, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}

	module, err := gomod.Parse(file, contents)
	if err != nil {
		return nil, err
	}

	if len(module.Require) == 0 {
		return nil, errors.New("modfile has no dependencies to generate notice")
	}

	var paths = make([]string, 0, len(module.Require))
	// Each dependency is processed and the full filepath is constructed.
	// Additionally, dependencies that have Uppercase characters are converted
	// to !<lowercase> since it's how go modules are downloaded in the local
	// filesystem.
	for dep, version := range module.Require {
		re, err := regexp.Compile("([A-Z]+)")
		if err != nil {
			return nil, err
		}

		for _, letter := range re.FindAllString(dep, -1) {
			dep = strings.Replace(dep, letter,
				"!"+strings.ToLower(letter), -1,
			)
		}

		p := filepath.Join(
			append([]string{goPkgPath()}, strings.Split(dep, "/")...)...,
		)
		p = fmt.Sprint(p, "@", version)
		paths = append(paths, p)
	}

	sort.Strings(paths)
	return paths, nil
}

// getLicenses returns a []Dependency which is populated by the analyseFunc.
// It relies on the local downloaded dependencies in order for the license
// discovery mechanism to work. The Dependencies slice is sorted by License.
func getLicenses(analyseFunc func(args ...string) []licensedb.Result, paths ...string) []Dependency {
	var dependencies = make([]Dependency, 0, len(paths))
	for _, res := range analyseFunc(paths...) {
		// Replaces the prefix in front of the Go dependency name and the
		// @<version> that follows, so that a clean name is returned.
		arg := strings.Replace(res.Arg, goPkgPath()+"/", "", -1)
		if depv := strings.Split(arg, "@"); len(depv) > 1 {
			arg = strings.Replace(arg, "@"+depv[1], "", -1)
		}

		// When the license type hasn't been detected, the "Error" is treated
		// as the License so it can be reflected.
		if len(res.Matches) == 0 {
			// Removes the break line at the end of the error string and
			// the local goPath if it's there.
			res.Matches = []licensedb.Match{{
				License: strings.Replace(
					strings.Replace(res.ErrStr, "\n", "", 1),
					goPkgPath()+"/", "", -1,
				),
			}}
		}

		// The ! character is removed for any uppercase dependencies and the
		// dependency on the NOTICE will have theose characters lowercased.
		dependencies = append(dependencies, Dependency{
			Name:    strings.Replace(arg, "!", "", -1),
			License: res.Matches[0].License,
		})
	}

	sort.SliceStable(dependencies, func(i, j int) bool {
		return strings.ToLower(dependencies[i].License) < strings.ToLower(dependencies[j].License)
	})

	return dependencies
}

// writeTemplate writes a notice templated output by using the specified format
// and on to the writer. The DependenciesBlob field on the passed Notice will
// be populated as:
//	{{.Dependency}}    {{.License}}
// The tab space will be determined by the length of the dependency name.
func writeTemplate(notice *Notice, format string, writer io.Writer) error {
	var buf = new(bytes.Buffer)
	var w = tabwriter.NewWriter(buf, 4, 2, 4, ' ', 0)
	var deps = notice.Dependencies
	for i := range deps {
		w.Write([]byte(deps[i].Name + "\t" + deps[i].License + "\n"))
	}

	w.Flush()
	notice.DependencyBlob = buf.String()

	t, err := template.New("").Parse(format)
	if err != nil {
		return err
	}

	return t.Execute(writer, notice)
}
