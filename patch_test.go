package main

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestNewMarkdownURLs(t *testing.T) {
	added := []string{
		"- [Second](https://example.com/second)",
		"- [First](https://example.com/first)",
		"- [New](https://example.com/new)",
	}

	removed := []string{
		"- [First](https://example.com/first)",
		"- [Second](https://example.com/second)",
	}

	got := NewMarkdownURLs(added, removed)

	want := []MarkdownURL{
		{
			Name: "New",
			URL:  "https://example.com/new",
		},
	}

	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("NewMarkdownURLs() mismatch (-want +got):\n%s", diff)
	}
}

func TestParsePatchHunkMultipleHeaders(t *testing.T) {
	patch := `@@ -1,1 +1,1 @@ example header
- [First](https://example.com/first)
+ [Second](https://example.com/second)
@@ -10,1 +10,2 @@
 [Unchanged](https://example.com/unchanged)
+ [New](https://example.com/new)
`

	got, err := ParsePatchHunk(patch)
	if err != nil {
		t.Fatalf("ParsePatchHunk() error = %v", err)
	}

	want := PatchHunk{
		Header: PatchHeader{
			Old: PatchVersion{
				StartLine: 1,
				Lines:     1,
			},
			New: PatchVersion{
				StartLine: 1,
				Lines:     1,
			},
		},
		Added: []string{
			" [Second](https://example.com/second)",
			" [New](https://example.com/new)",
		},
		Removed: []string{
			" [First](https://example.com/first)",
		},
	}

	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("ParsePatchHunk() mismatch (-want +got):\n%s", diff)
	}
}

func TestParsePatchHunk(t *testing.T) {
	patch := `@@ -10,2 +10,3 @@ example header
 unchanged
-removed
+added
+another added
`

	got, err := ParsePatchHunk(patch)
	if err != nil {
		t.Fatalf("ParsePatchHunk() error = %v", err)
	}

	want := PatchHunk{
		Header: PatchHeader{
			Old: PatchVersion{
				StartLine: 10,
				Lines:     2,
			},
			New: PatchVersion{
				StartLine: 10,
				Lines:     3,
			},
		},
		Added: []string{
			"added",
			"another added",
		},
		Removed: []string{
			"removed",
		},
	}

	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("ParsePatchHunk() mismatch (-want +got):\n%s", diff)
	}
}
