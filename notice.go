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
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/elastic/go-licenser/licensing"
)

func doNotice(path string, params runParams) error {
	// TODO: potentially use git to extract the date.
	var startDate, _ = time.Parse("2006", params.noticeYear)

	// When the noticeYear hasn't been specified, it tries to autodiscover
	// which year the project started by gathering the oldest modification
	// timestamp in files by walking the directory.
	if params.noticeYear == "" {
		startDate = time.Now()
		if err := filepath.Walk(path,
			func(p string, info os.FileInfo, err error) error {
				if err != nil {
					return err
				}
				if info.IsDir() || strings.Contains(p, ".git") {
					return nil
				}
				if fileMod := info.ModTime(); fileMod.Before(startDate) {
					startDate = fileMod
				}
				return nil
			}); err != nil {
			return err
		}
	}

	var noticeWriter = params.out
	var noticeMessage = "Dumping NOTICE to output...\n\n"
	if !params.dry {
		var noticeFilePath = filepath.Join(path, noticeFile)
		f, err := openTruncateFile(noticeFilePath)
		if err != nil {
			return &Error{err: err, code: errOpenFileFailed}
		}
		defer f.Close()
		noticeWriter = f
		noticeMessage = "Generating NOTICE file...\n\n"
	}
	fmt.Fprintf(params.out, noticeMessage)

	if params.noticeProject == "" {
		absPath, _ := filepath.Abs(path)
		params.noticeProject = filepath.Base(absPath)
	}
	if _, err := licensing.GenerateNotice(licensing.GenerateNoticeParams{
		GoModFile:    filepath.Join(path, "go.mod"),
		Writer:       noticeWriter,
		Project:      params.noticeProject,
		Licensor:     licensor,
		StartYear:    startDate.Year(),
		NoticeHeader: params.noticeHeader,
		AnalyseFunc:  params.analyseFunc,
	}); err != nil {
		return &Error{err: err, code: errGenerateNoticeFailed}
	}

	return nil
}
