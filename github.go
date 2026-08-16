package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type CommitIndex struct {
	SHA    string     `json:"sha"`
	Commit CommitData `json:"commit"`
}

type CommitIndexData struct {
	Author CommitAuthor `json:"author"`
	URL    string       `json:"url"`
}

type CommitAuthor struct {
	Name  string `json:"name"`
	Email string `json:"email"`
	Date  string `json:"date"`
}

type CommitData struct {
	SHA   string       `json:"sha"`
	Files []CommitFile `json:"files"`
}

type CommitFile struct {
	Filename string `json:"filename"`
	Patch    string `json:"patch"`
}

var client = &http.Client{
	Timeout: 10 * time.Second,
}

func NewGitHubRequest(path string) (*http.Request, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("https://api.github.com/%s"), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", config.Server.GithubToken))

	return req, nil
}

func FetchCommits(repo string) ([]CommitIndex, error) {
	req, err := NewGitHubRequest(fmt.Sprintf("repos/%s/commits", repo))
	if err != nil {
		return nil, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	var list []CommitIndex

	err = json.NewDecoder(resp.Body).Decode(&list)
	if err != nil {
		return nil, err
	}

	return list, nil
}
