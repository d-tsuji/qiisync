package main

import (
	"io"
	"reflect"
	"strings"
	"testing"
)

func Test_loadConfig(t *testing.T) {
	type args struct {
		r io.Reader
	}
	tests := []struct {
		name    string
		args    args
		want    *config
		wantErr bool
	}{
		{
			name: "normal_linux_relative",
			args: args{
				r: strings.NewReader(`[qiita]
api_token = "1234567890abcdefghijklmnopqrstuvwxyz1234"

[local]
base_dir = "./testdata/qiita"`),
			},
			want: &config{
				Qiita: qiitaConfig{Token: "1234567890abcdefghijklmnopqrstuvwxyz1234"},
				Local: localConfig{Dir: "./testdata/qiita"},
			},
			wantErr: false,
		},
		{
			name: "normal_windows_relative",
			args: args{
				r: strings.NewReader(`[qiita]
api_token = "1234567890abcdefghijklmnopqrstuvwxyz1234"

[local]
base_dir = ".\\testdata\\qiita"`),
			},
			want: &config{
				Qiita: qiitaConfig{Token: "1234567890abcdefghijklmnopqrstuvwxyz1234"},
				Local: localConfig{Dir: `.\testdata\qiita`},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := loadConfig(tt.args.r)
			if (err != nil) != tt.wantErr {
				t.Errorf("loadConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("loadConfig() got = %v, want %v", got, tt.want)
			}
		})
	}
}
