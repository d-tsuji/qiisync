package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/urfave/cli/v2"
)

var errCommandHelp = fmt.Errorf("command help shown")

func main() {
	app := cli.NewApp()
	app.Commands = []*cli.Command{
		commandPull,
		//commandPush,
		commandPost,
		//commandList,
	}
	app.Version = fmt.Sprintf("%s (%s)", version, revision)
	err := app.Run(os.Args)
	if err != nil {
		if err != errCommandHelp {
			logf("error", "%+v", err)
		}
	}
}

var commandPull = &cli.Command{
	Name:  "pull",
	Usage: "Pull entries from remote",
	Action: func(c *cli.Context) error {
		conf, err := loadConfiguration()
		if err != nil {
			return fmt.Errorf("load config: %w", err)
		}
		b := NewBroker(conf)
		remoteArticles, err := b.fetchRemoteArticles()
		if err != nil {
			return err
		}
		localArticles, err := b.fetchLocalArticles()
		if err != nil {
			return err
		}
		for i := range remoteArticles {
			if _, err := b.StoreFresh(localArticles, remoteArticles[i]); err != nil {
				return err
			}
		}
		return nil
	},
}

var commandPost = &cli.Command{
	Name:  "post",
	Usage: "Post a new entry to remote",
	Flags: []cli.Flag{
		&cli.StringFlag{Name: "path"},
		&cli.StringFlag{Name: "title"},
		&cli.StringFlag{Name: "tag"},
		&cli.BoolFlag{Name: "private"},
	},
	Action: func(c *cli.Context) error {
		blog := c.Args().First()
		if blog == "" {
			cli.ShowCommandHelp(c, "post")
			return errCommandHelp
		}

		conf, err := loadConfiguration()
		if err != nil {
			return err
		}

		private := c.Bool("private")

		title := c.String("title")
		if title == "" {
			return fmt.Errorf("title is required")
		}

		tag := c.String("tag")
		if tag == "" {
			return fmt.Errorf("one or more tag is required")
		}

		path := c.String("path")
		if path == "" {
			return fmt.Errorf("path is required")
		}

		a, err := articleFromFile(path)
		if err != nil {
			return err
		}

		post := &PostItem{
			Body:    a.Item.Body,
			Private: private,
			Tags:    MarshalTag(tag),
			Title:   title,
		}

		b := NewBroker(conf)
		err = b.PostArticle(post)
		if err != nil {
			return err
		}
		return nil
	},
}

func loadConfiguration() (*config, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	fname := filepath.Join(home, ".config", "qsync", "config")
	f, err := os.Open(fname)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	conf, err := loadConfig(f)
	if err != nil {
		return nil, err
	}
	return conf, nil
}
