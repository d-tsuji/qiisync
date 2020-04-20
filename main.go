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
		//commandPost,
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

func loadConfiguration() (*config, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	fname := filepath.Join(home, ".config", "qsync", "config")
	file, err := os.Open(fname)
	if err != nil {
		return nil, err
	}
	conf, err := loadConfig(file)
	if err != nil {
		return nil, err
	}
	return conf, nil
}
