package qiisync

import (
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestLoadConfiguration(t *testing.T) {
	tempDir, err := ioutil.TempDir("testdata", "temp")
	t.Cleanup(func() {
		if err := os.RemoveAll(tempDir); err != nil {
			t.Errorf("remove tempDir: %v", err)
		}
	})
	if err != nil {
		t.Errorf("create tempDir: %v", err)
	}

	os.Setenv("HOME", tempDir)
	os.Setenv("home", tempDir)
	os.Setenv("USERPROFILE", tempDir)

	if err := os.MkdirAll(filepath.Join(tempDir, ".config", "qiisync"), 0755); err != nil {
		t.Errorf("create config dir: %v", err)
	}

	f, err := os.Create(filepath.Join(tempDir, ".config", "qiisync", "config"))
	if err != nil {
		t.Errorf("create config: %v", err)
	}
	defer f.Close()

	f.WriteString(`[qiita]
api_token = "1234567890abcdefghijklmnopqrstuvwxyz1234"

[local]
base_dir = "./testdata/qiita"
filename_mode = "title"
`)

	got, err := LoadConfiguration()
	if err != nil {
		t.Errorf("loca config: %v", err)
	}
	want := &Config{
		Qiita: qiitaConfig{Token: "1234567890abcdefghijklmnopqrstuvwxyz1234"},
		Local: localConfig{Dir: "./testdata/qiita", FileNameMode: "title"},
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("LoadConfiguration() mismatch (-want +got):\n%s", diff)
	}
}

func Test_loadConfig(t *testing.T) {
	type args struct {
		r io.Reader
	}
	tests := []struct {
		name    string
		args    args
		want    *Config
		wantErr bool
	}{
		{
			name: "normal_linux_relative_title",
			args: args{
				r: strings.NewReader(`[qiita]
api_token = "1234567890abcdefghijklmnopqrstuvwxyz1234"

[local]
base_dir = "./testdata/qiita"
filename_mode = "title"`),
			},
			want: &Config{
				Qiita: qiitaConfig{Token: "1234567890abcdefghijklmnopqrstuvwxyz1234"},
				Local: localConfig{Dir: "./testdata/qiita", FileNameMode: "title"},
			},
			wantErr: false,
		},
		{
			name: "normal_linux_relative_id",
			args: args{
				r: strings.NewReader(`[qiita]
api_token = "1234567890abcdefghijklmnopqrstuvwxyz1234"

[local]
base_dir = "./testdata/qiita"
filename_mode = "id"`),
			},
			want: &Config{
				Qiita: qiitaConfig{Token: "1234567890abcdefghijklmnopqrstuvwxyz1234"},
				Local: localConfig{Dir: "./testdata/qiita", FileNameMode: "id"},
			},
			wantErr: false,
		},
		{
			name: "normal_windows_relative",
			args: args{
				r: strings.NewReader(`[qiita]
api_token = "1234567890abcdefghijklmnopqrstuvwxyz1234"

[local]
base_dir = ".\\testdata\\qiita"
filename_mode = "title"`),
			},
			want: &Config{
				Qiita: qiitaConfig{Token: "1234567890abcdefghijklmnopqrstuvwxyz1234"},
				Local: localConfig{Dir: `.\testdata\qiita`, FileNameMode: "title"},
			},
			wantErr: false,
		},
		{
			name: "invalid_linux_relative_title",
			args: args{
				r: strings.NewReader(`invalid`),
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := loadConfig(tt.args.r)
			if (err != nil) != tt.wantErr {
				t.Errorf("loadConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("loadConfig() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
