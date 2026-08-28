package targets

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/sergi/go-diff/diffmatchpatch"
)

func reconstruct(diffs []diffmatchpatch.Diff, op diffmatchpatch.Operation) string {
	var b strings.Builder
	for _, d := range diffs {
		if d.Type == diffmatchpatch.DiffEqual || d.Type == op {
			b.WriteString(d.Text)
		}
	}
	return b.String()
}

func TestLineDiffsRoundTrip(t *testing.T) {
	tests := []struct {
		name string
		old  string
		new  string
	}{
		{"identical", "a\nb\nc\n", "a\nb\nc\n"},
		{"insert", "a\nb\n", "a\nx\nb\n"},
		{"delete", "a\nx\nb\n", "a\nb\n"},
		{"replace", "one\ntwo\n", "1\ntwo\n"},
		{"empty old", "", "a\nb\n"},
		{"empty new", "a\nb\n", ""},
		{"both empty", "", ""},
		{"no trailing newline", "a\nb", "a\nc"},
		{"blank lines", "a\n\nb\n", "a\nb\n"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			diffs := LineDiffs(tt.old, tt.new)

			if got := reconstruct(diffs, diffmatchpatch.DiffDelete); got != tt.old {
				t.Errorf("old not reconstructed:\n got %q\nwant %q", got, tt.old)
			}
			if got := reconstruct(diffs, diffmatchpatch.DiffInsert); got != tt.new {
				t.Errorf("new not reconstructed:\n got %q\nwant %q", got, tt.new)
			}
		})
	}
}

// go-diff v1.4.0 remapped its line encoding to skip the UTF-16 surrogate
// range. Verify the pipeline still round-trips past 0xD800 (55296) lines.
func TestLineDiffsManyUniqueLines(t *testing.T) {
	if testing.Short() {
		t.Skip("slow")
	}

	const n = 60000
	oldLines := make([]string, n)
	for i := range oldLines {
		oldLines[i] = fmt.Sprintf("line-%d", i)
	}
	old := strings.Join(oldLines, "\n") + "\n"

	newLines := make([]string, 0, n)
	for i := range oldLines {
		if i == n/2 {
			continue // delete one line
		}
		if i == n-1 {
			newLines = append(newLines, "last replaced")
			continue
		}
		newLines = append(newLines, oldLines[i])
	}
	new := strings.Join(newLines, "\n") + "\n"

	diffs := LineDiffs(old, new)

	if got := reconstruct(diffs, diffmatchpatch.DiffDelete); got != old {
		t.Errorf("old not reconstructed past surrogate range")
	}
	if got := reconstruct(diffs, diffmatchpatch.DiffInsert); got != new {
		t.Errorf("new not reconstructed past surrogate range")
	}
}

func TestDiffPrettyText(t *testing.T) {
	diffs := []diffmatchpatch.Diff{
		{Type: diffmatchpatch.DiffEqual, Text: "same\n"},
		{Type: diffmatchpatch.DiffDelete, Text: "gone\n"},
		{Type: diffmatchpatch.DiffInsert, Text: "added\n"},
	}

	want := "\x1b[31mgone\x1b[0m\n\x1b[32madded\x1b[0m\n"
	if got := DiffPrettyText(diffs); got != want {
		t.Errorf("got %q\nwant %q", got, want)
	}
}

func TestDiffFilesMissingPaths(t *testing.T) {
	dir := t.TempDir()
	existing := filepath.Join(dir, "exists")
	if err := os.WriteFile(existing, []byte("a\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	DiffFiles(existing, filepath.Join(dir, "missing"))
	DiffFiles(filepath.Join(dir, "missing"), existing)
}
