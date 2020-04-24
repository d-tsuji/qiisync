package qiisync

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/Songmu/flextime"
	"github.com/google/go-cmp/cmp"
)

func TestHeaderString(t *testing.T) {
	a := &Article{
		ArticleHeader: &ArticleHeader{
			ID:      "1234567890abcdefghij",
			Title:   "はじめてのGo",
			Tags:    "Go:1.14",
			Author:  "d-tsuji",
			Private: false,
		},
	}

	got, err := a.headerString()
	if err != nil {
		t.Errorf("headerString(): %v", err)
		return
	}
	want := `---
ID: 1234567890abcdefghij
Title: はじめてのGo
Tags: Go:1.14
Author: d-tsuji
Private: false
---

`
	if got != want {
		t.Errorf("Header string: %v, want %v", got, want)
	}
}

func TestFullContent(t *testing.T) {
	a := &Article{
		ArticleHeader: &ArticleHeader{
			ID:      "1234567890abcdefghij",
			Title:   "はじめてのGo",
			Tags:    "Go:1.14",
			Author:  "d-tsuji",
			Private: false,
		},
		Item: &Item{
			Body: "# はじめに\n\nはじめてのGoです\n",
		},
	}

	got, err := a.fullContent()
	if err != nil {
		t.Errorf("fullContent(): %v", err)
		return
	}
	want := `---
ID: 1234567890abcdefghij
Title: はじめてのGo
Tags: Go:1.14
Author: d-tsuji
Private: false
---

# はじめに

はじめてのGoです
`
	if got != want {
		t.Errorf("Header string: %v, want %v", got, want)
	}
}

func Test_articleFromFile(t *testing.T) {
	now := time.Date(2020, 4, 22, 16, 59, 59, 0, time.UTC)
	flextime.Fix(now)
	type args struct {
		filepath string
	}
	tests := []struct {
		name      string
		inputData string
		args      args
		want      *Article
		wantErr   bool
	}{
		{
			name: "normal",
			inputData: `---
ID: 1234567890abcdefghij
Title: テストTitle
Tags: Test:v0.0.1
Author: d-tsuji
Private: true
---

# はじめに

はじめてのQiitaです
`,
			args: args{filepath.Join("temp", "test.md")},
			want: &Article{
				ArticleHeader: &ArticleHeader{
					ID:      "1234567890abcdefghij",
					Title:   "テストTitle",
					Tags:    "Test:v0.0.1",
					Author:  "d-tsuji",
					Private: true,
				},
				Item:     &Item{Body: "# はじめに\n\nはじめてのQiitaです\n", UpdatedAt: now},
				FilePath: filepath.Join(".", "temp", "test.md"),
			},
			wantErr: false,
		},
		{
			name: "invalid_yaml",
			inputData: `---
ID: 1234567890abcdefghij
Title: テストTitle
Tags: Test:v0.0.1
Author: d-tsuji
Private: true

# はじめに

はじめてのQiitaです
`,
			args:    args{filepath.Join("temp", "test.md")},
			wantErr: true,
		},
		{
			name: "invalid_yaml_format",
			inputData: `---
- ID: 1234567890abcdefghij
---

# はじめに

はじめてのQiitaです
`,
			args:    args{filepath.Join("temp", "test.md")},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Cleanup(func() {
				if err := os.RemoveAll("temp"); err != nil {
					t.Error("tempDir remove:", err)
				}
			})

			if err := os.Mkdir("temp", 0777); err != nil {
				t.Errorf("create tempDir: %v", err)
				return
			}

			f, err := os.Create(filepath.Join("temp", "test.md"))
			if err != nil {
				t.Errorf("create tempFile: %v", err)
				return
			}
			defer f.Close()

			f.WriteString(tt.inputData)
			os.Chtimes(filepath.Join("temp", "test.md"), now, now)

			got, err := ArticleFromFile(tt.args.filepath)
			if (err != nil) != tt.wantErr {
				t.Errorf("ArticleFromFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("ArticleFromFile() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
