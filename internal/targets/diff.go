package targets

import (
	"bytes"
	"fmt"
	"os"
	"strings"

	"github.com/sergi/go-diff/diffmatchpatch"
)

// DiffFiles takes two file paths and compares them line by line
func DiffFiles(path1, path2 string) {
	old, err := os.ReadFile(path1)
	if err != nil {
		old = []byte{}
	}

	new, err := os.ReadFile(path2)
	if err != nil {
		new = []byte{}
	}

	Diff(string(old), string(new))
}

// Diff takes two strings and compares them line by line
func Diff(old, new string) {
	dmp := diffmatchpatch.New()

	fileAdmp, fileBdmp, dmpStrings := dmp.DiffLinesToChars(old, new)
	diffs := dmp.DiffMain(fileAdmp, fileBdmp, false)
	diffs = dmp.DiffCharsToLines(diffs, dmpStrings)
	diffs = dmp.DiffCleanupSemantic(diffs)

	fmt.Println(DiffPrettyText(diffs))
}

// DiffPrettyText converts a []Diff into a colored text report.
func DiffPrettyText(diffs []diffmatchpatch.Diff) string {
	var buff bytes.Buffer
	for _, diff := range diffs {
		text := diff.Text

		switch diff.Type {
		case diffmatchpatch.DiffInsert:
			_, _ = buff.WriteString("\x1b[32m")
			_, _ = buff.WriteString(strings.Trim(text, "\n"))
			_, _ = buff.WriteString("\x1b[0m")
			_, _ = buff.WriteString("\n")
		case diffmatchpatch.DiffDelete:
			_, _ = buff.WriteString("\x1b[31m")
			_, _ = buff.WriteString(strings.Trim(text, "\n"))
			_, _ = buff.WriteString("\x1b[0m")
			_, _ = buff.WriteString("\n")
		case diffmatchpatch.DiffEqual:
			// _, _ = buff.WriteString(text)
		}

	}

	return buff.String()
}
