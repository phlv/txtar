package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/phlv/txtar/internal"
)

var diffCmd = &cobra.Command{
	Use:   "diff [LEFT] [RIGHT]",
	Short: "Compare two txtar archives or a directory and archive",
	Long: `Compare two txtar archives or compare a directory with an archive.
Shows added, deleted, and modified files.`,
	Args: cobra.ExactArgs(2),
	RunE: runDiff,
}

var (
	diffDir     bool
	diffContent bool
)

func init() {
	rootCmd.AddCommand(diffCmd)

	diffCmd.Flags().BoolVar(&diffDir, "dir", false, "Treat first argument as directory")
	diffCmd.Flags().BoolVarP(&diffContent, "content", "c", false, "Show content differences")
}

func runDiff(cmd *cobra.Command, args []string) error {
	opts := internal.DiffOptions{
		Left:  args[0],
		Right: args[1],
		IsDir: diffDir,
	}

	diffs, err := internal.Diff(opts)
	if err != nil {
		return fmt.Errorf("diff failed: %w", err)
	}

	if len(diffs) == 0 {
		fmt.Println("No differences found")
		return nil
	}

	for _, diff := range diffs {
		internal.PrintDiff(os.Stdout, diff, diffContent)
	}

	return nil
}
