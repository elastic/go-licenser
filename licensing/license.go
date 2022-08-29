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
	"bufio"
	"bytes"
	"errors"
	"io"
	"os"
	"reflect"
	"strings"
	"sync"
)

var (
	startPrefixes = []string{"// Copyright", "// copyright", "// Licensed", "// licensed", "// ELASTICSEARCH CONFIDENTIAL"}
	endPrefixes   = []string{"package ", "// Package ", "// +build ", "// Code generated", "// code generated", "//go:"}

	errHeaderIsTooShort = errors.New("header is too short")

	defaulBufSize int
	bufPool       = sync.Pool{
		New: func() interface{} {
			buf := make([]byte, defaulBufSize)
			return buf
		},
	}
)

func init() {
	// Iterate over the supported licenses to make sure everything fit
	// without any additional allocation.
	for _, v := range Headers {
		var l int
		for _, v2 := range v {
			l += len(v2)
		}

		if l > defaulBufSize {
			defaulBufSize = l
		}
	}
}

// ContainsHeader reads the first N lines of a file and checks if the header
// matches the one that is expected
func ContainsHeader(r io.Reader, headerLines []string) bool {
	var scanner = bufio.NewScanner(r)
	var i int

	buf := bufPool.Get().([]byte)
	defer bufPool.Put(buf)
	scanner.Buffer(buf, defaulBufSize)

	for i = 0; scanner.Scan(); i++ {
		line := scanner.Bytes()

		// end of license, break out of the loop
		if i == len(headerLines) {
			break
		}

		// compare line by line without storing the whole file
		// in memory
		if !bytes.Equal(line, []byte(headerLines[i])) {
			return false
		}
	}

	// file is shorter than license
	if i < len(headerLines) {
		return false
	}

	return true
}

// RewriteFileWithHeader reads a file from a path and rewrites it with a header
func RewriteFileWithHeader(path string, header []byte) error {
	if len(header) < 2 {
		return errHeaderIsTooShort
	}

	info, err := os.Stat(path)
	if err != nil {
		return err
	}

	origin, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	data := RewriteWithHeader(origin, header)
	return os.WriteFile(path, data, info.Mode())
}

// RewriteWithHeader rewrites the src byte buffers header with the new header.
func RewriteWithHeader(src []byte, header []byte) []byte {
	// Ensures that the header includes two break lines as the last bytes
	for !reflect.DeepEqual(header[len(header)-2:], []byte("\n\n")) {
		header = append(header, []byte("\n")...)
	}

	var oldHeader = headerBytes(bytes.NewReader(src))
	return bytes.Replace(src, oldHeader, header, 1)
}

// headerBytes detects the header lines of an io.Reader contents and returns
// what it considerst to be the header as a slice of bytes.
func headerBytes(r io.Reader) []byte {
	var scanner = bufio.NewScanner(r)
	var replaceableHeader []byte
	var continuedHeader bool
	for scanner.Scan() {
		var t = scanner.Text()

		for i := range endPrefixes {
			if strings.HasPrefix(t, endPrefixes[i]) {
				return replaceableHeader
			}
		}

		for i := range startPrefixes {
			if strings.HasPrefix(t, startPrefixes[i]) {
				continuedHeader = true
			}
		}

		if continuedHeader {
			replaceableHeader = append(replaceableHeader, []byte(t+"\n")...)
		}
	}

	return replaceableHeader
}

// containsHeaderLine reads the first N lines of a file and checks if the header
// matches the one that is expected
func containsHeaderLine(r io.Reader, headerLines []string) bool {
	var scanner = bufio.NewScanner(r)
	for scanner.Scan() {
		for i := range headerLines {
			if scanner.Text() == headerLines[i] {
				return true
			}
		}
	}

	return false
}
