package cmd

import (
	"github.com/spf13/cobra"
)

var buildCmd = &cobra.Command{
	Use:   "build",
	Short: "Build the Spring Boot production artifacts",
	Long:  `Detects whether the project uses Maven or Gradle and executes clean and package/build steps.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runLifecycleCommand("build")
	},
}

func init() {
	rootCmd.AddCommand(buildCmd)
}
