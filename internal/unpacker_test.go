package internal

import (
	"os"
	"path/filepath"
	"testing"

	"golang.org/x/tools/txtar"
)

func TestUnpack(t *testing.T) {
	tmpDir := t.TempDir()

	archive := &txtar.Archive{
		Files: []txtar.File{
			{Name: "test.txt", Data: []byte("hello world")},
			{Name: "subdir/nested.txt", Data: []byte("nested content")},
		},
	}

	opts := UnpackOptions{
		Dir: tmpDir,
	}

	err := Unpack(archive, opts)
	if err != nil {
		t.Fatalf("Unpack failed: %v", err)
	}

	content, err := os.ReadFile(filepath.Join(tmpDir, "test.txt"))
	if err != nil {
		t.Fatalf("Failed to read unpacked file: %v", err)
	}

	if string(content) != "hello world" {
		t.Errorf("Expected 'hello world', got %q", string(content))
	}

	nestedContent, err := os.ReadFile(filepath.Join(tmpDir, "subdir/nested.txt"))
	if err != nil {
		t.Fatalf("Failed to read nested file: %v", err)
	}

	if string(nestedContent) != "nested content" {
		t.Errorf("Expected 'nested content', got %q", string(nestedContent))
	}
}

func TestUnpackPathTraversal(t *testing.T) {
	tmpDir := t.TempDir()

	archive := &txtar.Archive{
		Files: []txtar.File{
			{Name: "../evil.txt", Data: []byte("malicious")},
		},
	}

	opts := UnpackOptions{
		Dir: tmpDir,
	}

	err := Unpack(archive, opts)
	if err == nil {
		t.Fatal("Expected error for path traversal, got nil")
	}
}

func TestUnpackBackup(t *testing.T) {
	tmpDir := t.TempDir()

	existingFile := filepath.Join(tmpDir, "test.txt")
	os.WriteFile(existingFile, []byte("original"), 0644)

	archive := &txtar.Archive{
		Files: []txtar.File{
			{Name: "test.txt", Data: []byte("new content")},
		},
	}

	opts := UnpackOptions{
		Dir:    tmpDir,
		Backup: true,
	}

	err := Unpack(archive, opts)
	if err != nil {
		t.Fatalf("Unpack failed: %v", err)
	}

	backupFile := filepath.Join(tmpDir, "test.txt.bak")
	backupContent, err := os.ReadFile(backupFile)
	if err != nil {
		t.Fatalf("Backup file not created: %v", err)
	}

	if string(backupContent) != "original" {
		t.Errorf("Expected backup to contain 'original', got %q", string(backupContent))
	}

	newContent, err := os.ReadFile(existingFile)
	if err != nil {
		t.Fatalf("Failed to read new file: %v", err)
	}

	if string(newContent) != "new content" {
		t.Errorf("Expected 'new content', got %q", string(newContent))
	}
}
