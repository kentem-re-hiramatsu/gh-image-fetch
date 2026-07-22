package main

import (
	"os"

	"github.com/kentem-re-hiramatsu/gh-image-fetch/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
