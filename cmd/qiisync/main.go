package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"

	"github.com/d-tsuji/qiisync"
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
	app.Version = fmt.Sprintf("%s", qiisync.Version)
	err := app.Run(os.Args)
	if err != nil {
		if err != errCommandHelp {
			qiisync.Logf("error", "%v", err)
		}
	}
}

var commandPull = &cli.Command{
	Name:  "pull",
	Usage: "Pull articles from remote",
	Action: func(c *cli.Context) error {
		conf, err := qiisync.LoadConfiguration()
		if err != nil {
			return fmt.Errorf("load config: %w", err)
		}
		b := qiisync.NewBroker(conf)
		remoteArticles, err := b.FetchRemoteArticles()
		if err != nil {
			return err
		}
		localArticles, err := b.FetchLocalArticles()
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
	Usage: "Post a new Article to remote",
	Action: func(c *cli.Context) error {
		filename := c.Args().First()
		if filename == "" {
			cli.ShowCommandHelp(c, "post")
			return errCommandHelp
		}

		conf, err := qiisync.LoadConfiguration()
		if err != nil {
			return err
		}

		// Receives parameters from the stdin.
		sc := bufio.NewScanner(os.Stdin)
		var (
			title   string
			tag     string
			private bool
		)

		fmt.Fprintln(os.Stdout, "")
		fmt.Fprintln(os.Stdout, `Please enter the "title" of the Article you want to post.`)
		if sc.Scan() {
			title = sc.Text()
			if title == "" {
				return fmt.Errorf("title is required")
			}
		}
		if err := sc.Err(); err != nil {
			return fmt.Errorf("an unexpected error has occurred when scanning: %w", err)
		}

		fmt.Fprintln(os.Stdout, "")
		fmt.Fprintln(os.Stdout, `Please enter the "tag" of the Article you want to post.`)
		fmt.Fprintln(os.Stdout, `Tag is like "React,redux,TypeScript" or "Go" or "Python:3.7". To specify more than one, separate them with ",".`)
		if sc.Scan() {
			tag = sc.Text()
			if tag == "" {
				return fmt.Errorf("more than one tag is required")
			}
		}
		if err := sc.Err(); err != nil {
			return fmt.Errorf("an unexpected error has occurred when scanning: %w", err)
		}

		fmt.Fprintln(os.Stdout, "")
		fmt.Fprintln(os.Stdout, `Do you make the Article you post private? "true" is private, "false" is public.`)
		if sc.Scan() {
			text := sc.Text()
			private, err = strconv.ParseBool(text)
			if err != nil {
				return fmt.Errorf("input string (%s) could not be parsed into bool", text)
			}
		}
		if err := sc.Err(); err != nil {
			return fmt.Errorf("an unexpected error has occurred when scanning: %w", err)
		}

		a, err := qiisync.ArticleFromFile(filename)
		if err != nil {
			return err
		}

		post := &qiisync.PostItem{
			Body:    a.Item.Body,
			Private: private,
			Tags:    qiisync.MarshalTag(tag),
			Title:   title,
		}

		b := qiisync.NewBroker(conf)
		err = b.PostArticle(post)
		if err != nil {
			return err
		}
		return nil
	},
}

var commandUpdate = &cli.Command{
	Name:  "update",
	Usage: "Push local Article to remote",
	Action: func(c *cli.Context) error {
		filename := c.Args().First()
		if filename == "" {
			cli.ShowCommandHelp(c, "update")
			return errCommandHelp
		}

		conf, err := qiisync.LoadConfiguration()
		if err != nil {
			return err
		}

		a, err := qiisync.ArticleFromFile(filename)
		if err != nil {
			return err
		}

		b := qiisync.NewBroker(conf)
		_, err = b.UploadFresh(a)
		if err != nil {
			return err
		}
		return nil
	},
}
