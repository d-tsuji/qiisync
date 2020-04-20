package main

import (
	"strconv"
	"strings"
	"time"
)

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

func (item *Item) Date() string {
	return item.CreatedAt.Format("20060102")
}

func (item *Item) AllTags() string {
	tags := make([]string, len(item.Tags))
	for i := range item.Tags {
		tags[i] = strconv.Quote(item.Tags[i].Name)
	}
	return strings.Join(tags, ",")
}
