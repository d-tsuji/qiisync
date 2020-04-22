package main

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
	defaultBaseURL      = "https://qiita.com/"
	defaultItemsPerPage = 20
	defaultExtension    = ".md"
)

type Broker struct {
	*config
	BaseURL *url.URL
}

func NewBroker(c *config) *Broker {
	baseURL, _ := url.Parse(defaultBaseURL)
	return &Broker{
		config:  c,
		BaseURL: baseURL,
	}
}

func (b *Broker) do(req *http.Request) (*http.Response, error) {
	return http.DefaultClient.Do(req)
}

func (b *Broker) fetchRemoteArticle(a *article) (*article, error) {
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

func (b *Broker) fetchRemoteArticles() ([]*article, error) {
	var articles []*article
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

func (b *Broker) fetchRemoteItemsPerPage(page int) ([]*article, bool, error) {
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

func (b *Broker) fetchLocalArticles() (articles map[string]*article, err error) {
	articles = make(map[string]*article)
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("recoverd when dirwalk(%s): %v", b.BaseDir(), r)
		}
	}()
	fnameList := dirwalk(b.BaseDir())
	for i := range fnameList {
		a, err := articleFromFile(fnameList[i])
		if err != nil {
			return nil, err
		}
		// If ArticleHeader.ID is empty, it just indicates a new file.
		if a.ArticleHeader.ID == "" {
			continue
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

func (b *Broker) LocalPath(article *article) string {
	extension := ".md"
	paths := []string{b.BaseDir()}
	paths = append(paths, DateFormat(article.Item.CreatedAt), article.ID+extension)
	return filepath.Join(paths...)
}

func (b *Broker) StoreFresh(localArticles map[string]*article, remoteArticle *article) (bool, error) {
	var localLastModified time.Time
	path := filepath.Join(b.BaseDir(), DateFormat(remoteArticle.Item.CreatedAt), remoteArticle.ID+defaultExtension)

	a, exists := localArticles[remoteArticle.ID]
	if exists {
		localLastModified = a.Item.UpdatedAt
		path = a.FilePath
	}
	if remoteArticle.Item.UpdatedAt.After(localLastModified) {
		logf("fresh", "remote=%s > local=%s", remoteArticle.Item.UpdatedAt, localLastModified)
		if err := b.Store(path, remoteArticle); err != nil {
			return false, err
		}
		return true, nil
	}

	return false, nil
}

func (b *Broker) Store(path string, article *article) error {
	logf("store", "%s", path)

	dir, _ := filepath.Split(path)
	err := os.MkdirAll(dir, 0755)
	if err != nil {
		return err
	}

	f, err := os.Create(path)
	if err != nil {
		return err
	}

	fullContext, err := article.FullContent()
	if err != nil {
		return err
	}
	_, err = f.WriteString(fullContext)
	if err != nil {
		return err
	}

	if err := f.Close(); err != nil {
		return err
	}

	return os.Chtimes(path, article.Item.UpdatedAt, article.Item.UpdatedAt)
}

func (b *Broker) convertItemsArticles(items []*Item) []*article {
	articles := make([]*article, len(items))
	for i := range items {
		articles[i] = b.convertItemsArticle(items[i])
	}
	return articles
}

func (b *Broker) convertItemsArticle(item *Item) *article {
	return &article{
		ArticleHeader: &ArticleHeader{
			ID:      item.ID,
			Title:   item.Title,
			Tags:    UnmarshalTag(item.Tags),
			Author:  item.User.Name,
			Private: item.Private,
		},
		Item: item,
	}
}

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
	logf("post", "URL: %s", r.URL)

	article := &article{
		ArticleHeader: &ArticleHeader{
			ID:      r.ID,
			Title:   r.Title,
			Tags:    UnmarshalTag(r.Tags),
			Author:  r.User.Name,
			Private: r.Private,
		},
		Item: &Item{
			Body:      r.Body,
			CreatedAt: r.CreatedAt,
			UpdatedAt: r.UpdatedAt,
		},
	}

	path := filepath.Join(b.BaseDir(), r.CreatedAt.Format("20060102"), article.ID+defaultExtension)
	if err := b.Store(path, article); err != nil {
		return err
	}
	return nil
}

func (b *Broker) UploadFresh(a *article) (bool, error) {
	ra, err := b.fetchRemoteArticle(a)
	if err != nil {
		return false, err
	}

	if a.Item.UpdatedAt.After(ra.Item.UpdatedAt) == false {
		logf("", "article is not uploaded, remote=%s > local=%s", ra.Item.UpdatedAt, a.Item.UpdatedAt)
		return false, nil
	}

	body := &PostItem{
		Body:    a.Item.Body,
		Private: a.Private,
		Tags:    MarshalTag(a.Tags),
		Title:   a.Title,
	}
	for _, v := range MarshalTag(a.Tags) {
		fmt.Printf("%#v\n", v)
	}

	u := fmt.Sprintf("api/v2/items/%s", a.ID)
	req, err := b.NewRequest(http.MethodPatch, u, body)
	if err != nil {
		return false, err
	}

	resp, err := b.do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return false, errors.New(resp.Status)
	}

	return true, nil
}
