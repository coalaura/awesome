package main

import (
	"slices"
	"strings"
	"time"
)

type FeedConfig struct {
	Type       string `yaml:"type"`
	Repository string `yaml:"repository"`
	Branch     string `yaml:"branch"`
}

func UpdateFeed(feed FeedConfig) error {
	log.Printf("Updating %s feed...\n", feed.Type)

	sha, err := database.GetLatestCommitSha(feed.Type)
	if err != nil {
		return err
	}

	if sha == "" {
		sha = feed.Branch
	}

	list, err := FetchCommits(feed.Repository, sha)
	if err != nil {
		return err
	}

	slices.Reverse(list) // new2old -> old2new

	for _, commit := range list {
		if commit.SHA == sha {
			continue
		}

		log.Printf("Updating %s/%s...\n", feed.Repository, commit.SHA)

		now := time.Now()

		data, err := FetchCommit(commit.URL)
		if err != nil {
			return err
		}

		message, _ := strings.CutSuffix(commit.Commit.Message, "\n")

		var patch string

		for _, file := range data.Files {
			if file.Filename == "README.md" {
				patch = file.Patch

				break
			}
		}

		var added []MarkdownURL

		if patch != "" {
			hunk, err := ParsePatchHunk(patch)
			if err != nil {
				return err
			}

			added = make([]MarkdownURL, 0, len(hunk.Added))

			for _, line := range hunk.Added {
				added = AppendSimpleMarkdownURLs(added, line)
			}
		} else {
			added = make([]MarkdownURL, 0)
		}

		err = database.AddNewCommit(feed.Type, commit.SHA, commit.User.Login, message, added, commit.Commit.Author.Date, now)
		if err != nil {
			return err
		}
	}

	return nil
}
