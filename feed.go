package main

import (
	"context"
	"errors"
	"fmt"
	"html"
	"slices"
	"strings"
	"time"

	"github.com/gorilla/feeds"
)

type FeedConfig struct {
	Type       string `yaml:"type"`
	Repository string `yaml:"repository"`
	Branch     string `yaml:"branch"`
}

const FeedUpdateInterval = 10 * time.Minute

func (f FeedConfig) Build(commits []StoredCommit) *feeds.Feed {
	repoURL := "https://github.com/" + f.Repository

	now := time.Now()

	out := &feeds.Feed{
		Title:       fmt.Sprintf("awesome-%s additions", f.Type),
		Link:        &feeds.Link{Href: repoURL},
		Description: fmt.Sprintf("Newly added links in %s", f.Repository),
		Id:          repoURL,
		Created:     now,
		Updated:     now,
	}

	if len(commits) > 0 && !commits[0].CreatedAt.IsZero() {
		out.Updated = commits[0].CreatedAt
	}

	var n int

	for _, commit := range commits {
		n += len(commit.AddedURLs)
	}

	out.Items = make([]*feeds.Item, 0, n)

	for _, commit := range commits {
		commitURL := repoURL + "/commit/" + commit.SHA

		shortSHA := commit.SHA
		if len(shortSHA) > 7 {
			shortSHA = shortSHA[:7]
		}

		var author *feeds.Author

		if commit.Author != "" {
			author = &feeds.Author{Name: commit.Author}
		}

		for _, added := range commit.AddedURLs {
			title := added.Name
			if title == "" {
				title = added.URL
			}

			out.Items = append(out.Items, &feeds.Item{
				Id:          f.Type + ":" + commit.SHA + ":" + added.URL,
				IsPermaLink: "false",
				Title:       title,
				Link:        &feeds.Link{Href: added.URL},
				Source:      &feeds.Link{Href: commitURL},
				Author:      author,
				Created:     commit.CreatedAt,
				Description: fmt.Sprintf("Added to %s", f.Repository),
				Content:     FeedItemContent(f, commit, added, commitURL, shortSHA),
			})
		}
	}

	return out
}

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

		message, _, _ := strings.Cut(commit.Commit.Message, "\n")

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

func FeedItemContent(f FeedConfig, commit StoredCommit, added MarkdownURL, commitURL, shortSHA string) string {
	var b strings.Builder

	name := added.Name
	if name == "" {
		name = added.URL
	}

	b.WriteString("<p><strong>")
	b.WriteString(html.EscapeString(name))
	b.WriteString("</strong>")

	if added.URL != "" {
		escaped := html.EscapeString(added.URL)

		b.WriteString(" - <a href=\"")
		b.WriteString(escaped)
		b.WriteString("\">")
		b.WriteString(escaped)
		b.WriteString("</a>")
	}

	b.WriteString("</p>\n<p>Added to <a href=\"")
	b.WriteString(html.EscapeString("https://github.com/" + f.Repository))
	b.WriteString("\">")
	b.WriteString(html.EscapeString(f.Repository))
	b.WriteString("</a>")

	if commit.SHA != "" {
		b.WriteString(" in <a href=\"")
		b.WriteString(html.EscapeString(commitURL))
		b.WriteString("\"><code>")
		b.WriteString(html.EscapeString(shortSHA))
		b.WriteString("</code></a>")
	}

	if commit.Author != "" {
		b.WriteString(" by <a href=\"")
		b.WriteString(html.EscapeString("https://github.com/" + commit.Author))
		b.WriteString("\">")
		b.WriteString(html.EscapeString(commit.Author))
		b.WriteString("</a>")
	}

	b.WriteString(".</p>")

	if commit.Message != "" {
		b.WriteString("\n<pre>")
		b.WriteString(html.EscapeString(commit.Message))
		b.WriteString("</pre>")
	}

	return b.String()
}
