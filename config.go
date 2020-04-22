package main

import (
	"io"

	"github.com/BurntSushi/toml"
)

type config struct {
	Qiita qiitaConfig `toml:"qiita"`
	Local localConfig `toml:"local"`
}

type qiitaConfig struct {
	Token string `toml:"api_token"`
}

type localConfig struct {
	Dir string `toml:"base_dir"`
}

func loadConfig(r io.Reader) (*config, error) {
	var config config
	if _, err := toml.DecodeReader(r, &config); err != nil {
		return nil, err
	}
	return &config, nil
}

func (c *config) BaseDir() string {
	return c.Local.Dir
}
