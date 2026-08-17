package main

import "strings"

type MarkdownURL struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

func NewMarkdownURLs(addedLines, removedLines []string) []MarkdownURL {
	removedURLs := make(map[string]int)

	for _, line := range removedLines {
		for _, url := range AppendSimpleMarkdownURLs(nil, line) {
			removedURLs[url.URL]++
		}
	}

	urls := make([]MarkdownURL, 0)

	for _, line := range addedLines {
		for _, url := range AppendSimpleMarkdownURLs(nil, line) {
			if removedURLs[url.URL] > 0 {
				removedURLs[url.URL]--

				continue
			}

			urls = append(urls, url)
		}
	}

	return urls
}

func AppendSimpleMarkdownURLs(urls []MarkdownURL, input string) []MarkdownURL {
	var index int

	for index < len(input) {
		if input[index] == '!' && index+1 < len(input) && input[index+1] == '[' {
			index += 2

			continue
		}

		if input[index] != '[' {
			index++

			continue
		}

		index++

		var (
			completed bool
			name      strings.Builder
		)

		for index < len(input) {
			if input[index] == ']' {
				completed = true

				break
			}

			name.WriteByte(input[index])

			index++
		}

		index++

		if !completed || name.Len() == 0 || index >= len(input) || input[index] != '(' {
			continue
		}

		index++

		var url strings.Builder

		for index < len(input) {
			if input[index] == ')' {
				completed = true

				break
			}

			url.WriteByte(input[index])

			index++
		}

		index++

		if !completed || url.Len() == 0 {
			continue
		}

		urls = append(urls, MarkdownURL{
			Name: name.String(),
			URL:  url.String(),
		})
	}

	return urls
}
