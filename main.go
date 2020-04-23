package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/urfave/cli/v2"
)

var errCommandHelp = fmt.Errorf("command help shown")

func main() {
	app := cli.NewApp()
	app.Commands = []*cli.Command{
		commandPull,
		commandPost,
		commandUpdate,
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
	Usage: "Pull articles from remote",
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
	Usage: "Post a new article to remote",
	Action: func(c *cli.Context) error {
		filename := c.Args().First()
		if filename == "" {
			cli.ShowCommandHelp(c, "post")
			return errCommandHelp
		}

		conf, err := loadConfiguration()
		if err != nil {
			return err
		}

		// Receives parameters from the stdin.
		sc := bufio.NewScanner(os.Stdin)

		fmt.Fprintln(os.Stdout, "")
		fmt.Fprintln(os.Stdout, `Please enter the "title" of the article you want to post.`)
		_ = sc.Scan()
		title := sc.Text()
		if title == "" {
			return fmt.Errorf("title is required")
		}

		fmt.Fprintln(os.Stdout, "")
		fmt.Fprintln(os.Stdout, `Please enter the "tag" of the article you want to post.`)
		fmt.Fprintln(os.Stdout, `Tag is like "React,redux,TypeScript" or "Go" or "Python:3.7". To specify more than one, separate them with ",".`)
		_ = sc.Scan()
		tag := sc.Text()
		if tag == "" {
			return fmt.Errorf("more than one tag is required")
		}

		fmt.Fprintln(os.Stdout, "")
		fmt.Fprintln(os.Stdout, `Do you make the article you post public? "true" is public, "false" is private.`)
		_ = sc.Scan()
		text := sc.Text()
		private, err := strconv.ParseBool(text)
		if err != nil {
			return fmt.Errorf("input string (%s) could not be parsed into bool", text)
		}

		a, err := articleFromFile(filename)
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
	filename := filepath.Join(home, ".config", "qiisync", "config")
	f, err := os.Open(filename)
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

var commandUpdate = &cli.Command{
	Name:  "update",
	Usage: "push local article to remote",
	Action: func(c *cli.Context) error {
		filename := c.Args().First()
		if filename == "" {
			cli.ShowCommandHelp(c, "update")
			return errCommandHelp
		}

		conf, err := loadConfiguration()
		if err != nil {
			return err
		}

		a, err := articleFromFile(filename)
		if err != nil {
			return err
		}

		if a.Private {
			return errors.New("Once published, an article cannot be made a private publication.\n" +
				"\tPlease check if the Private item in the header of the article is set to false.")
		}

		b := NewBroker(conf)
		_, err = b.UploadFresh(a)
		if err != nil {
			return err
		}
		return nil
	},
}
