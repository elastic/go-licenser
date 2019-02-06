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
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestContainsHeader(t *testing.T) {
	exampleHeader := []string{
		fmt.Sprintf("// Copyright %d The elastic/go-licenser Authors. All rights reserved.", time.Now().Year()),
	}

	type args struct {
		r      io.Reader
		header []string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "ContainsHeader returns false empty line reader",
			args: args{
				r:      strings.NewReader(""),
				header: exampleHeader,
			},
			want: false,
		},
		{
			name: "ContainsHeader returns false on single line reader",
			args: args{
				r: strings.NewReader(
					"package main\n",
				),
				header: exampleHeader,
			},
			want: false,
		},
		{
			name: "ContainsHeader returns false on single line reader",
			args: args{
				r: strings.NewReader(
					"a reader has just one line\n",
				),
				header: exampleHeader,
			},
			want: false,
		},
		{
			name: "ContainsHeader returns false on two line reader",
			args: args{
				r: strings.NewReader(
					fmt.Sprintln(
						"a reader has just one line\na reader has two lines",
					),
				),
				header: exampleHeader,
			},
			want: false,
		},
		{
			name: "ContainsHeader returns false on three line reader",
			args: args{
				r: strings.NewReader(
					fmt.Sprintln(
						"a reader has just one line\na reader has two lines\na reader has three lines",
					),
				),
				header: exampleHeader,
			},
			want: false,
		},
		{
			name: "ContainsHeader returns false on reader that doesn't match the license year",
			args: args{
				r: strings.NewReader(
					fmt.Sprint(`
// Copyright 2017 The elastic/go-licenser Authors. All rights reserved.
// Use of this source code is governed by Apache License 2.0 that can
// be found in the LICENSE file.
`[1:],
					),
				),
				header: exampleHeader,
			},
			want: false,
		},
		{
			name: "ContainsHeader returns true on reader that matches the license text",
			args: args{
				r: strings.NewReader(
					fmt.Sprintf(`
// Copyright %d The elastic/go-licenser Authors. All rights reserved.

package main
`[1:], time.Now().Year(),
					),
				),
				header: exampleHeader,
			},
			want: true,
		},
		{
			name: "ContainsHeader returns true on reader that matches the ASL license text",
			args: args{
				r: strings.NewReader(`
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

package mypackage
`[1:],
				),
				header: []string{
					`// Licensed to Elasticsearch B.V. under one or more contributor`,
					`// license agreements. See the NOTICE file distributed with`,
					`// this work for additional information regarding copyright`,
					`// ownership. Elasticsearch B.V. licenses this file to you under`,
					`// the Apache License, Version 2.0 (the "License"); you may`,
					`// not use this file except in compliance with the License.`,
					`// You may obtain a copy of the License at`,
					"//",
					"//     http://www.apache.org/licenses/LICENSE-2.0",
					"//",
					`// Unless required by applicable law or agreed to in writing,`,
					`// software distributed under the License is distributed on an`,
					`// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY`,
					`// KIND, either express or implied.  See the License for the`,
					`// specific language governing permissions and limitations`,
					`// under the License.`,
				},
			},
			want: true,
		},
		{
			name: "ContainsHeader returns true on reader that matches the ASL license text after package",
			args: args{
				r: strings.NewReader(`
package mypackage

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
// under the License."

func somefunc() {}
`[1:],
				),
				header: []string{
					`// Licensed to Elasticsearch B.V. under one or more contributor`,
					`// license agreements. See the NOTICE file distributed with`,
					`// this work for additional information regarding copyright`,
					`// ownership. Elasticsearch B.V. licenses this file to you under`,
					`// the Apache License, Version 2.0 (the "License"); you may`,
					`// not use this file except in compliance with the License.`,
					`// You may obtain a copy of the License at`,
					"//",
					"//     http://www.apache.org/licenses/LICENSE-2.0",
					"//",
					`// Unless required by applicable law or agreed to in writing,`,
					`// software distributed under the License is distributed on an`,
					`// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY`,
					`// KIND, either express or implied.  See the License for the`,
					`// specific language governing permissions and limitations`,
					`// under the License.`,
				},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ContainsHeader(tt.args.r, tt.args.header); got != tt.want {
				t.Errorf("ContainsHeader() = %v, want %v", got, tt.want)
			}
		})
	}
}

func CreateFileObtainName(t *testing.T) (string, func()) {
	f, err := ioutil.TempFile(os.TempDir(), "TestHelperCreateFile")
	if err != nil {
		t.Error("failed creating temp file", err)
	}
	defer f.Close()
	return f.Name(), func() { os.Remove(f.Name()) }
}

func writeContents(t *testing.T, path string, content []byte) {
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}

	if err := ioutil.WriteFile(path, content, info.Mode()); err != nil {
		t.Fatal(err)
	}
}

func TestRewriteFileWithHeader(t *testing.T) {
	simple, cleanup := CreateFileObtainName(t)
	defer cleanup()

	complex, cleanup := CreateFileObtainName(t)
	defer cleanup()

	licenseWeirdLocation, cleanup := CreateFileObtainName(t)
	defer cleanup()

	writeContents(t, complex, []byte(`
// something something

// Package main is the main drill
package main
`[1:]))

	writeContents(t, licenseWeirdLocation, []byte(`
// Package main is the main drill
package main

// Licensed to Elasticsearch B.V. under one or more contributor
// license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright
// ownership. Elasticsearch B.V. licenses this file to you under
// the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.
// You may obtain a copy of the License at
`[1:]))

	type args struct {
		path   string
		header []byte
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
		want    []byte
	}{
		{
			name: "Rewrite succeeds on an empty file",
			args: args{
				path:   simple,
				header: []byte("// This is the header I want to see"),
			},
			wantErr: false,
			want:    []byte("// This is the header I want to see\n\n"),
		},
		{
			name: "Rewrite succeeds with a file that has some contents",
			args: args{
				path: complex,
				header: []byte(`
// Copyright 2018 The elastic/go-licenser Authors. All rights reserved.
// Use of this source code is governed by Apache License 2.0 that can
// be found in the LICENSE file.
`[1:]),
			},
			wantErr: false,
			want: []byte(`
// Copyright 2018 The elastic/go-licenser Authors. All rights reserved.
// Use of this source code is governed by Apache License 2.0 that can
// be found in the LICENSE file.

// something something

// Package main is the main drill
package main
`[1:]),
		},
		{
			name: "Rewrite succeeds with a file that has some contents",
			args: args{
				path: licenseWeirdLocation,
				header: []byte(`
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
`[1:]),
			},
			wantErr: false,
			want: []byte(`
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

// Package main is the main drill
package main

// Licensed to Elasticsearch B.V. under one or more contributor
// license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright
// ownership. Elasticsearch B.V. licenses this file to you under
// the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.
// You may obtain a copy of the License at
`[1:]),
		},
		{
			name: "Rewrite fails when the file doesn't exist",
			args: args{
				path:   "unexisting_file",
				header: nil,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := RewriteFileWithHeader(tt.args.path, tt.args.header); (err != nil) != tt.wantErr {
				t.Errorf("RewriteFileWithHeader() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.want != nil {
				got, err := ioutil.ReadFile(tt.args.path)
				if err != nil {
					t.Error("failed reading contents of temp file")
				}
				if !reflect.DeepEqual(got, tt.want) {
					t.Errorf("RewriteFileWithHeader() = \n%v\n, want \n%v\n", string(got), string(tt.want))
				}
			}
		})
	}
}

func TestContainsHeaderLine(t *testing.T) {
	type args struct {
		r           io.Reader
		headerLines []string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "ContainsHeader returns true on reader that matches some the lines",
			args: args{
				r: strings.NewReader(
					fmt.Sprint(`
// Copyright 2017 The elastic/go-licenser Authors. All rights reserved.
// Use of this source code is governed by Apache License 2.0 that can
// be found in the LICENSE file.
`[1:],
					),
				),
				headerLines: []string{
					"elastic/go-licenser Authors. All rights reserved.",
					"// Use of this source code is governed by Apache License 2.0 that can",
					"// be found in the LICENSE file.",
				},
			},
			want: true,
		},
		{
			name: "ContainsHeader returns false on reader that matches some the lines",
			args: args{
				r: strings.NewReader(
					fmt.Sprint(`
// Package mypackage this package does things.
`[1:],
					),
				),
				headerLines: []string{
					"elastic/go-licenser Authors. All rights reserved.",
					"// Use of this source code is governed by Apache License 2.0 that can",
					"// be found in the LICENSE file.",
				},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := containsHeaderLine(tt.args.r, tt.args.headerLines); got != tt.want {
				t.Errorf("containsHeaderLine() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_headerBytes(t *testing.T) {
	var singleHeader = `
// Copyright 2017 The elastic/go-licenser Authors. All rights reserved.
// Use of this source code is governed by Apache License 2.0 that can
// be found in the LICENSE file.

`[1:]
	type args struct {
		r io.Reader
	}
	tests := []struct {
		name string
		args args
		want []byte
	}{
		{
			name: "Simple header is detected",
			args: args{
				r: strings.NewReader(`
// Copyright 2017 The elastic/go-licenser Authors. All rights reserved.
// Use of this source code is governed by Apache License 2.0 that can
// be found in the LICENSE file.

package mypackage
`[1:],
				),
			},
			want: []byte(singleHeader),
		},
		{
			name: "Simple header with build tag is detected",
			args: args{
				r: strings.NewReader(`
// Copyright 2017 The elastic/go-licenser Authors. All rights reserved.
// Use of this source code is governed by Apache License 2.0 that can
// be found in the LICENSE file.

`[1:] + "// +build mybuildtag\n\n" + `
package mypackage
`[1:],
				),
			},
			want: []byte(singleHeader),
		},
		{
			name: "Simple header with package comments build tag is detected",
			args: args{
				r: strings.NewReader(`
// Copyright 2017 The elastic/go-licenser Authors. All rights reserved.
// Use of this source code is governed by Apache License 2.0 that can
// be found in the LICENSE file.

// Package mypackage does a lot of stuff, here's what it does
package mypackage
`[1:],
				),
			},
			want: []byte(singleHeader),
		},
		{
			name: "duplicated header header with package comments is detected",
			args: args{
				r: strings.NewReader(`
// Copyright 2018 The elastic/go-licenser Authors. All rights reserved.
// Use of this source code is governed by Apache License 2.0 that can
// be found in the LICENSE file.

// Copyright 2017 The elastic/go-licenser Authors. All rights reserved.
// Use of this source code is governed by Apache License 2.0 that can
// be found in the LICENSE file.

// Package mypackage does a lot of stuff, here's what it does
package mypackage
`[1:],
				),
			},
			want: []byte(`
// Copyright 2018 The elastic/go-licenser Authors. All rights reserved.
// Use of this source code is governed by Apache License 2.0 that can
// be found in the LICENSE file.

// Copyright 2017 The elastic/go-licenser Authors. All rights reserved.
// Use of this source code is governed by Apache License 2.0 that can
// be found in the LICENSE file.

`[1:],
			),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := headerBytes(tt.args.r); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("headerBytes() = \n%v\n, want \n%v\n", string(got), string(tt.want))
			}
		})
	}
}
