package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/phlv/txtar/internal"
	"golang.org/x/tools/txtar"
)

var packCmd = &cobra.Command{
	Use:   "pack [DIR]",
	Short: "Pack a directory into txtar format",
	Long: `Pack a directory or Git changeset into txtar archive format.
Supports filtering, Git integration, and various output options.`,
	RunE: runPack,
}

var packOpts internal.PackOptions

func init() {
	rootCmd.AddCommand(packCmd)

	packCmd.Flags().StringVarP(&packOpts.Output, "output", "o", "-", "Output file path, '-' for stdout")
	packCmd.Flags().StringSliceVar(&packOpts.Include, "include", []string{}, "Include patterns (glob)")
	packCmd.Flags().StringSliceVar(&packOpts.Exclude, "exclude", []string{}, "Exclude patterns (glob)")
	packCmd.Flags().BoolVar(&packOpts.Git, "git", false, "Enable Git-aware mode")
	packCmd.Flags().StringVar(&packOpts.Commit, "commit", "", "Pack specific commit (requires --git)")
	packCmd.Flags().IntVar(&packOpts.Since, "since", 0, "Pack files changed in last N commits (requires --git)")
	packCmd.Flags().BoolVar(&packOpts.Staged, "staged", false, "Pack staged changes (requires --git)")
	packCmd.Flags().BoolVar(&packOpts.Worktree, "worktree", false, "Pack worktree changes (requires --git)")
	packCmd.Flags().StringVar(&packOpts.StripPrefix, "strip-prefix", "", "Strip prefix from file paths")
	packCmd.Flags().BoolVar(&packOpts.DryRun, "dry-run", false, "Show files to be packed without creating archive")
	packCmd.Flags().BoolVar(&packOpts.IgnoreBinary, "ignore-binary", false, "Skip binary files")
	packCmd.Flags().StringVar(&packOpts.TxtarIgnore, "txtarignore", ".txtarignore", "Path to txtarignore file")

	viper.BindPFlag("pack.output", packCmd.Flags().Lookup("output"))
	viper.BindPFlag("pack.exclude", packCmd.Flags().Lookup("exclude"))
	viper.BindPFlag("pack.ignore_binary", packCmd.Flags().Lookup("ignore-binary"))
}

func runPack(cmd *cobra.Command, args []string) error {
	packOpts.Dir = "."
	if len(args) > 0 {
		packOpts.Dir = args[0]
	}

	if viper.IsSet("pack.default_exclude") {
		defaultExclude := viper.GetStringSlice("pack.default_exclude")
		packOpts.Exclude = append(defaultExclude, packOpts.Exclude...)
	}

	if viper.IsSet("pack.ignore_binary") && !cmd.Flags().Changed("ignore-binary") {
		packOpts.IgnoreBinary = viper.GetBool("pack.ignore_binary")
	}

	if (packOpts.Commit != "" || packOpts.Since > 0 || packOpts.Staged || packOpts.Worktree) && !packOpts.Git {
		return fmt.Errorf("Git-specific flags require --git")
	}

	archive, files, err := internal.Pack(context.Background(), packOpts)
	if err != nil {
		return fmt.Errorf("pack failed: %w", err)
	}

	if packOpts.DryRun {
		fmt.Fprintln(os.Stderr, "Files to be packed:")
		for _, f := range files {
			fmt.Println(f)
		}
		return nil
	}

	data := txtar.Format(archive)

	if packOpts.Output == "-" {
		_, err = os.Stdout.Write(data)
	} else {
		err = os.WriteFile(packOpts.Output, data, 0644)
	}

	if err != nil {
		return fmt.Errorf("failed to write output: %w", err)
	}

	return nil
}
