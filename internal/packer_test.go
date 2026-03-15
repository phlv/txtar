package internal

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
)

func TestPack(t *testing.T) {
	tmpDir := t.TempDir()

	os.WriteFile(filepath.Join(tmpDir, "test.txt"), []byte("hello world"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "test.log"), []byte("log content"), 0644)

	opts := PackOptions{
		Dir:     tmpDir,
		Exclude: []string{"*.log"},
	}

	archive, files, err := Pack(context.Background(), opts)
	if err != nil {
		t.Fatalf("Pack failed: %v", err)
	}

	if len(files) != 1 {
		t.Errorf("Expected 1 file, got %d", len(files))
	}

	if archive.Files[0].Name != "test.txt" {
		t.Errorf("Expected test.txt, got %s", archive.Files[0].Name)
	}
}

func TestPackDryRun(t *testing.T) {
	tmpDir := t.TempDir()

	os.WriteFile(filepath.Join(tmpDir, "test.txt"), []byte("hello"), 0644)

	opts := PackOptions{
		Dir:    tmpDir,
		DryRun: true,
	}

	archive, files, err := Pack(context.Background(), opts)
	if err != nil {
		t.Fatalf("Pack failed: %v", err)
	}

	if archive != nil {
		t.Error("Expected nil archive in dry-run mode")
	}

	if len(files) != 1 {
		t.Errorf("Expected 1 file, got %d", len(files))
	}
}

func TestPackGitDiffIncludesStagedUnstagedAndUntracked(t *testing.T) {
	tmpDir := t.TempDir()

	repo, err := git.PlainInit(tmpDir, false)
	if err != nil {
		t.Fatalf("PlainInit failed: %v", err)
	}

	stagedPath := filepath.Join(tmpDir, "staged.txt")
	unstagedPath := filepath.Join(tmpDir, "unstaged.txt")

	if err := os.WriteFile(stagedPath, []byte("v1"), 0644); err != nil {
		t.Fatalf("WriteFile staged initial failed: %v", err)
	}
	if err := os.WriteFile(unstagedPath, []byte("v1"), 0644); err != nil {
		t.Fatalf("WriteFile unstaged initial failed: %v", err)
	}

	w, err := repo.Worktree()
	if err != nil {
		t.Fatalf("Worktree failed: %v", err)
	}

	if _, err := w.Add("staged.txt"); err != nil {
		t.Fatalf("Add staged initial failed: %v", err)
	}
	if _, err := w.Add("unstaged.txt"); err != nil {
		t.Fatalf("Add unstaged initial failed: %v", err)
	}

	if _, err := w.Commit("initial commit", &git.CommitOptions{
		Author: &object.Signature{
			Name:  "test",
			Email: "test@example.com",
			When:  time.Now(),
		},
	}); err != nil {
		t.Fatalf("Commit failed: %v", err)
	}

	if err := os.WriteFile(stagedPath, []byte("staged change"), 0644); err != nil {
		t.Fatalf("WriteFile staged updated failed: %v", err)
	}
	if _, err := w.Add("staged.txt"); err != nil {
		t.Fatalf("Add staged updated failed: %v", err)
	}

	if err := os.WriteFile(unstagedPath, []byte("unstaged change"), 0644); err != nil {
		t.Fatalf("WriteFile unstaged updated failed: %v", err)
	}

	if err := os.WriteFile(filepath.Join(tmpDir, "untracked.txt"), []byte("untracked"), 0644); err != nil {
		t.Fatalf("WriteFile untracked failed: %v", err)
	}

	opts := PackOptions{
		Dir:  tmpDir,
		Git:  true,
		Diff: true,
	}

	archive, files, err := Pack(context.Background(), opts)
	if err != nil {
		t.Fatalf("Pack failed: %v", err)
	}

	if len(files) != 3 {
		t.Fatalf("Expected 3 files, got %d: %v", len(files), files)
	}

	got := make(map[string]string)
	for _, f := range archive.Files {
		got[f.Name] = string(f.Data)
	}

	if got["staged.txt"] != "staged change" {
		t.Fatalf("Expected staged.txt to contain staged change, got %q", got["staged.txt"])
	}
	if got["unstaged.txt"] != "unstaged change" {
		t.Fatalf("Expected unstaged.txt to contain unstaged change, got %q", got["unstaged.txt"])
	}
	if got["untracked.txt"] != "untracked" {
		t.Fatalf("Expected untracked.txt to contain untracked, got %q", got["untracked.txt"])
	}
}
