package cmd

import (
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/phlv/txtar/internal"
	"golang.org/x/tools/txtar"
)

var unpackCmd = &cobra.Command{
	Use:   "unpack [ARCHIVE]",
	Short: "Unpack a txtar archive to filesystem",
	Long: `Unpack a txtar archive to the filesystem.
Reads from stdin if ARCHIVE is '-' or not specified.`,
	RunE: runUnpack,
}

var unpackOpts internal.UnpackOptions

func init() {
	rootCmd.AddCommand(unpackCmd)

	unpackCmd.Flags().StringVarP(&unpackOpts.Dir, "dir", "C", ".", "Output directory")
	unpackCmd.Flags().BoolVar(&unpackOpts.Backup, "backup", false, "Backup existing files before overwriting")
	unpackCmd.Flags().BoolVar(&unpackOpts.DryRun, "dry-run", false, "Show operations without writing files")
	unpackCmd.Flags().BoolVar(&unpackOpts.NoOverwrite, "no-overwrite", false, "Fail if files exist (mutually exclusive with --backup)")

	viper.BindPFlag("unpack.backup", unpackCmd.Flags().Lookup("backup"))
	viper.BindPFlag("unpack.dir", unpackCmd.Flags().Lookup("dir"))
}

func runUnpack(cmd *cobra.Command, args []string) error {
	var data []byte
	var err error

	archivePath := "-"
	if len(args) > 0 {
		archivePath = args[0]
	}
	unpackOpts.Archive = archivePath

	if viper.IsSet("unpack.backup") && !cmd.Flags().Changed("backup") {
		unpackOpts.Backup = viper.GetBool("unpack.backup")
	}

	if viper.IsSet("unpack.dir") && !cmd.Flags().Changed("dir") {
		unpackOpts.Dir = viper.GetString("unpack.dir")
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

	if err := internal.Unpack(archive, unpackOpts); err != nil {
		return fmt.Errorf("unpack failed: %w", err)
	}

	if !unpackOpts.DryRun {
		fmt.Fprintf(os.Stderr, "Successfully unpacked %d files to %s\n", len(archive.Files), unpackOpts.Dir)
	}

	return nil
}
