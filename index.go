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
	//go:embed index.html
	indexTemplate string

	indexBytes []byte
)

type IndexData struct {
	Feeds    []FeedConfig
	Interval string
}

func RenderIndex() error {
	tmpl, err := template.New("index").Parse(indexTemplate)
	if err != nil {
		return err
	}

	var buffer bytes.Buffer

	err = tmpl.Execute(&buffer, IndexData{
		Feeds:    config.Feeds,
		Interval: FormatInterval(FeedUpdateInterval),
	})

	if err != nil {
		return err
	}

	indexBytes = buffer.Bytes()

	return nil
}

func FormatInterval(d time.Duration) string {
	if d%time.Minute == 0 {
		return fmt.Sprintf("%d minutes", int(d/time.Minute))
	}

	return d.String()
}

func HandleIndex(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	w.Write(indexBytes)
}
