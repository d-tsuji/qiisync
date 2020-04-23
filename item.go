package main

import (
	"strings"
	"time"
)

const defaultDataFormat = "20060102"

// Item is a structure that represents the QiitaAPI.
//
// See also https://qiita.com/api/v2/docs#%E6%8A%95%E7%A8%BF.
type Item struct {
	ID           string    `json:"id"`
	URL          string    `json:"url"`
	User         User      `json:"user"`
	Title        string    `json:"title"`
	Body         string    `json:"body"`
	RenderedBody string    `json:"rendered_body"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	Tags         []*Tag    `json:"tags"`
	Private      bool      `json:"private"`
}

// Tag is a structure that represents the QiitaAPI.
//
// See also https://qiita.com/api/v2/docs#%E6%8A%95%E7%A8%BF.
type Tag struct {
	Name     string   `json:"name"`
	Versions []string `json:"versions"`
}

// User is a structure that represents the QiitaAPI.
//
// See also https://qiita.com/api/v2/docs#%E6%8A%95%E7%A8%BF.
type User struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// PostItem is a structure that represents the Qiita API
// required to post an Article to Qiita.
//
// See also https://qiita.com/api/v2/docs#post-apiv2items.
type PostItem struct {
	Body     string `json:"body"`
	Private  bool   `json:"private"`
	Tags     []*Tag `json:"tags"`
	Title    string `json:"title"`
	ID       string
	URL      string
	FilePath string
}

// PostItemResult is a structure that represents the response body
// when an Article is posted to Qiita.
//
// See also https://qiita.com/api/v2/docs#post-apiv2items.
type PostItemResult struct {
	Body      string    `json:"body"`
	CreatedAt time.Time `json:"created_at"`
	ID        string    `json:"id"`
	Private   bool      `json:"private"`
	Tags      []*Tag    `json:"tags"`
	Title     string    `json:"title"`
	UpdatedAt time.Time `json:"updated_at"`
	URL       string    `json:"url"`
	User      User      `json:"user"`
}

func dateFormat(time time.Time) string {
	return time.Format(defaultDataFormat)
}

func unmarshalTag(Tags []*Tag) string {
	tags := make([]string, len(Tags))
	for i := range Tags {
		tags[i] = Tags[i].Name
		if len(Tags[i].Versions) >= 1 {
			tags[i] += ":" + strings.Join(Tags[i].Versions, ":")
		}
	}
	return strings.Join(tags, ",")
}

func marshalTag(tagString string) []*Tag {
	var tags []*Tag
	for _, v := range strings.Split(tagString, ",") {
		tag := strings.Split(v, ":")
		// Encoding a nil slice into a JSON will result in a null slice,
		// so we define an empty slice.
		version := []string{}
		for i := 1; i < len(tag); i++ {
			version = append(version, tag[i])
		}
		tags = append(tags, &Tag{
			Name:     tag[0],
			Versions: version,
		})
	}
	return tags
}
