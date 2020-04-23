package qiisync

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
)

func Test_fetchRemoteArticle(t *testing.T) {
	broker, mux, _, teardown := setup()
	defer teardown()

	mux.HandleFunc("/api/v2/items/c686397e4a0f4f11683d", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		fmt.Fprint(w, `
					{
						"rendered_body": "<h1>Example</h1>",
						"body": "# Example",
						"coediting": false,
						"comments_count": 100,
						"created_at": "2000-01-01T00:00:00+00:00",
						"group": {
							"created_at": "2000-01-01T00:00:00+00:00",
							"id": 1,
							"name": "Dev",
							"private": false,
							"updated_at": "2000-01-01T00:00:00+00:00",
							"url_name": "dev"
						},
						"id": "c686397e4a0f4f11683d",
						"likes_count": 100,
						"private": false,
						"reactions_count": 100,
						"tags": [
							{
								"name": "Ruby",
								"versions": [
									"0.0.1"
								]
							}
						],
						"title": "Example title",
						"updated_at": "2000-01-01T00:00:00+00:00",
						"url": "https://qiita.com/Qiita/items/c686397e4a0f4f11683d",
						"user": {
							"description": "Hello, world.",
							"facebook_id": "qiita",
							"followees_count": 100,
							"followers_count": 200,
							"github_login_name": "qiitan",
							"id": "qiita",
							"items_count": 300,
							"linkedin_id": "qiita",
							"location": "Tokyo, Japan",
							"name": "Qiita キータ",
							"organization": "Increments Inc",
							"permanent_id": 1,
							"profile_image_url": "https://s3-ap-northeast-1.amazonaws.com/qiita-image-store/0/88/ccf90b557a406157dbb9d2d7e543dae384dbb561/large.png?1575443439",
							"team_only": false,
							"twitter_screen_name": "qiita",
							"website_url": "https://qiita.com"
						},
						"page_views_count": 100
					}
`)
	})

	got, err := broker.fetchRemoteArticle(&Article{
		ArticleHeader: &ArticleHeader{ID: "c686397e4a0f4f11683d"},
	})
	if err != nil {
		t.Errorf("fetchRemoteArticle(): %v", err)
	}
	want := &Article{
		ArticleHeader: &ArticleHeader{
			ID:      "c686397e4a0f4f11683d",
			Title:   "Example title",
			Tags:    "Ruby:0.0.1",
			Author:  "Qiita キータ",
			Private: false,
		},
		Item: &Item{
			ID:  "c686397e4a0f4f11683d",
			URL: "https://qiita.com/Qiita/items/c686397e4a0f4f11683d",
			User: User{
				ID:   "qiita",
				Name: "Qiita キータ",
			},
			Title:        "Example title",
			Body:         "# Example",
			RenderedBody: "<h1>Example</h1>",
			CreatedAt:    time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC),
			UpdatedAt:    time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC),
			Tags: []*Tag{
				{
					Name:     "Ruby",
					Versions: []string{"0.0.1"},
				},
			},
			Private: false,
		},
		FilePath: "",
	}

	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("fetchRemoteItemsPerPage() mismatch (-want +got):\n%s", diff)
	}
}

func Test_fetchRemoteArticles(t *testing.T) {
	broker, mux, _, teardown := setup()
	t.Cleanup(func() {
		teardown()
	})

	mux.HandleFunc("/api/v2/authenticated_user/items", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		w.Header().Set("Total-Count", "2")
		pageNum, err := strconv.Atoi(r.FormValue("page"))
		if err != nil {
			t.Errorf("convert int: %s, %v", r.FormValue("page"), err)
			return
		}
		if pageNum == 1 {
			fmt.Fprint(w, `
					[
						{
							"rendered_body": "<h1>Example</h1>",
							"body": "# Example",
							"created_at": "2000-01-01T00:00:00+00:00",
							"id": "c686397e4a0f4f11683d",
							"private": false,
							"tags": [
								{
									"name": "Ruby",
									"versions": [
										"0.0.1"
									]
								}
							],
							"title": "Example title",
							"updated_at": "2000-01-01T00:00:00+00:00",
							"url": "https://qiita.com/Qiita/items/c686397e4a0f4f11683d",
							"user": {
								"id": "qiita",
								"name": "Qiita キータ"
							},
							"page_views_count": 100
						}
					]
`)
		} else if pageNum == 2 {
			fmt.Fprint(w, `
					[
						{
							"rendered_body": "<h1>Example2</h1>",
							"body": "# Example2",
							"created_at": "2000-01-01T00:00:00+00:00",
							"id": "c686397e4a0f4f11683d",
							"private": false,
							"tags": [
								{
									"name": "Ruby",
									"versions": [
										"0.0.1"
									]
								}
							],
							"title": "Example title2",
							"updated_at": "2000-01-01T00:00:00+00:00",
							"url": "https://qiita.com/Qiita/items/c686397e4a0f4f11683d",
							"user": {
								"id": "qiita2",
								"name": "Qiita キータ2"
							},
							"page_views_count": 100
						}
					]
`)
		}
	})

	defaultItemsPerPage = 1
	got, err := broker.FetchRemoteArticles()
	if err != nil {
		t.Errorf("FetchRemoteArticles(): %v", err)
		return
	}
	want := []*Article{
		{
			ArticleHeader: &ArticleHeader{
				ID:      "c686397e4a0f4f11683d",
				Title:   "Example title",
				Tags:    "Ruby:0.0.1",
				Author:  "Qiita キータ",
				Private: false,
			},
			Item: &Item{
				ID:  "c686397e4a0f4f11683d",
				URL: "https://qiita.com/Qiita/items/c686397e4a0f4f11683d",
				User: User{
					ID:   "qiita",
					Name: "Qiita キータ",
				},
				Title:        "Example title",
				Body:         "# Example",
				RenderedBody: "<h1>Example</h1>",
				CreatedAt:    time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC),
				UpdatedAt:    time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC),
				Tags: []*Tag{
					{
						Name:     "Ruby",
						Versions: []string{"0.0.1"},
					},
				},
				Private: false,
			},
		},
		{
			ArticleHeader: &ArticleHeader{
				ID:      "c686397e4a0f4f11683d",
				Title:   "Example title2",
				Tags:    "Ruby:0.0.1",
				Author:  "Qiita キータ2",
				Private: false,
			},
			Item: &Item{
				ID:  "c686397e4a0f4f11683d",
				URL: "https://qiita.com/Qiita/items/c686397e4a0f4f11683d",
				User: User{
					ID:   "qiita2",
					Name: "Qiita キータ2",
				},
				Title:        "Example title2",
				Body:         "# Example2",
				RenderedBody: "<h1>Example2</h1>",
				CreatedAt:    time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC),
				UpdatedAt:    time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC),
				Tags: []*Tag{
					{
						Name:     "Ruby",
						Versions: []string{"0.0.1"},
					},
				},
				Private: false,
			},
		},
	}

	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("FetchRemoteArticles() mismatch (-want +got):\n%s", diff)
	}
}

func Test_fetchRemoteItemsPerPage(t *testing.T) {
	broker, mux, _, teardown := setup()
	defer teardown()

	mux.HandleFunc("/api/v2/authenticated_user/items", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		w.Header().Set("Total-Count", "1")
		fmt.Fprint(w, `
					[
						{
							"rendered_body": "<h1>Example</h1>",
							"body": "# Example",
							"coediting": false,
							"comments_count": 100,
							"created_at": "2000-01-01T00:00:00+00:00",
							"group": {
								"created_at": "2000-01-01T00:00:00+00:00",
								"id": 1,
								"name": "Dev",
								"private": false,
								"updated_at": "2000-01-01T00:00:00+00:00",
								"url_name": "dev"
							},
							"id": "c686397e4a0f4f11683d",
							"likes_count": 100,
							"private": false,
							"reactions_count": 100,
							"tags": [
								{
									"name": "Ruby",
									"versions": [
										"0.0.1"
									]
								}
							],
							"title": "Example title",
							"updated_at": "2000-01-01T00:00:00+00:00",
							"url": "https://qiita.com/Qiita/items/c686397e4a0f4f11683d",
							"user": {
								"description": "Hello, world.",
								"facebook_id": "qiita",
								"followees_count": 100,
								"followers_count": 200,
								"github_login_name": "qiitan",
								"id": "qiita",
								"items_count": 300,
								"linkedin_id": "qiita",
								"location": "Tokyo, Japan",
								"name": "Qiita キータ",
								"organization": "Increments Inc",
								"permanent_id": 1,
								"profile_image_url": "https://s3-ap-northeast-1.amazonaws.com/qiita-image-store/0/88/ccf90b557a406157dbb9d2d7e543dae384dbb561/large.png?1575443439",
								"team_only": false,
								"twitter_screen_name": "qiita",
								"website_url": "https://qiita.com"
							},
							"page_views_count": 100
						}
					]
`)
	})

	got, _, err := broker.fetchRemoteItemsPerPage(1)
	if err != nil {
		t.Errorf("fetchRemoteItemsPerPage(): %v", err)
	}
	want := []*Article{
		{
			ArticleHeader: &ArticleHeader{
				ID:      "c686397e4a0f4f11683d",
				Title:   "Example title",
				Tags:    "Ruby:0.0.1",
				Author:  "Qiita キータ",
				Private: false,
			},
			Item: &Item{
				ID:  "c686397e4a0f4f11683d",
				URL: "https://qiita.com/Qiita/items/c686397e4a0f4f11683d",
				User: User{
					ID:   "qiita",
					Name: "Qiita キータ",
				},
				Title:        "Example title",
				Body:         "# Example",
				RenderedBody: "<h1>Example</h1>",
				CreatedAt:    time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC),
				UpdatedAt:    time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC),
				Tags: []*Tag{
					{
						Name:     "Ruby",
						Versions: []string{"0.0.1"},
					},
				},
				Private: false,
			},
		},
	}

	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("fetchRemoteItemsPerPage() mismatch (-want +got):\n%s", diff)
	}
}

func TestLocalPath(t *testing.T) {
	a := &Article{
		ArticleHeader: &ArticleHeader{
			ID:      "1234567890abcdefghij",
			Title:   "はじめてのGo",
			Tags:    "Go:1.14",
			Author:  "d-tsuji",
			Private: false,
		},
		Item: &Item{
			Title:     "はじめてのGo",
			Body:      "# はじめに\n\nはじめてのGoです\n",
			CreatedAt: time.Date(2020, 4, 22, 16, 59, 59, 0, time.UTC),
		},
	}

	b := &Broker{
		Config: &Config{
			Local: localConfig{Dir: filepath.Join("testdata", "article")},
		},
	}

	got := b.localPath(a)
	want := filepath.Join("testdata", "article", "20200422", "はじめてのGo.md")

	if got != want {
		t.Errorf("localPath() = %v, want %v", got, want)
	}
}

func TestStoreFresh(t *testing.T) {
	type fields struct {
		config  *Config
		BaseURL *url.URL
	}
	type args struct {
		localArticles map[string]*Article
		remoteArticle *Article
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    bool
		wantErr bool
	}{
		{
			name: "fresh",
			fields: fields{
				config: &Config{
					Qiita: qiitaConfig{Token: "1234567890abcdefghijklmnopqrstuvwxyz1234"},
					Local: localConfig{
						Dir: "./testdata/broker",
					},
				},
			},
			args: args{
				localArticles: map[string]*Article{
					"1234567890abcdefghij": {
						ArticleHeader: &ArticleHeader{
							ID:      "1234567890abcdefghij",
							Title:   "はじめてのGo",
							Tags:    "Go:1.14",
							Author:  "d-tsuji",
							Private: false,
						},
						Item: &Item{
							Body:      "# はじめに\n\nはじめてのGoです\n",
							UpdatedAt: time.Date(2020, 4, 22, 16, 59, 59, 0, time.UTC),
						},
						FilePath: filepath.Join("testdata", "broker", "TestStoreFresh.md"),
					},
				},
				remoteArticle: &Article{
					ArticleHeader: &ArticleHeader{
						ID:      "1234567890abcdefghij",
						Title:   "はじめてのGo",
						Tags:    "Go:1.14",
						Author:  "d-tsuji",
						Private: false,
					},
					Item: &Item{
						Body:      "# はじめに\n\nはじめてのGoです\n",
						UpdatedAt: time.Date(2020, 4, 22, 17, 00, 00, 0, time.UTC),
					},
				},
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "not_fresh",
			fields: fields{
				config: &Config{
					Qiita: qiitaConfig{Token: "1234567890abcdefghijklmnopqrstuvwxyz1234"},
					Local: localConfig{
						Dir: "./testdata/broker",
					},
				},
			},
			args: args{
				localArticles: map[string]*Article{
					"1234567890abcdefghij": {
						ArticleHeader: &ArticleHeader{
							ID:      "1234567890abcdefghij",
							Title:   "はじめてのGo",
							Tags:    "Go:1.14",
							Author:  "d-tsuji",
							Private: false,
						},
						Item: &Item{
							Body:      "# はじめに\n\nはじめてのGoです\n",
							UpdatedAt: time.Date(2020, 4, 22, 16, 59, 59, 0, time.UTC),
						},
						FilePath: filepath.Join("testdata", "broker", "TestStoreFresh.md"),
					},
				},
				remoteArticle: &Article{
					ArticleHeader: &ArticleHeader{
						ID:      "1234567890abcdefghij",
						Title:   "はじめてのGo",
						Tags:    "Go:1.14",
						Author:  "d-tsuji",
						Private: false,
					},
					Item: &Item{
						Body:      "# はじめに\n\nはじめてのGoです\n",
						UpdatedAt: time.Date(2020, 4, 22, 16, 59, 58, 0, time.UTC),
					},
				},
			},
			want:    false,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			baseURL, _ := url.Parse(tt.fields.config.Local.Dir)
			b := &Broker{Config: tt.fields.config, BaseURL: baseURL}

			got, err := b.StoreFresh(tt.args.localArticles, tt.args.remoteArticle)
			if (err != nil) != tt.wantErr {
				t.Errorf("StoreFresh() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("StoreFresh() got = %v, want %v", got, tt.want)
			}
		})
	}
}
func TestStore(t *testing.T) {
	tempDir, err := ioutil.TempDir("testdata", "temp")
	if err != nil {
		t.Errorf("create tempDir: %v", err)
		return
	}
	t.Cleanup(func() {
		if err := os.RemoveAll(tempDir); err != nil {
			t.Errorf("remove tempDir: %v", err)
		}
	})

	a := &Article{
		ArticleHeader: &ArticleHeader{
			ID:      "1234567890abcdefghij",
			Title:   "はじめてのGo",
			Tags:    "Go:1.14",
			Author:  "d-tsuji",
			Private: false,
		},
		Item: &Item{
			Body:      "# はじめに\n\nはじめてのGoです\n",
			CreatedAt: time.Date(2020, 4, 22, 16, 59, 59, 0, time.UTC),
		},
	}

	b := &Broker{
		Config: &Config{
			Local: localConfig{Dir: filepath.Join(tempDir)},
		},
	}

	fpath := filepath.Join(tempDir, "test.md")
	if err := b.store(fpath, a); err != nil {
		t.Errorf("store(): %v", err)
		return
	}

	fbyte, err := ioutil.ReadFile(fpath)
	if err != nil {
		t.Errorf("read file: %s, %v", fpath, err)
	}
	got := string(fbyte)
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
		t.Errorf("Stored file string: %v, want %v", got, want)
	}
}

func Test_convertItemsArticles(t *testing.T) {
	b := &Broker{Config: &Config{Local: localConfig{FileNameMode: "title"}}}
	items := []*Item{
		{
			ID:        "111",
			Title:     "111",
			Body:      "111",
			CreatedAt: time.Date(2020, 4, 22, 00, 00, 00, 0, time.UTC),
		},
		{
			ID:        "222",
			Title:     "222",
			Body:      "222",
			CreatedAt: time.Date(2020, 4, 22, 00, 00, 00, 0, time.UTC),
		},
		{
			ID:        "333",
			Title:     "333",
			Body:      "333",
			CreatedAt: time.Date(2020, 4, 23, 00, 00, 00, 0, time.UTC),
		},
		{
			ID:        "444",
			Title:     "333",
			Body:      "333",
			CreatedAt: time.Date(2020, 4, 23, 00, 00, 00, 0, time.UTC),
		},
		{
			ID:        "555",
			Title:     "333",
			Body:      "333",
			CreatedAt: time.Date(2020, 4, 24, 00, 00, 00, 0, time.UTC),
		},
	}

	got := b.convertItemsArticles(items)
	want := []*Article{
		{
			ArticleHeader: &ArticleHeader{ID: "111", Title: "111"},
			Item: &Item{ID: "111", Title: "111", Body: "111",
				CreatedAt: time.Date(2020, 4, 22, 00, 00, 00, 0, time.UTC),
			},
		},
		{
			ArticleHeader: &ArticleHeader{ID: "222", Title: "222"},
			Item: &Item{ID: "222", Title: "222", Body: "222",
				CreatedAt: time.Date(2020, 4, 22, 00, 00, 00, 0, time.UTC),
			},
		},
		{
			ArticleHeader: &ArticleHeader{ID: "333", Title: "333"},
			Item: &Item{ID: "333", Title: "333", Body: "333",
				CreatedAt: time.Date(2020, 4, 23, 00, 00, 00, 0, time.UTC),
			},
		},
		{
			ArticleHeader: &ArticleHeader{ID: "444", Title: "333"},
			Item: &Item{ID: "444", Title: "333_2", Body: "333",
				CreatedAt: time.Date(2020, 4, 23, 00, 00, 00, 0, time.UTC),
			},
		},
		{
			ArticleHeader: &ArticleHeader{ID: "555", Title: "333"},
			Item: &Item{ID: "555", Title: "333", Body: "333",
				CreatedAt: time.Date(2020, 4, 24, 00, 00, 00, 0, time.UTC),
			},
		},
	}

	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("convertItemsArticles() mismatch (-want +got):\n%s", diff)
	}
}

func TestPostArticle(t *testing.T) {
	b, mux, _, teardown := setup()
	t.Cleanup(func() {
		teardown()
		if err := os.RemoveAll(b.baseDir()); err != nil {
			t.Errorf("remove tempDir: %v", err)
		}
	})

	mux.HandleFunc("/api/v2/items", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "POST")
		w.WriteHeader(http.StatusCreated)
		fmt.Fprint(w, `
					{
						"rendered_body": "<h1>Example</h1>",
						"body": "# Example",
						"coediting": false,
						"comments_count": 100,
						"created_at": "2000-01-01T00:00:00+00:00",
						"group": {
							"created_at": "2000-01-01T00:00:00+00:00",
							"id": 1,
							"name": "Dev",
							"private": false,
							"updated_at": "2000-01-01T00:00:00+00:00",
							"url_name": "dev"
						},
						"id": "c686397e4a0f4f11683d",
						"likes_count": 100,
						"private": false,
						"reactions_count": 100,
						"tags": [
							{
								"name": "Ruby",
								"versions": [
									"0.0.1"
								]
							}
						],
						"title": "Example title",
						"updated_at": "2000-01-01T00:00:00+00:00",
						"url": "https://localhost/Test/items/c686397e4a0f4f11683d",
						"user": {
							"description": "Hello, world.",
							"facebook_id": "qiita",
							"followees_count": 100,
							"followers_count": 200,
							"github_login_name": "qiitan",
							"id": "qiita",
							"items_count": 300,
							"linkedin_id": "qiita",
							"location": "Tokyo, Japan",
							"name": "Qiita キータ",
							"organization": "Increments Inc",
							"permanent_id": 1,
							"profile_image_url": "https://s3-ap-northeast-1.amazonaws.com/qiita-image-store/0/88/ccf90b557a406157dbb9d2d7e543dae384dbb561/large.png?1575443439",
							"team_only": false,
							"twitter_screen_name": "qiita",
							"website_url": "https://qiita.com"
						},
						"page_views_count": 100
					}
`)
	})

	err := b.PostArticle(&PostItem{
		Body:    "# Example",
		Private: false,
		Tags: []*Tag{
			{
				Name:     "Ruby",
				Versions: []string{"0.0.1"},
			},
		},
		Title: "Example title",
	})
	if err != nil {
		t.Errorf("PostArticle(): %v", err)
		return
	}
}

func TestPatchArticle(t *testing.T) {
	b, mux, _, teardown := setup()
	t.Cleanup(func() {
		teardown()
		if err := os.RemoveAll(b.baseDir()); err != nil {
			t.Errorf("remove tempDir: %v", err)
		}
	})

	mux.HandleFunc("/api/v2/items/c686397e4a0f4f11683d", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "PATCH")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `
					{
						"rendered_body": "<h1>Example</h1>",
						"body": "# Example",
						"coediting": false,
						"comments_count": 100,
						"created_at": "2000-01-01T00:00:00+00:00",
						"group": {
							"created_at": "2000-01-01T00:00:00+00:00",
							"id": 1,
							"name": "Dev",
							"private": false,
							"updated_at": "2000-01-01T00:00:00+00:00",
							"url_name": "dev"
						},
						"id": "c686397e4a0f4f11683d",
						"likes_count": 100,
						"private": false,
						"reactions_count": 100,
						"tags": [
							{
								"name": "Ruby",
								"versions": [
									"0.0.1"
								]
							}
						],
						"title": "Example title",
						"updated_at": "2000-01-01T00:00:00+00:00",
						"url": "https://localhost/Test/items/c686397e4a0f4f11683d",
						"user": {
							"description": "Hello, world.",
							"facebook_id": "qiita",
							"followees_count": 100,
							"followers_count": 200,
							"github_login_name": "qiitan",
							"id": "qiita",
							"items_count": 300,
							"linkedin_id": "qiita",
							"location": "Tokyo, Japan",
							"name": "Qiita キータ",
							"organization": "Increments Inc",
							"permanent_id": 1,
							"profile_image_url": "https://s3-ap-northeast-1.amazonaws.com/qiita-image-store/0/88/ccf90b557a406157dbb9d2d7e543dae384dbb561/large.png?1575443439",
							"team_only": false,
							"twitter_screen_name": "qiita",
							"website_url": "https://qiita.com"
						},
						"page_views_count": 100
					}
`)
	})

	err := b.patchArticle(&PostItem{
		Body:    "# Example",
		Private: false,
		Tags: []*Tag{
			{
				Name:     "Ruby",
				Versions: []string{"0.0.1"},
			},
		},
		Title: "Example title",
		ID:    "c686397e4a0f4f11683d",
	})
	if err != nil {
		t.Errorf("patchArticle(): %v", err)
		return
	}
}

func Test_fetchLocalArticles(t *testing.T) {
	updateAt := time.Date(2020, 4, 22, 16, 59, 59, 0, time.UTC)

	type fields struct {
		config  *Config
		BaseURL *url.URL
	}
	tests := []struct {
		name         string
		fields       fields
		wantArticles map[string]*Article
		wantErr      bool
	}{
		{
			name: "normal",
			fields: fields{config: &Config{
				Qiita: qiitaConfig{Token: "1234567890abcdefghijklmnopqrstuvwxyz1234"},
				Local: localConfig{Dir: filepath.Join("testdata", "article")}}},
			wantArticles: map[string]*Article{
				"1234567890abcdefghij": {
					ArticleHeader: &ArticleHeader{
						ID:      "1234567890abcdefghij",
						Title:   "はじめてのGo",
						Tags:    "Go:1.14",
						Author:  "d-tsuji",
						Private: false,
					},
					Item: &Item{
						Body:      "# はじめに\n\nはじめてのGoです\n",
						UpdatedAt: updateAt,
					},
					FilePath: filepath.Join("testdata", "article", "20_test_article_posted.md"),
				},
			},
			wantErr: false,
		},
		{
			name: "id_duplicated",
			fields: fields{config: &Config{
				Qiita: qiitaConfig{Token: "1234567890abcdefghijklmnopqrstuvwxyz1234"},
				Local: localConfig{Dir: filepath.Join("testdata", "duplicate")}}},
			wantErr: true,
		},
	}

	if err := os.Chtimes(filepath.Join("testdata", "article", "20_test_article_posted.md"), updateAt, updateAt); err != nil {
		t.Error(err)
	}
	if err := os.Chtimes(filepath.Join("testdata", "duplicate", "20_test_article_posted.md"), updateAt, updateAt); err != nil {
		t.Error(err)
	}
	if err := os.Chtimes(filepath.Join("testdata", "duplicate", "21_test_article_posted_copy.md"), updateAt, updateAt); err != nil {
		t.Error(err)
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			baseURL, _ := url.Parse(tt.fields.config.Local.Dir)
			b := &Broker{Config: tt.fields.config, BaseURL: baseURL}

			gotArticles, err := b.FetchLocalArticles()
			if (err != nil) != tt.wantErr {
				t.Errorf("FetchLocalArticles() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if diff := cmp.Diff(tt.wantArticles, gotArticles); diff != "" {
				t.Errorf("FetchLocalArticles() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestUploadFresh(t *testing.T) {
	tests := []struct {
		name         string
		localArticle *Article
		want         bool
		wantErr      bool
	}{
		{
			name: "fresh",
			localArticle: &Article{
				ArticleHeader: &ArticleHeader{
					ID:      "c686397e4a0f4f11683d",
					Title:   "Update title",
					Tags:    "Go:1.14",
					Author:  "d-tsuji",
					Private: false,
				},
				Item: &Item{
					Body:      "# Update Example",
					UpdatedAt: time.Date(2020, 4, 23, 05, 41, 36, 0, time.UTC),
				},
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "not_fresh",
			localArticle: &Article{
				ArticleHeader: &ArticleHeader{
					ID:      "c686397e4a0f4f11683d",
					Title:   "Update title",
					Tags:    "Go:1.14",
					Author:  "d-tsuji",
					Private: false,
				},
				Item: &Item{
					Body:      "# Update Example",
					UpdatedAt: time.Date(2020, 4, 23, 05, 41, 34, 0, time.UTC),
				},
			},
			want:    false,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			b, mux, _, teardown := setup()
			t.Cleanup(func() {
				teardown()
			})

			mux.HandleFunc("/api/v2/items/c686397e4a0f4f11683d", func(w http.ResponseWriter, r *http.Request) {
				switch r.Method {
				case "GET":
					fmt.Fprint(w, `
					{
						"rendered_body": "<h1>Example</h1>",
						"body": "# Example",
						"created_at": "2020-04-23T05:41:35+00:00",
						"id": "c686397e4a0f4f11683d",
						"private": false,
						"tags": [
							{
								"name": "Ruby",
								"versions": [
									"0.0.1"
								]
							}
						],
						"title": "Example title",
						"updated_at": "2020-04-23T05:41:35+00:00",
						"url": "https://localhost/Test/items/c686397e4a0f4f11683d",
						"user": {
							"id": "qiita",
							"name": "Qiita キータ"
						}
					}
`)
				case "PATCH":
					fmt.Fprint(w, `
					{
						"rendered_body": "<h1>Example</h1>",
						"body": "# Update Example",
						"created_at": "2020-04-23T05:41:35+00:00",
						"id": "c686397e4a0f4f11683d",
						"private": false,
						"tags": [
							{
								"name": "Go",
								"versions": [
									"1.14"
								]
							}
						],
						"title": "Update title",
						"updated_at": "2020-04-23T05:41:36+00:00",
						"url": "https://localhost/Test/items/c686397e4a0f4f11683d",
						"user": {
							"id": "d-tsuji",
							"name": "d-tsuji"
						}
					}
`)
				}
			})

			got, err := b.UploadFresh(tt.localArticle)
			if (err != nil) != tt.wantErr {
				t.Errorf("UploadFresh() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("UploadFresh() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStoreFilename(t *testing.T) {
	type fields struct {
		config *Config
	}
	tests := []struct {
		name   string
		fields fields
		a      *Article
		want   string
	}{
		{
			name:   "default",
			fields: fields{config: &Config{Local: localConfig{FileNameMode: ""}}},
			a:      &Article{Item: &Item{ID: "1234567890abcdefghij", Title: "はじめてのGo"}},
			want:   "はじめてのGo.md",
		},
		{
			name:   "title",
			fields: fields{config: &Config{Local: localConfig{FileNameMode: "title"}}},
			a:      &Article{Item: &Item{ID: "1234567890abcdefghij", Title: "はじめてのGo"}},
			want:   "はじめてのGo.md",
		},
		{
			name:   "id",
			fields: fields{config: &Config{Local: localConfig{FileNameMode: "id"}}},
			a:      &Article{Item: &Item{ID: "1234567890abcdefghij", Title: "はじめてのGo"}},
			want:   "1234567890abcdefghij.md",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &Broker{Config: tt.fields.config}
			if got := b.storeFileName(tt.a); got != tt.want {
				t.Errorf("storeFileName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func setup() (broker *Broker, mux *http.ServeMux, serverURL string, teardown func()) {
	mux = http.NewServeMux()

	apiHandler := http.NewServeMux()
	apiHandler.Handle("/", mux)

	server := httptest.NewServer(apiHandler)

	broker = NewBroker(&Config{
		Qiita: qiitaConfig{Token: "1234567890abcdefghijklmnopqrstuvwxyz1234"},
		Local: localConfig{
			Dir: "./testdata/broker",
		},
	})
	baseURL, _ := url.Parse(server.URL + "/")
	broker.BaseURL = baseURL

	return broker, mux, server.URL, server.Close
}

func testMethod(t *testing.T, r *http.Request, want string) {
	t.Helper()
	if got := r.Method; got != want {
		t.Errorf("Request method: %v, want %v", got, want)
	}
}
