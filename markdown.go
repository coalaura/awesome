package main

import (
	"iter"
	"strings"
)

type MarkdownURL struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

func NewMarkdownURLs(addedLines, removedLines []string) []MarkdownURL {
	removed := make([]MarkdownURL, 0, len(removedLines))

	for _, line := range removedLines {
		for url := range FindSimpleMarkdownURLs(line) {
			removed = append(removed, url)
		}
	}

	urls := make([]MarkdownURL, 0, len(addedLines))

	for _, line := range addedLines {
		for added := range FindSimpleMarkdownURLs(line) {
			if i := IndexMatchingMarkdownURL(removed, added); i >= 0 {
				removed = append(removed[:i], removed[i+1:]...)

				continue
			}

			urls = append(urls, added)
		}
	}

	return urls
}

func IndexMatchingMarkdownURL(removed []MarkdownURL, added MarkdownURL) int {
	for i, item := range removed {
		if item.URL == added.URL {
			return i
		}
	}

	for i, item := range removed {
		if item.Name == added.Name {
			return i
		}
	}

	return -1
}

func FindSimpleMarkdownURLs(input string) iter.Seq[MarkdownURL] {
	return func(yield func(MarkdownURL) bool) {
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

			ok := yield(MarkdownURL{
				Name: name.String(),
				URL:  url.String(),
			})

			if !ok {
				return
			}
		}
	}
}
