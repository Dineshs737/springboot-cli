package main

import (
	"os"

	"github.com/springcli/springcli/cmd"
)

// Version is injected at build time via -ldflags.
var Version = "dev"

func main() {
	cmd.SetVersion(Version)
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
