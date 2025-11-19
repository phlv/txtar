package cmd

import (
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"
	"golang.org/x/tools/txtar"
)

var listCmd = &cobra.Command{
	Use:   "list [ARCHIVE]",
	Short: "List files in a txtar archive",
	Long: `List all files contained in a txtar archive.
Reads from stdin if ARCHIVE is '-' or not specified.`,
	RunE: runList,
}

func init() {
	rootCmd.AddCommand(listCmd)
}

func runList(cmd *cobra.Command, args []string) error {
	var data []byte
	var err error

	archivePath := "-"
	if len(args) > 0 {
		archivePath = args[0]
	}

	if archivePath == "-" {
		data, err = io.ReadAll(os.Stdin)
	} else {
		data, err = os.ReadFile(archivePath)
	}

	if err != nil {
		return fmt.Errorf("failed to read archive: %w", err)
	}

	archive := txtar.Parse(data)

	for _, file := range archive.Files {
		fmt.Println(file.Name)
	}

	return nil
}
