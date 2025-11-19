package internal

import (
	"context"
	"os"
	"path/filepath"
	"testing"
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
