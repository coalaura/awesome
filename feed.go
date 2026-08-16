package main

import "net/http"

type FeedConfig struct {
	Type       string `yaml:"type"`
	Repository string `yaml:"repository"`
	Branch     string `yaml:"branch"`
}

func HandleFeed(w http.ResponseWriter, r *http.Request) {}
