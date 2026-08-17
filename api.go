package main

import (
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
)

func HandleFeed(w http.ResponseWriter, r *http.Request) {
	feed, ok := GetFeed(r)
	if !ok {
		w.WriteHeader(http.StatusBadRequest)

		return
	}

	ctx := r.Context()

	commits, err := database.GetCommitsByType(ctx, feed.Type)
	if err != nil {
		log.Warnf("%s: %v\n", feed, err)

		w.WriteHeader(http.StatusInternalServerError)

		return
	}

	built := feed.Build(commits)

	format := chi.URLParam(r, "format")

	switch format {
	case "atom":
		w.Header().Set("Content-Type", "application/atom+xml; charset=utf-8")

		err = built.WriteAtom(w)
	case "json":
		w.Header().Set("Content-Type", "application/feed+json; charset=utf-8")

		err = built.WriteJSON(w)
	default:
		w.Header().Set("Content-Type", "application/rss+xml; charset=utf-8")

		err = built.WriteRss(w)
	}

	if err != nil {
		log.Warnf("%s: %v\n", feed, err)
	}
}

func GetFeed(r *http.Request) (FeedConfig, bool) {
	name := chi.URLParam(r, "feed")

	for _, feed := range config.Feeds {
		if strings.EqualFold(feed.Type, name) {
			return feed, true
		}
	}

	return FeedConfig{}, false
}
