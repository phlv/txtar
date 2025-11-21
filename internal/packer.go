package internal

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/bmatcuk/doublestar/v4"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"golang.org/x/tools/txtar"
)

type PackOptions struct {
	Dir           string
	Output        string
	Include       []string
	Exclude       []string
	Git           bool
	Commit        string
	Since         int
	Staged        bool
	Worktree      bool
	StripPrefix   string
	DryRun        bool
	IgnoreBinary  bool
	TxtarIgnore   string
}

type Filter struct {
	include       []string
	exclude       []string
	gitignore     []string
	txtarignore   []string
	ignoreBinary  bool
}

func NewFilter(opts PackOptions) (*Filter, error) {
	f := &Filter{
		include:      opts.Include,
		exclude:      opts.Exclude,
		ignoreBinary: opts.IgnoreBinary,
	}

	if opts.Git {
		gitignore, err := readGitignore(opts.Dir)
		if err == nil {
			f.gitignore = gitignore
		}
	}

	if opts.TxtarIgnore != "" {
		txtarignore, err := readIgnoreFile(filepath.Join(opts.Dir, opts.TxtarIgnore))
		if err == nil {
			f.txtarignore = txtarignore
		}
	}

	return f, nil
}

func (f *Filter) ShouldInclude(path string) bool {
	if len(f.include) > 0 {
		matched := false
		for _, pattern := range f.include {
			if m, _ := doublestar.Match(pattern, path); m {
				matched = true
				break
			}
		}
		if !matched {
			return false
		}
	}

	for _, pattern := range f.exclude {
		if m, _ := doublestar.Match(pattern, path); m {
			return false
		}
	}

	for _, pattern := range f.gitignore {
		if pattern == "" || strings.HasPrefix(pattern, "#") {
			continue
		}
		if m, _ := doublestar.Match(pattern, path); m {
			return false
		}
	}

	for _, pattern := range f.txtarignore {
		if pattern == "" || strings.HasPrefix(pattern, "#") {
			continue
		}
		if m, _ := doublestar.Match(pattern, path); m {
			return false
		}
	}

	return true
}

func readGitignore(dir string) ([]string, error) {
	return readIgnoreFile(filepath.Join(dir, ".gitignore"))
}

func readIgnoreFile(path string) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var patterns []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			patterns = append(patterns, line)
		}
	}
	return patterns, scanner.Err()
}

func isBinaryFile(path string) (bool, error) {
	f, err := os.Open(path)
	if err != nil {
		return false, err
	}
	defer f.Close()

	buf := make([]byte, 1024)
	n, err := f.Read(buf)
	if err != nil && err != io.EOF {
		return false, err
	}

	return bytes.Contains(buf[:n], []byte{0}), nil
}

func Pack(ctx context.Context, opts PackOptions) (*txtar.Archive, []string, error) {
	filter, err := NewFilter(opts)
	if err != nil {
		return nil, nil, err
	}

	var files []string
	var fileContents map[string][]byte

	if opts.Git {
		files, fileContents, err = packGit(opts, filter)
	} else {
		files, fileContents, err = packDir(opts, filter)
	}

	if err != nil {
		return nil, nil, err
	}

	if opts.DryRun {
		return nil, files, nil
	}

	archive := &txtar.Archive{}
	for _, file := range files {
		relativePath := file
		if opts.StripPrefix != "" {
			relativePath = strings.TrimPrefix(file, opts.StripPrefix)
			relativePath = strings.TrimPrefix(relativePath, "/")
		}

		relativePath = filepath.ToSlash(relativePath)

		content := fileContents[file]
		archive.Files = append(archive.Files, txtar.File{
			Name: relativePath,
			Data: content,
		})
	}

	return archive, files, nil
}

func packDir(opts PackOptions, filter *Filter) ([]string, map[string][]byte, error) {
	dir := opts.Dir
	if dir == "" {
		dir = "."
	}

	var files []string
	fileContents := make(map[string][]byte)

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

		if !filter.ShouldInclude(relPath) {
			return nil
		}

		if filter.ignoreBinary {
			isBinary, err := isBinaryFile(path)
			if err != nil {
				return err
			}
			if isBinary {
				return nil
			}
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		files = append(files, relPath)
		fileContents[relPath] = content

		return nil
	})

	return files, fileContents, err
}

func packGit(opts PackOptions, filter *Filter) ([]string, map[string][]byte, error) {
	repo, err := git.PlainOpen(opts.Dir)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to open git repository: %w", err)
	}

	if opts.Commit != "" {
		return packCommit(repo, opts.Commit, filter)
	}

	if opts.Since > 0 {
		return packSince(repo, opts.Since, filter)
	}

	if opts.Staged {
		return packStaged(repo, opts.Dir, filter)
	}

	if opts.Worktree {
		return packWorktree(repo, opts.Dir, filter)
	}

	head, err := repo.Head()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get HEAD: %w", err)
	}

	return packCommit(repo, head.Hash().String(), filter)
}

func packCommit(repo *git.Repository, commitHash string, filter *Filter) ([]string, map[string][]byte, error) {
	hash := plumbing.NewHash(commitHash)
	commit, err := repo.CommitObject(hash)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get commit: %w", err)
	}

	tree, err := commit.Tree()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get tree: %w", err)
	}

	var files []string
	fileContents := make(map[string][]byte)

	err = tree.Files().ForEach(func(f *object.File) error {
		if !filter.ShouldInclude(f.Name) {
			return nil
		}

		content, err := f.Contents()
		if err != nil {
			return err
		}

		if filter.ignoreBinary && bytes.Contains([]byte(content)[:min(1024, len(content))], []byte{0}) {
			return nil
		}

		files = append(files, f.Name)
		fileContents[f.Name] = []byte(content)
		return nil
	})

	return files, fileContents, err
}

func packSince(repo *git.Repository, n int, filter *Filter) ([]string, map[string][]byte, error) {
	head, err := repo.Head()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get HEAD: %w", err)
	}

	commits, err := repo.Log(&git.LogOptions{
		From: head.Hash(),
	})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get log: %w", err)
	}

	var commitList []*object.Commit
	err = commits.ForEach(func(c *object.Commit) error {
		if len(commitList) < n+1 {
			commitList = append(commitList, c)
		}
		return nil
	})
	if err != nil {
		return nil, nil, err
	}

	if len(commitList) < n+1 {
		return nil, nil, fmt.Errorf("not enough commits in history")
	}

	changedFiles := make(map[string]bool)
	for i := 0; i < n; i++ {
		var parent *object.Commit
		if i+1 < len(commitList) {
			parent = commitList[i+1]
		}

		if parent != nil {
			patch, err := parent.Patch(commitList[i])
			if err != nil {
				return nil, nil, err
			}

			for _, filePatch := range patch.FilePatches() {
				from, to := filePatch.Files()
				if to != nil {
					changedFiles[to.Path()] = true
				} else if from != nil {
					changedFiles[from.Path()] = false
				}
			}
		}
	}

	headCommit := commitList[0]
	tree, err := headCommit.Tree()
	if err != nil {
		return nil, nil, err
	}

	var files []string
	fileContents := make(map[string][]byte)

	for path, exists := range changedFiles {
		if !exists {
			continue
		}

		if !filter.ShouldInclude(path) {
			continue
		}

		file, err := tree.File(path)
		if err != nil {
			continue
		}

		content, err := file.Contents()
		if err != nil {
			continue
		}

		if filter.ignoreBinary && bytes.Contains([]byte(content)[:min(1024, len(content))], []byte{0}) {
			continue
		}

		files = append(files, path)
		fileContents[path] = []byte(content)
	}

	return files, fileContents, nil
}

func packStaged(repo *git.Repository, dir string, filter *Filter) ([]string, map[string][]byte, error) {
	w, err := repo.Worktree()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get worktree: %w", err)
	}

	status, err := w.Status()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get status: %w", err)
	}

	var files []string
	fileContents := make(map[string][]byte)

	for path, fileStatus := range status {
		if fileStatus.Staging == git.Unmodified {
			continue
		}

		if !filter.ShouldInclude(path) {
			continue
		}

		fullPath := filepath.Join(dir, path)
		content, err := os.ReadFile(fullPath)
		if err != nil {
			continue
		}

		if filter.ignoreBinary && bytes.Contains(content[:min(1024, len(content))], []byte{0}) {
			continue
		}

		files = append(files, path)
		fileContents[path] = content
	}

	return files, fileContents, nil
}

func packWorktree(repo *git.Repository, dir string, filter *Filter) ([]string, map[string][]byte, error) {
	w, err := repo.Worktree()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get worktree: %w", err)
	}

	status, err := w.Status()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get status: %w", err)
	}

	var files []string
	fileContents := make(map[string][]byte)

	for path, fileStatus := range status {
		if fileStatus.Worktree == git.Unmodified {
			continue
		}

		if !filter.ShouldInclude(path) {
			continue
		}

		fullPath := filepath.Join(dir, path)
		content, err := os.ReadFile(fullPath)
		if err != nil {
			continue
		}

		if filter.ignoreBinary && bytes.Contains(content[:min(1024, len(content))], []byte{0}) {
			continue
		}

		files = append(files, path)
		fileContents[path] = content
	}

	return files, fileContents, nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
