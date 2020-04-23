package qiisync

import (
	"io"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

// Config stores Qiita's configuration and local environment settings.
type Config struct {
	Qiita qiitaConfig `toml:"qiita"`
	Local localConfig `toml:"local"`
}

type qiitaConfig struct {
	Token string `toml:"api_token"`
}

type localConfig struct {
	Dir          string `toml:"base_dir"`
	FileNameMode string `toml:"filename_mode"`
}

// LoadConfiguration gets its configuration from "~/.config/qiisync/config".
func LoadConfiguration() (*Config, error) {
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
	return loadConfig(f)
}

func loadConfig(r io.Reader) (*Config, error) {
	var config Config
	if _, err := toml.DecodeReader(r, &config); err != nil {
		return nil, err
	}
	return &config, nil
}

func (c *Config) baseDir() string {
	return c.Local.Dir
}
