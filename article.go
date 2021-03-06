package qiisync

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"regexp"
	"strings"

	"gopkg.in/yaml.v2"
)

var delimReg = regexp.MustCompile(`---\n+`)

// ArticleHeader is a structure that represents the metadata of a Article.
type ArticleHeader struct {
	ID      string `yaml:"ID"`
	Title   string `yaml:"Title"`
	Tags    string `yaml:"Tags"`
	Author  string `yaml:"Author"`
	Private bool   `yaml:"Private"`
}

// Article is a structure that holds the metadata of a file and the contents of an article.
type Article struct {
	*ArticleHeader
	Item     *Item
	FilePath string
}

func (a *Article) headerString() (string, error) {
	d, err := yaml.Marshal(a.ArticleHeader)
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	headers := []string{
		"---",
		string(d),
	}
	return strings.Join(headers, "\n") + "---\n\n", nil
}

func (a *Article) fullContent() (string, error) {
	header, err := a.headerString()
	if err != nil {
		return "", nil
	}
	c := header + a.Item.Body
	if !strings.HasSuffix(c, "\n") {
		// fill newline for suppressing diff "No newline at end of file"
		c += "\n"
	}
	return c, nil
}

// ArticleFromFile extracts an article from local filesysytem.
func ArticleFromFile(filepath string) (*Article, error) {
	b, err := ioutil.ReadFile(filepath)
	if err != nil {
		return nil, err
	}
	content := string(b)
	isNew := !strings.HasPrefix(content, "---\n")
	ah := ArticleHeader{}
	if !isNew {
		c := delimReg.Split(content, 3)
		if len(c) != 3 || c[0] != "" {
			return nil, fmt.Errorf("article format is invalid")
		}

		if err := yaml.Unmarshal([]byte(c[1]), &ah); err != nil {
			return nil, err
		}
		content = c[2]
	}
	a := &Article{
		ArticleHeader: &ah,
		Item:          &Item{Body: content},
		FilePath:      filepath,
	}

	fi, err := os.Stat(filepath)
	if err != nil {
		return nil, err
	}
	a.Item.UpdatedAt = fi.ModTime()

	return a, nil
}
