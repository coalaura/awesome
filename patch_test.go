package main

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestParsePatchHunk(t *testing.T) {
	patch := `@@ -10,2 +10,3 @@
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
