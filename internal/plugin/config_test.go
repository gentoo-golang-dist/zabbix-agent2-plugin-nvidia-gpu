/*
** Zabbix
** Copyright (C) 2001-2026 Zabbix SIA
**
** Licensed under the Apache License, Version 2.0 (the "License");
** you may not use this file except in compliance with the License.
** You may obtain a copy of the License at
**
**     http://www.apache.org/licenses/LICENSE-2.0
**
** Unless required by applicable law or agreed to in writing, software
** distributed under the License is distributed on an "AS IS" BASIS,
** WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
** See the License for the specific language governing permissions and
** limitations under the License.
**/

package plugin

import (
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	nvmlmock "golang.zabbix.com/plugin/nvidia/pkg/nvml-mock"
	"golang.zabbix.com/sdk/errs"
	"golang.zabbix.com/sdk/log"
	"golang.zabbix.com/sdk/plugin"
)

func Test_nvmlPlugin_Configure(t *testing.T) {
	t.Parallel()

	type fields struct {
		config *pluginConfig
	}

	type args struct {
		global  *plugin.GlobalOptions
		options any
	}

	tests := []struct {
		name   string
		fields fields
		args   args
		want   *pluginConfig
	}{
		{
			"+valid",
			fields{},
			args{
				&plugin.GlobalOptions{Timeout: 3},
				[]byte(`Timeout=30`),
			},
			&pluginConfig{
				Timeout: 30,
			},
		},
		{
			"+prevConfig",
			fields{
				&pluginConfig{
					Timeout: 44,
				},
			},
			args{
				&plugin.GlobalOptions{Timeout: 3},
				[]byte(`Timeout=30`),
			},
			&pluginConfig{
				Timeout: 30,
			},
		},
		{
			"+globalOverwrite",
			fields{},
			args{
				&plugin.GlobalOptions{Timeout: 30},
				[]byte(``),
			},
			&pluginConfig{
				Timeout: 30,
			},
		},
		{
			"-unmarshalErr",
			fields{},
			args{
				&plugin.GlobalOptions{Timeout: 3},
				[]byte(
					strings.Join(
						[]string{"Timeout=2", "invalid"},
						"\n",
					),
				),
			},
			nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			p := &nvmlPlugin{
				config: tt.fields.config,
				Base: plugin.Base{
					Logger: log.New("test"),
				},
			}

			p.Configure(tt.args.global, tt.args.options)

			if diff := cmp.Diff(tt.want, p.config); diff != "" {
				t.Fatalf("nvmlPlugin_Configure() = %s", diff)
			}
		})
	}
}
func Test_nvmlPlugin_Validate(t *testing.T) {
	t.Parallel()

	type fields struct {
		failed       bool
		runnerExpect []*nvmlmock.Expectation
	}

	type args struct {
		options any
	}

	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			"+valid",
			fields{
				false,
				[]*nvmlmock.Expectation{
					nvmlmock.NewExpectation("Close").ProvideError(nil),
				},
			},
			args{
				[]byte(`Timeout=30`),
			},
			false,
		},
		{
			"-unmarshalErr",
			fields{
				false,
				[]*nvmlmock.Expectation{
					nvmlmock.NewExpectation("Close").ProvideError(nil),
				},
			},
			args{
				[]byte(
					strings.Join(
						[]string{"Timeout=2", "invalid"},
						"\n",
					),
				),
			},
			true,
		},
		{
			"-initNVMLRunnerErr",
			fields{
				true,
				[]*nvmlmock.Expectation{},
			},
			args{},
			true,
		},
		{
			"-initNVMLRunnerCloseErr",
			fields{
				false,
				[]*nvmlmock.Expectation{
					nvmlmock.NewExpectation("Close").ProvideError(errs.New("fail")),
				},
			},
			args{},
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			runner := nvmlmock.NewMockRunner(t).ExpectCalls(tt.fields.runnerExpect...)

			rm := runnerSetMock{
				failed: tt.fields.failed,
			}

			p := &nvmlPlugin{
				nvmlRunner:    runner,
				setNvmlRunner: rm.init,
			}

			if err := p.Validate(tt.args.options); (err != nil) != tt.wantErr {
				t.Fatalf("nvmlPlugin.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
