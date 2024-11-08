/*
** Copyright (C) 2001-2024 Zabbix SIA
**
** Permission is hereby granted, free of charge, to any person obtaining a copy of this software and associated
** documentation files (the "Software"), to deal in the Software without restriction, including without limitation the
** rights to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of the Software, and to
** permit persons to whom the Software is furnished to do so, subject to the following conditions:
**
** The above copyright notice and this permission notice shall be included in all copies or substantial portions
** of the Software.
**
** THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE
** WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR
** COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT,
** TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
** SOFTWARE.
**/

package handlers

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"golang.zabbix.com/plugin/nvidia/pkg/nvml"
	nvmlmock "golang.zabbix.com/plugin/nvidia/pkg/nvml-mock"
	"golang.zabbix.com/sdk/errs"
)

// func TestHandler_GoEnvironment(t *testing.T) {
// 	t.Parallel()

// 	type args struct {
// 		extraParams []string
// 	}

// 	tests := []struct {
// 		name    string
// 		args    args
// 		want    any
// 		wantErr bool
// 	}{
// 		{
// 			"+valid",
// 			args{
// 				[]string{"foo"},
// 			},
// 			map[string]string{"foo": "bar"},
// 			false,
// 		},
// 		{
// 			"+multiple",
// 			args{[]string{"foo", "abc"}},
// 			map[string]string{
// 				"foo": "bar",
// 				"abc": "def",
// 			},
// 			false,
// 		},
// 		{
// 			"+all",
// 			args{nil},
// 			map[string]string{
// 				"abc": "def",
// 				"bar": "foo",
// 				"foo": "bar",
// 			},
// 			false,
// 		},
// 		{
// 			"-notFound",
// 			args{[]string{"test"}},
// 			nil,
// 			true,
// 		},
// 	}
// 	for _, tt := range tests {
// 		tt := tt
// 		t.Run(tt.name, func(t *testing.T) {
// 			t.Parallel()

// 			h := &Handler{
// 				sysCalls: sysCallsMock{},
// 			}

// 			got, err := h.GoEnvironment(context.Background(), nil, tt.args.extraParams...)
// 			if (err != nil) != tt.wantErr {
// 				t.Fatalf("Handler.GoEnvironment() error = %v, wantErr %v", err, tt.wantErr)
// 			}

// 			if diff := cmp.Diff(tt.want, got); diff != "" {
// 				t.Fatalf("Handler.GoEnvironment() = %s", diff)
// 			}
// 		})
// 	}
// }

// func TestNew(t *testing.T) {
// 	t.Parallel()

// 	tests := []struct {
// 		name string
// 		want *Handler
// 	}{
// 		{
// 			"+valid",
// 			&Handler{
// 				client:   http.DefaultClient,
// 				sysCalls: osWrapper{},
// 			},
// 		},
// 	}
// 	for _, tt := range tests {
// 		tt := tt
// 		t.Run(tt.name, func(t *testing.T) {
// 			t.Parallel()

// 			got := New()
// 			if diff := cmp.Diff(tt.want, got, cmp.AllowUnexported(Handler{})); diff != "" {
// 				t.Fatalf("New() = %s", diff)
// 			}
// 		})
// 	}
// }

func TestWithJSONResponse(t *testing.T) {
	t.Parallel()

	type args struct {
		value  any
		gotErr bool
	}

	tests := []struct {
		name    string
		args    args
		want    any
		wantErr bool
	}{
		{
			"+valid",
			args{
				value: "foobar",
			},
			`"foobar"`,
			false,
		},
		{
			"+jsonObject",
			args{
				value: map[string]string{
					"foo":  "bar",
					"test": "true",
				},
			},
			`{"foo":"bar","test":"true"}`,
			false,
		},
		{
			"-jsonMarshalErr",
			args{
				value: map[struct{ test string }]string{
					{test: "1"}: "bar",
					{test: "2"}: "foo",
				},
			},
			nil,
			true,
		},
		{
			"-handlerErr",
			args{gotErr: true},
			nil,
			true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := WithJSONResponse(
				func(ctx context.Context, metricParams map[string]string, extraParams ...string) (any, error) {
					if tt.args.gotErr {
						return nil, errs.New("fail")
					}

					return tt.args.value, nil
				},
			)(context.Background(), nil)

			if (err != nil) != tt.wantErr {
				t.Fatalf("WithJSONResponse() error = %v, wantErr %v", err, tt.wantErr)
			}

			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Fatalf("WithJSONResponse() = %s", diff)
			}
		})
	}
}

func Test_getAll(t *testing.T) {
	t.Parallel()

	type args struct {
		env []string
	}

	tests := []struct {
		name    string
		args    args
		want    map[string]string
		wantErr bool
	}{
		{
			"+valid",
			args{[]string{"foo=bar", "abc=def"}},
			map[string]string{"foo": "bar", "abc": "def"},
			false,
		},
		{
			"+single",
			args{[]string{"foo=bar"}},
			map[string]string{"foo": "bar"},
			false,
		},
		{
			"-invalidVar",
			args{[]string{"foo:bar"}},
			nil,
			true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := getAll(tt.args.env)
			if (err != nil) != tt.wantErr {
				t.Fatalf("getAll() error = %v, wantErr %v", err, tt.wantErr)
			}

			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Fatalf("getAll() = %s", diff)
			}
		})
	}
}

func TestHandler_DriverVersion(t *testing.T) {
	t.Parallel()

	type fields struct {
		nvmlRunner nvml.Runner
	}

	tests := []struct {
		name    string
		fields  fields
		want    any
		wantErr bool
	}{
		{
			name: "+valid",
			fields: fields{
				&nvmlmock.MockRunner{
					DriverVersion: "Mock Driver Version",
				},
			},
			want:    "Mock Driver Version",
			wantErr: false,
		},
		{
			name: "-invalid",
			fields: fields{
				&nvmlmock.MockRunner{
					DriverVersion: "",
					WantedErr:     nvml.ErrNotFound,
				},
			},
			want:    "",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			h := &Handler{
				nvmlRunner: tt.fields.nvmlRunner,
			}

			got, err := h.DriverVersion(context.TODO(), nil, nil...)
			if (err != nil) != tt.wantErr {
				t.Fatalf("Handler.DriverVersion() error = %v, wantErr %v", err, tt.wantErr)
			}

			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Fatalf("Handler.DriverVersion() = %s", diff)
			}
		})
	}
}
