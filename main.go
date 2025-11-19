package main

import (
	"os"

	"github.com/phlv/txtar/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
