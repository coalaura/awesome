package main

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

type PatchHeader struct {
	Old PatchVersion
	New PatchVersion
}

type PatchVersion struct {
	StartLine int64
	Lines     int64
}

type PatchHunk struct {
	Header  PatchHeader
	Added   []string
	Removed []string
}

func ParsePatchHunk(patch string) (PatchHunk, error) {
	var parsed PatchHunk

	var (
		hasHunk  bool
		header   PatchHeader
		oldLines int64
		newLines int64
	)

	finishHunk := func() error {
		if !hasHunk {
			return nil
		}

		if oldLines != header.Old.Lines {
			return fmt.Errorf(
				"old hunk line count mismatch: header says %d, found %d",
				header.Old.Lines,
				oldLines,
			)
		}

		if newLines != header.New.Lines {
			return fmt.Errorf(
				"new hunk line count mismatch: header says %d, found %d",
				header.New.Lines,
				newLines,
			)
		}

		return nil
	}

	for remaining := patch; len(remaining) > 0; {
		line, next, _ := strings.Cut(remaining, "\n")

		line = strings.TrimSuffix(line, "\r")

		remaining = next

		if strings.HasPrefix(line, "@@ ") {
			err := finishHunk()
			if err != nil {
				return PatchHunk{}, err
			}

			parsedHeader, _, err := ParsePatchHeader(line)
			if err != nil {
				return PatchHunk{}, err
			}

			if parsed.Header == (PatchHeader{}) {
				parsed.Header = parsedHeader
			}

			header = parsedHeader
			hasHunk = true
			oldLines = 0
			newLines = 0

			continue
		}

		if !hasHunk {
			return PatchHunk{}, errors.New("invalid header prefix")
		}

		if line == `\ No newline at end of file` {
			continue
		}

		if len(line) == 0 {
			return PatchHunk{}, errors.New("invalid empty hunk line")
		}

		switch line[0] {
		case ' ':
			oldLines++
			newLines++
		case '+':
			newLines++

			parsed.Added = append(parsed.Added, line[1:])
		case '-':
			oldLines++

			parsed.Removed = append(parsed.Removed, line[1:])
		default:
			return PatchHunk{}, fmt.Errorf("invalid hunk line prefix %q", line[0])
		}
	}

	err := finishHunk()
	if err != nil {
		return PatchHunk{}, err
	}

	return parsed, nil
}

func ParsePatchHeader(patch string) (PatchHeader, int, error) {
	if !strings.HasPrefix(patch, "@@ ") {
		return PatchHeader{}, 0, errors.New("invalid header prefix")
	}

	offset := 3

	oldVersion, length, err := ParsePatchVersion(patch[offset:], '-')
	if err != nil {
		return PatchHeader{}, 0, err
	}

	offset += length

	if offset >= len(patch) || patch[offset] != ' ' {
		return PatchHeader{}, 0, errors.New("missing patch version separator")
	}

	offset++

	newVersion, length, err := ParsePatchVersion(patch[offset:], '+')
	if err != nil {
		return PatchHeader{}, 0, err
	}

	offset += length

	if !strings.HasPrefix(patch[offset:], " @@") {
		return PatchHeader{}, 0, errors.New("invalid header suffix")
	}

	offset += 3

	if offset < len(patch) && patch[offset] != ' ' && patch[offset] != '\n' && patch[offset] != '\r' {
		return PatchHeader{}, 0, errors.New("invalid text after header suffix")
	}

	return PatchHeader{
		Old: oldVersion,
		New: newVersion,
	}, offset, nil
}

func ParsePatchVersion(patch string, prefix byte) (PatchVersion, int, error) {
	if len(patch) == 0 || patch[0] != prefix {
		return PatchVersion{}, 0, errors.New("missing patch version prefix")
	}

	offset := 1

	startLine, length, err := parsePatchHeaderNumber(patch[offset:])
	if err != nil {
		return PatchVersion{}, 0, errors.New("invalid patch version starting line")
	}

	offset += length

	lines := int64(1)

	if offset < len(patch) && patch[offset] == ',' {
		offset++

		lines, length, err = parsePatchHeaderNumber(patch[offset:])
		if err != nil {
			return PatchVersion{}, 0, errors.New("invalid patch version lines")
		}

		offset += length
	}

	return PatchVersion{
		StartLine: startLine,
		Lines:     lines,
	}, offset, nil
}

func parsePatchHeaderNumber(value string) (int64, int, error) {
	var length int

	for length < len(value) && value[length] >= '0' && value[length] <= '9' {
		length++
	}

	if length == 0 {
		return 0, 0, errors.New("missing number")
	}

	number, err := strconv.ParseInt(value[:length], 10, 64)
	if err != nil {
		return 0, 0, err
	}

	return number, length, nil
}
