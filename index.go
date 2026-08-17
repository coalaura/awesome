package main

import (
	"bytes"
	_ "embed"
	"fmt"
	"html/template"
	"net/http"
	"time"
)

var (
	//go:embed html/index.html
	indexTemplate string

	//go:embed html/feed.html
	feedTemplate string

	//go:embed html/style.css
	stylesheet []byte

	indexBytes []byte
	feedPage   *template.Template
)

type IndexData struct {
	Feeds    []FeedConfig
	Interval string
}

type FeedIndexData struct {
	Feed     FeedConfig
	Entries  []FeedEntry
	Interval string
}

type FeedEntry struct {
	Name      string
	URL       string
	CommitURL string
	ShortSHA  string
	Author    string
	AuthorURL string
	Message   string
	CreatedAt time.Time
}

func RenderIndex() error {
	index, err := template.New("index").Parse(indexTemplate)
	if err != nil {
		return err
	}

	feed, err := template.New("feed").Funcs(template.FuncMap{
		"FormatDate": FormatDate,
	}).Parse(feedTemplate)
	if err != nil {
		return err
	}

	var buffer bytes.Buffer

	err = index.Execute(&buffer, IndexData{
		Feeds:    config.Feeds,
		Interval: FormatInterval(FeedUpdateInterval),
	})
	if err != nil {
		return err
	}

	indexBytes = buffer.Bytes()
	feedPage = feed

	return nil
}

func RenderFeed(feed FeedConfig, commits []StoredCommit) ([]byte, error) {
	var n int

	for _, commit := range commits {
		n += len(commit.AddedURLs)
	}

	entries := make([]FeedEntry, 0, n)

	for _, commit := range commits {
		commitURL := "https://github.com/" + feed.Repository + "/commit/" + commit.SHA

		shortSHA := commit.SHA
		if len(shortSHA) > 7 {
			shortSHA = shortSHA[:7]
		}

		for _, added := range commit.AddedURLs {
			name := added.Name
			if name == "" {
				name = added.URL
			}

			entries = append(entries, FeedEntry{
				Name:      name,
				URL:       added.URL,
				CommitURL: commitURL,
				ShortSHA:  shortSHA,
				Author:    commit.Author,
				AuthorURL: "https://github.com/" + commit.Author,
				Message:   commit.Message,
				CreatedAt: commit.CreatedAt,
			})
		}
	}

	var buffer bytes.Buffer

	err := feedPage.Execute(&buffer, FeedIndexData{
		Feed:     feed,
		Entries:  entries,
		Interval: FormatInterval(FeedUpdateInterval),
	})
	if err != nil {
		return nil, err
	}

	return buffer.Bytes(), nil
}

func HandleIndex(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	w.Write(indexBytes)
}

func HandleFeedIndex(w http.ResponseWriter, r *http.Request) {
	feed, ok := GetFeed(r)
	if !ok {
		w.WriteHeader(http.StatusBadRequest)

		return
	}

	commits, err := database.GetCommitsByType(r.Context(), feed.Type)
	if err != nil {
		log.Warnf("%s: %v\n", feed.Type, err)

		w.WriteHeader(http.StatusInternalServerError)

		return
	}

	page, err := RenderFeed(feed, commits)
	if err != nil {
		log.Warnf("%s: %v\n", feed.Type, err)

		w.WriteHeader(http.StatusInternalServerError)

		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	w.Write(page)
}

func HandleStylesheet(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/css; charset=utf-8")

	w.Write(stylesheet)
}

func FormatInterval(d time.Duration) string {
	if d%time.Minute == 0 {
		return fmt.Sprintf("%d minutes", int(d/time.Minute))
	}

	return d.String()
}

func FormatDate(d time.Time) string {
	return d.Format("02 Jan 2006")
}
