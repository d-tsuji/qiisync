package main

import (
	"strings"
	"time"
)

const defaultDataFormat = "20060102"

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

type Tag struct {
	Name     string   `json:"name"`
	Versions []string `json:"versions"`
}

type User struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type PostItem struct {
	Body     string `json:"body"`
	Private  bool   `json:"private"`
	Tags     []*Tag `json:"tags"`
	Title    string `json:"title"`
	ID       string
	URL      string
	FilePath string
}

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

func DateFormat(time time.Time) string {
	return time.Format(defaultDataFormat)
}

func UnmarshalTag(Tags []*Tag) string {
	tags := make([]string, len(Tags))
	for i := range Tags {
		tags[i] = Tags[i].Name
		if len(Tags[i].Versions) >= 1 {
			tags[i] += ":" + strings.Join(Tags[i].Versions, ":")
		}
	}
	return strings.Join(tags, ",")
}

func MarshalTag(tagString string) []*Tag {
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
