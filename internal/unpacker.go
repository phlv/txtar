package internal

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"golang.org/x/tools/txtar"
)

type UnpackOptions struct {
	Archive      string
	Dir          string
	Backup       bool
	DryRun       bool
	NoOverwrite  bool
}

func Unpack(archive *txtar.Archive, opts UnpackOptions) error {
	if opts.Backup && opts.NoOverwrite {
		return fmt.Errorf("--backup and --no-overwrite are mutually exclusive")
	}

	if opts.Dir == "" {
		opts.Dir = "."
	}

	for _, file := range archive.Files {
		if err := validatePath(file.Name); err != nil {
			return fmt.Errorf("invalid path %q: %w", file.Name, err)
		}

		targetPath := filepath.Join(opts.Dir, file.Name)

		if opts.DryRun {
			fmt.Printf("Would write: %s\n", targetPath)
			continue
		}

		if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
			return fmt.Errorf("failed to create directory for %q: %w", targetPath, err)
		}

		if _, err := os.Stat(targetPath); err == nil {
			if opts.NoOverwrite {
				return fmt.Errorf("file exists: %s (use --backup to backup or remove --no-overwrite)", targetPath)
			}

			if opts.Backup {
				backupPath := targetPath + ".bak"
				if _, err := os.Stat(backupPath); err == nil {
					timestamp := time.Now().Format("20060102T150405")
					backupPath = fmt.Sprintf("%s.bak.%s", targetPath, timestamp)
				}

				if err := os.Rename(targetPath, backupPath); err != nil {
					return fmt.Errorf("failed to backup %q: %w", targetPath, err)
				}
				fmt.Fprintf(os.Stderr, "Backed up: %s -> %s\n", targetPath, backupPath)
			}
		}

		if err := os.WriteFile(targetPath, file.Data, 0644); err != nil {
			return fmt.Errorf("failed to write %q: %w", targetPath, err)
		}
	}

	return nil
}

func validatePath(path string) error {
	if filepath.IsAbs(path) {
		return fmt.Errorf("absolute paths not allowed")
	}

	cleaned := filepath.Clean(path)
	if strings.HasPrefix(cleaned, "..") || strings.Contains(cleaned, "/../") {
		return fmt.Errorf("path traversal detected")
	}

	return nil
}
