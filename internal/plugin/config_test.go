/*
TODO: Add Apache-2.0 licence
*/

package plugin

import (
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"golang.zabbix.com/sdk/log"
	"golang.zabbix.com/sdk/plugin"
)

func Test_examplePlugin_Configure(t *testing.T) {
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
		tt := tt
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
				t.Fatalf("examplePlugin_Configure() = %s", diff)
			}
		})
	}
}

func Test_examplePlugin_Validate(t *testing.T) {
	t.Parallel()

	type args struct {
		options any
	}

	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			"+valid",
			args{
				[]byte(`Timeout=30`),
			},
			false,
		},
		{
			"-unmarshalErr",
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
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			e := &nvmlPlugin{}
			if err := e.Validate(tt.args.options); (err != nil) != tt.wantErr {
				t.Fatalf("examplePlugin.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
