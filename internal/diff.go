package internal

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/sergi/go-diff/diffmatchpatch"
	"golang.org/x/tools/txtar"
)

type DiffOptions struct {
	Left   string
	Right  string
	IsDir  bool
}

type FileDiff struct {
	Path      string
	Status    string
	LeftData  []byte
	RightData []byte
}

func Diff(opts DiffOptions) ([]FileDiff, error) {
	var leftArchive, rightArchive *txtar.Archive
	var err error

	if opts.IsDir {
		leftArchive, err = dirToArchive(opts.Left)
		if err != nil {
			return nil, fmt.Errorf("failed to read directory %q: %w", opts.Left, err)
		}

		rightData, err := os.ReadFile(opts.Right)
		if err != nil {
			return nil, fmt.Errorf("failed to read archive %q: %w", opts.Right, err)
		}
		rightArchive = txtar.Parse(rightData)
	} else {
		leftData, err := os.ReadFile(opts.Left)
		if err != nil {
			return nil, fmt.Errorf("failed to read archive %q: %w", opts.Left, err)
		}
		leftArchive = txtar.Parse(leftData)

		rightData, err := os.ReadFile(opts.Right)
		if err != nil {
			return nil, fmt.Errorf("failed to read archive %q: %w", opts.Right, err)
		}
		rightArchive = txtar.Parse(rightData)
	}

	return compareArchives(leftArchive, rightArchive), nil
}

func dirToArchive(dir string) (*txtar.Archive, error) {
	archive := &txtar.Archive{}

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		relPath, err := filepath.Rel(dir, path)
		if err != nil {
			return err
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		archive.Files = append(archive.Files, txtar.File{
			Name: relPath,
			Data: content,
		})

		return nil
	})

	return archive, err
}

func compareArchives(left, right *txtar.Archive) []FileDiff {
	var diffs []FileDiff

	leftMap := make(map[string][]byte)
	rightMap := make(map[string][]byte)

	for _, f := range left.Files {
		leftMap[f.Name] = f.Data
	}

	for _, f := range right.Files {
		rightMap[f.Name] = f.Data
	}

	for name, leftData := range leftMap {
		if rightData, exists := rightMap[name]; exists {
			if !bytes.Equal(leftData, rightData) {
				diffs = append(diffs, FileDiff{
					Path:      name,
					Status:    "modified",
					LeftData:  leftData,
					RightData: rightData,
				})
			}
		} else {
			diffs = append(diffs, FileDiff{
				Path:     name,
				Status:   "deleted",
				LeftData: leftData,
			})
		}
	}

	for name, rightData := range rightMap {
		if _, exists := leftMap[name]; !exists {
			diffs = append(diffs, FileDiff{
				Path:      name,
				Status:    "added",
				RightData: rightData,
			})
		}
	}

	return diffs
}

func PrintDiff(w io.Writer, diff FileDiff, showContent bool) {
	switch diff.Status {
	case "added":
		fmt.Fprintf(w, "+ %s\n", diff.Path)
	case "deleted":
		fmt.Fprintf(w, "- %s\n", diff.Path)
	case "modified":
		fmt.Fprintf(w, "M %s\n", diff.Path)
		if showContent {
			dmp := diffmatchpatch.New()
			diffs := dmp.DiffMain(string(diff.LeftData), string(diff.RightData), false)
			fmt.Fprintln(w, dmp.DiffPrettyText(diffs))
		}
	}
}
