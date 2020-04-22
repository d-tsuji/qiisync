package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
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

	got, err := broker.fetchRemoteArticle(&article{
		ArticleHeader: &ArticleHeader{ID: "c686397e4a0f4f11683d"},
	})
	if err != nil {
		t.Errorf("fetchRemoteArticle(): %v", err)
	}
	want := &article{
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
	want := []*article{
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
			FilePath: "",
		},
	}

	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("fetchRemoteItemsPerPage() mismatch (-want +got):\n%s", diff)
	}
}

func setup() (broker *Broker, mux *http.ServeMux, serverURL string, teardown func()) {
	mux = http.NewServeMux()

	apiHandler := http.NewServeMux()
	apiHandler.Handle("/", mux)

	server := httptest.NewServer(apiHandler)

	broker = NewBroker(&config{
		Qiita: qiitaConfig{Token: "1234567890abcdefghij"},
		Local: localConfig{
			Dir: "./testdata",
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

func TestBroker_fetchLocalArticles(t *testing.T) {
	updateAt := time.Date(2020, 4, 22, 16, 59, 59, 0, time.UTC)

	type fields struct {
		config  *config
		BaseURL *url.URL
	}
	tests := []struct {
		name         string
		fields       fields
		wantArticles map[string]*article
		wantErr      bool
	}{
		{
			name: "normal",
			fields: fields{config: &config{
				Qiita: qiitaConfig{Token: "1234567890abcdefghijklmnopqrstuvwxyz1234"},
				Local: localConfig{Dir: filepath.Join("testdata", "article")}}},
			wantArticles: map[string]*article{
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
	}

	if err := os.Chtimes(filepath.Join("testdata", "article", "20_test_article_posted.md"), updateAt, updateAt); err != nil {
		t.Error(err)
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			baseURL, _ := url.Parse(tt.fields.config.Local.Dir)
			b := &Broker{config: tt.fields.config, BaseURL: baseURL}

			gotArticles, err := b.fetchLocalArticles()
			if (err != nil) != tt.wantErr {
				t.Errorf("fetchLocalArticles() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if diff := cmp.Diff(tt.wantArticles, gotArticles); diff != "" {
				t.Errorf("fetchLocalArticles() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
