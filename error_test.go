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
	"errors"
	"testing"
)

func TestCode(t *testing.T) {
	type args struct {
		e error
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{
			name: "nil error returns 0",
			args: args{},
			want: 0,
		},
		{
			name: "standard error returns 255",
			args: args{
				e: errors.New("an error"),
			},
			want: 255,
		},
		{
			name: "Error error returns 1",
			args: args{
				e: &Error{
					err:  errors.New("an error"),
					code: 1,
				},
			},
			want: 1,
		},
		{
			name: "Error error returns 2",
			args: args{
				e: &Error{
					err:  errors.New("an error"),
					code: 2,
				},
			},
			want: 2,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Code(tt.args.e); got != tt.want {
				t.Errorf("Code() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestError_Error(t *testing.T) {
	type fields struct {
		err  error
		code int
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name:   "Empty Err, returns <nil>",
			fields: fields{},
			want:   "<nil>",
		},
		{
			name:   "Empty Err, returns <nil>",
			fields: fields{code: 1},
			want:   "<nil>",
		},
		{
			name:   "Empty Err, returns <nil>",
			fields: fields{code: 1, err: errors.New("an error")},
			want:   "an error",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := Error{
				err:  tt.fields.err,
				code: tt.fields.code,
			}
			if got := e.Error(); got != tt.want {
				t.Errorf("Error.Error() = %v, want %v", got, tt.want)
			}
		})
	}
}
