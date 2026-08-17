package main

import (
	"context"
	"errors"
	"slices"
	"strings"
	"time"
)

type FeedConfig struct {
	Type       string `yaml:"type"`
	Repository string `yaml:"repository"`
	Branch     string `yaml:"branch"`
}

const FeedUpdateInterval = 10 * time.Minute

func RunFeedUpdates(ctx context.Context, feeds []FeedConfig) {
	updateAll := func() bool {
		for _, feed := range feeds {
			if err := ctx.Err(); err != nil {
				return false
			}

			err := UpdateFeed(ctx, feed)
			if err != nil {
				if errors.Is(err, context.Canceled) {
					return false
				}

				log.Warnln(err)
			}
		}

		return true
	}

	if !updateAll() {
		return
	}

	ticker := time.NewTicker(FeedUpdateInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if !updateAll() {
				return
			}
		}
	}
}

func UpdateFeed(ctx context.Context, feed FeedConfig) error {
	log.Printf("Updating %s feed...\n", feed.Type)

	latestSHA, err := database.GetLatestCommitSha(feed.Type)
	if err != nil {
		return err
	}

	list, err := FetchCommits(ctx, feed.Repository, feed.Branch)
	if err != nil {
		return err
	}

	if latestSHA != "" {
		for index, commit := range list {
			if commit.SHA == latestSHA {
				list = list[:index]

				break
			}
		}
	}

	slices.Reverse(list) // new2old -> old2new

	for _, commit := range list {
		log.Printf("Updating %s/%s...\n", feed.Repository, commit.SHA)

		now := time.Now()

		data, err := FetchCommit(ctx, commit.URL)
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

		added := make([]MarkdownURL, 0)

		if patch != "" {
			hunk, err := ParsePatchHunk(patch)
			if err != nil {
				log.Println(patch)

				return err
			}

			added = NewMarkdownURLs(hunk.Added, hunk.Removed)
		}

		err = database.AddNewCommit(feed.Type, commit.SHA, commit.User.Login, message, added, commit.Commit.Author.Date, now)
		if err != nil {
			return err
		}
	}

	log.Printf("Finished %s feed update\n", feed.Type)

	return nil
}
