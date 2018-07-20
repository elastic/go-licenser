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
	"path/filepath"
	"testing"
)

func Test_needsExclusion(t *testing.T) {
	type args struct {
		path    string
		exclude []string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "Path is excluded",
			args: args{
				path:    "apath/thatdoesNeed/exclusion",
				exclude: []string{"apath"},
			},
			want: true,
		},
		{
			name: "Path is not excluded",
			args: args{
				path:    "apath/thatdoesNOTNeed/exclusion",
				exclude: []string{"anotherpath"},
			},
			want: false,
		},
		{
			name: "Path is excluded",
			args: args{
				path:    "apath/thatdoesNeed/exclusion",
				exclude: []string{"apath/thatdoesNeed"},
			},
			want: true,
		},
		{
			name: "Path is excluded",
			args: args{
				path:    "apath/thatdoesNeed/exclusion",
				exclude: []string{"apath/thatdoesNeed/"},
			},
			want: true,
		},
		{
			name: "Path is excluded",
			args: args{
				path:    "apath/thatdoesNeed/exclusion",
				exclude: []string{"apath/thatdoesNeed/*"},
			},
			want: true,
		},
		{
			name: "Path is excluded",
			args: args{
				path:    "apath/thatdoesNeed/exclusion",
				exclude: []string{"apath/thatdoesNeed/exclusion"},
			},
			want: true,
		},
		{
			name: "Path is excluded",
			args: args{
				path:    filepath.Join("apath", "thatdoesNeed", "exclusion"),
				exclude: []string{filepath.Join("apath", "thatdoesNeed", "exclusion", "*")},
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := needsExclusion(tt.args.path, tt.args.exclude); got != tt.want {
				t.Errorf("needsExclusion() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_cleanPathSuffixes(t *testing.T) {
	type args struct {
		path    string
		sufixes []string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "Cleans the suffixes",
			args: args{
				path:    "apath/needsTheCLEANS/*",
				sufixes: []string{"*", "/"},
			},
			want: "apath/needsTheCLEANS",
		},
		{
			name: "Cleans multiple suffixes multiple times",
			args: args{
				path:    "apath/needsTheCLEANS////***",
				sufixes: []string{"*", "/"},
			},
			want: "apath/needsTheCLEANS",
		},
		{
			name: "Cleans the suffixes",
			args: args{
				path:    "apath/needsTheCLEANS/",
				sufixes: []string{"/"},
			},
			want: "apath/needsTheCLEANS",
		},
		{
			name: "Cleans a single suffix multiple times",
			args: args{
				path:    "apath/needsTheCLEANS/////",
				sufixes: []string{"/"},
			},
			want: "apath/needsTheCLEANS",
		},
		{
			name: "Cleans no suffixes if none are passed",
			args: args{
				path: "apath/needsTheCLEANS/",
			},
			want: "apath/needsTheCLEANS/",
		},
		{
			name: "empty string case",
			args: args{
				path: "",
			},
			want: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := cleanPathSuffixes(tt.args.path, tt.args.sufixes); got != tt.want {
				t.Errorf("cleanPathSuffixes() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_cleanPathPrefixes(t *testing.T) {
	type args struct {
		path     string
		prefixes []string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "Cleans path prefixes",
			args: args{
				path:     "/prefix",
				prefixes: []string{"/"},
			},
			want: "prefix",
		},
		{
			name: "Cleans no path prefixes",
			args: args{
				path:     "prefix",
				prefixes: []string{"/"},
			},
			want: "prefix",
		},
		{
			name: "Cleans path prefixes",
			args: args{
				path:     "xyzprefix",
				prefixes: []string{"xyz"},
			},
			want: "prefix",
		},
		{
			name: "Cleans no prefixes (Empty prefixes)",
			args: args{
				path: "something",
			},
			want: "something",
		},
		{
			name: "Cleans no prefixes (Empty path)",
			args: args{
				path:     "",
				prefixes: []string{"zyx"},
			},
			want: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := cleanPathPrefixes(tt.args.path, tt.args.prefixes); got != tt.want {
				t.Errorf("cleanPathPrefixes() = %v, want %v", got, tt.want)
			}
		})
	}
}
