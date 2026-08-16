package main

import (
	"bytes"
	"errors"
	"fmt"
	"os"

	"github.com/goccy/go-yaml"
)

const ConfigPath = "config.yml"

type ServerConfig struct {
	Port        int    `yaml:"port"`
	GithubToken string `yaml:"github_token"`
}

type Config struct {
	Server ServerConfig `yaml:"server"`
	Feeds  []FeedConfig `yaml:"feeds"`
}

func LoadConfig() (*Config, error) {
	// defaults
	cfg := &Config{
		Server: ServerConfig{
			Port: 5883,
		},
		Feeds: []FeedConfig{
			{
				Type:       "go",
				Repository: "avelino/awesome-go",
				Branch:     "main",
			},
		},
	}

	file, err := os.OpenFile(ConfigPath, os.O_RDONLY, 0)
	if err == nil {
		defer file.Close()

		err = yaml.NewDecoder(file).Decode(cfg)
		if err != nil {
			return nil, err
		}
	} else {
		if !os.IsNotExist(err) {
			return nil, err
		}

		err = cfg.Store()
		if err != nil {
			return nil, err
		}
	}

	err = cfg.Validate()
	if err != nil {
		return nil, err
	}

	return cfg, nil
}

func (c *Config) Addr() string {
	return fmt.Sprintf(":%d", c.Server.Port)
}

func (c *Config) Validate() error {
	if c.Server.Port < 1 || c.Server.Port > 65535 {
		return fmt.Errorf("invalid server.port %d", c.Server.Port)
	}

	if c.Server.GithubToken == "" {
		return errors.New("missing server.github_token")
	}

	return nil
}

func (c *Config) Store() error {
	var (
		buffer   bytes.Buffer
		comments = yaml.CommentMap{
			"$.feeds": {yaml.HeadComment("\n# list of feeds to pull from and provide as rss")},

			"$.server.port":         {yaml.HeadComment(" port to serve on (required; default 5883)")},
			"$.server.github_token": {yaml.HeadComment(" github api token (required)")},
		}
	)

	err := yaml.NewEncoder(&buffer, yaml.WithComment(comments)).Encode(c)
	if err != nil {
		return err
	}

	body := bytes.ReplaceAll(buffer.Bytes(), []byte("#\n"), []byte("\n"))

	return os.WriteFile(ConfigPath, body, 0644)
}
