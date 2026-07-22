package main

import (
	"os"

	"github.com/re-hiramatsu/gh-image-fetch/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
