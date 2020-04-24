package qiisync

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

var (
	// Define it with var so that it can be replaced with
	// the URL of the mock server of httptest when testing.
	defaultBaseURL      = "https://qiita.com/"
	defaultItemsPerPage = 20
	defaultExtension    = ".md"
)

// Broker is the core structure of qiisync that handles
// Qiita and the local filesystem with each other.
type Broker struct {
	*Config
	BaseURL *url.URL
}

// NewBroker create a Broker.
func NewBroker(c *Config) *Broker {
	baseURL, _ := url.Parse(defaultBaseURL)
	return &Broker{
		Config:  c,
		BaseURL: baseURL,
	}
}

func (b *Broker) do(req *http.Request) (*http.Response, error) {
	return http.DefaultClient.Do(req)
}

func (b *Broker) fetchRemoteArticle(a *Article) (*Article, error) {
	if a.ID == "" {
		return nil, errors.New("article ID is required")
	}
	u := fmt.Sprintf("api/v2/items/%s", a.ID)
	req, err := b.NewRequest(http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}
	resp, err := b.do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New(resp.Status)
	}

	var item Item
	if err := json.NewDecoder(resp.Body).Decode(&item); err != nil {
		return nil, err
	}

	return b.convertItemsArticle(&item), nil
}

// FetchRemoteArticles extracts articles from Qiita.
func (b *Broker) FetchRemoteArticles() ([]*Article, error) {
	var articles []*Article
	for i := 1; ; i++ {
		aarticles, hasNext, err := b.fetchRemoteItemsPerPage(i)
		if err != nil {
			return nil, err
		}
		articles = append(articles, aarticles...)

		if !hasNext {
			break
		}
	}
	return articles, nil
}

func (b *Broker) fetchRemoteItemsPerPage(page int) ([]*Article, bool, error) {
	u := fmt.Sprintf("api/v2/authenticated_user/items?page=%d&per_page=%d", page, defaultItemsPerPage)
	req, err := b.NewRequest(http.MethodGet, u, nil)
	if err != nil {
		return nil, false, err
	}

	resp, err := b.do(req)
	if err != nil {
		return nil, false, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, false, errors.New(resp.Status)
	}

	var items []*Item
	if err := json.NewDecoder(resp.Body).Decode(&items); err != nil {
		return nil, false, err
	}

	total, err := strconv.Atoi(resp.Header.Get("Total-Count"))
	if err != nil {
		return nil, false, err
	}

	return b.convertItemsArticles(items), defaultItemsPerPage*page < total, nil
}

// FetchLocalArticles searches base_dir of local filesystem and extracts articles.
func (b *Broker) FetchLocalArticles() (articles map[string]*Article, err error) {
	articles = make(map[string]*Article)
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("recoverd when dirwalk(%s): %v", b.baseDir(), r)
		}
	}()
	fnameList := dirwalk(b.baseDir())
	for i := range fnameList {
		a, err := ArticleFromFile(fnameList[i])
		if err != nil {
			return nil, err
		}
		// If ArticleHeader.ID is empty, it just indicates a new file.
		if a.ArticleHeader.ID == "" {
			continue
		}
		if ea, exists := articles[a.ArticleHeader.ID]; exists {
			return nil, fmt.Errorf("duplicate ID in local: %s and %s", ea.FilePath, a.FilePath)
		}
		articles[a.ArticleHeader.ID] = a
	}
	return articles, nil
}

func dirwalk(dir string) []string {
	err := os.MkdirAll(dir, 0755)
	if err != nil {
		panic(err)
	}
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		panic(err)
	}

	var paths []string
	for _, file := range files {
		if file.IsDir() {
			paths = append(paths, dirwalk(filepath.Join(dir, file.Name()))...)
			continue
		}
		paths = append(paths, filepath.Join(dir, file.Name()))
	}

	return paths
}

// NewRequest is a testable NewRequest that wraps http.NewRequest.
func (b *Broker) NewRequest(method, urlStr string, body interface{}) (*http.Request, error) {
	if !strings.HasSuffix(b.BaseURL.Path, "/") {
		return nil, fmt.Errorf("BaseURL must have a trailing slash, but %q does not", b.BaseURL)
	}
	u, err := b.BaseURL.Parse(urlStr)
	if err != nil {
		return nil, err
	}

	var buf io.ReadWriter
	if body != nil {
		buf = &bytes.Buffer{}
		enc := json.NewEncoder(buf)
		enc.SetEscapeHTML(false)
		err := enc.Encode(body)
		if err != nil {
			return nil, err
		}
	}
	req, err := http.NewRequest(method, u.String(), buf)
	if err != nil {
		return nil, err
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", b.Qiita.Token))

	return req, nil
}

func (b *Broker) localPath(a *Article) string {
	paths := []string{b.baseDir()}
	paths = append(paths, dateFormat(a.Item.CreatedAt), b.storeFileName(a))
	return filepath.Join(paths...)
}

// StoreFresh compares the files in the local filesystem with the articles retrieved from Qiita and
// updates the files in the local filesystem.
func (b *Broker) StoreFresh(localArticles map[string]*Article, remoteArticle *Article) (bool, error) {
	var localLastModified time.Time
	path := filepath.Join(b.baseDir(), dateFormat(remoteArticle.Item.CreatedAt), b.storeFileName(remoteArticle))

	a, exists := localArticles[remoteArticle.ID]
	if exists {
		localLastModified = a.Item.UpdatedAt
		path = a.FilePath
	}
	if remoteArticle.Item.UpdatedAt.After(localLastModified) {
		Logf("fresh", "remote=%s > local=%s", remoteArticle.Item.UpdatedAt, localLastModified)
		if err := b.store(path, remoteArticle); err != nil {
			return false, err
		}
		return true, nil
	}

	return false, nil
}

func (b *Broker) store(path string, article *Article) error {
	Logf("store", "%s", path)

	dir, _ := filepath.Split(path)
	err := os.MkdirAll(dir, 0755)
	if err != nil {
		return err
	}

	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	fullContext, err := article.fullContent()
	if err != nil {
		return err
	}
	f.WriteString(fullContext)

	return os.Chtimes(path, article.Item.UpdatedAt, article.Item.UpdatedAt)
}

func (b *Broker) convertItemsArticles(items []*Item) []*Article {
	articles := make([]*Article, len(items))
	fileCount := make(map[string]int, len(items))
	for i := range items {
		articles[i] = b.convertItemsArticle(items[i])

		// If we use the title of the article in Qiita as the file name to save locally,
		// the file name very rarely be duplicated.
		// Therefore, if occur, the file name is set to a sequential number to avoid it.
		if b.isFileNameModeTitle() {
			path := filepath.Join(dateFormat(items[i].CreatedAt), items[i].Title)
			cnt, exists := fileCount[path]
			if exists {
				articles[i].Item.Title = fmt.Sprintf("%s_%d", articles[i].Item.Title, cnt+1)
			}
			fileCount[path] = cnt + 1
		}
	}
	return articles
}

func (b *Broker) convertItemsArticle(item *Item) *Article {
	return &Article{
		ArticleHeader: &ArticleHeader{
			ID:      item.ID,
			Title:   item.Title,
			Tags:    unmarshalTag(item.Tags),
			Author:  item.User.Name,
			Private: item.Private,
		},
		Item: item,
	}
}

// PostArticle post the article on Qiita.
func (b *Broker) PostArticle(body *PostItem) error {
	req, err := b.NewRequest(http.MethodPost, "api/v2/items", body)
	if err != nil {
		return err
	}

	resp, err := b.do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return errors.New(resp.Status)
	}

	var r PostItemResult
	if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
		return fmt.Errorf("json decode: %w", err)
	}
	Logf("post", "article ---> %s", r.URL)

	article := &Article{
		ArticleHeader: &ArticleHeader{
			ID:      r.ID,
			Title:   r.Title,
			Tags:    unmarshalTag(r.Tags),
			Author:  r.User.Name,
			Private: r.Private,
		},
		Item: &Item{
			Body:      r.Body,
			CreatedAt: r.CreatedAt,
			UpdatedAt: r.UpdatedAt,
		},
	}

	path := filepath.Join(b.baseDir(), r.CreatedAt.Format("20060102"), b.storeFileName(article))
	if err := b.store(path, article); err != nil {
		return err
	}
	return nil
}

func (b *Broker) patchArticle(body *PostItem) error {
	if body.ID == "" {
		return errors.New("ID is required")
	}
	u := fmt.Sprintf("api/v2/items/%s", body.ID)
	req, err := b.NewRequest(http.MethodPatch, u, body)
	if err != nil {
		return err
	}

	resp, err := b.do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return errors.New(resp.Status)
	}
	Logf("post", "fresh article ---> %s", body.URL)
	return nil
}

// UploadFresh posts articles to Qiita.
// If an article on Qiita is newer than the one you have locally, we will not update it.
func (b *Broker) UploadFresh(a *Article) (bool, error) {
	ra, err := b.fetchRemoteArticle(a)
	if err != nil {
		return false, err
	}

	if a.Item.UpdatedAt.After(ra.Item.UpdatedAt) == false {
		Logf("", "Article is not updated. remote=%s > local=%s", ra.Item.UpdatedAt, a.Item.UpdatedAt)
		return false, nil
	}

	body := &PostItem{
		Body:    a.Item.Body,
		Private: a.Private,
		Tags:    MarshalTag(a.Tags),
		Title:   a.Title,
		ID:      a.ID,
		URL:     ra.Item.URL,
	}

	if err := b.patchArticle(body); err != nil {
		return false, err
	}

	return true, nil
}

func (b *Broker) storeFileName(a *Article) string {
	var filename string
	switch b.Local.FileNameMode {
	case "title":
		filename = a.Item.Title + defaultExtension
	case "id":
		filename = a.Item.ID + defaultExtension
	default:
		filename = a.Item.Title + defaultExtension
	}
	return filename
}

func (b *Broker) isFileNameModeTitle() bool {
	return b.Local.FileNameMode != "id"
}
