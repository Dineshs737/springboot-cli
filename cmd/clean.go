package cmd

import (
	"github.com/spf13/cobra"
)

var cleanCmd = &cobra.Command{
	Use:   "clean",
	Short: "Clean build artifacts",
	Long:  `Runs the build tool's clean command (mvnw clean or gradlew clean) to remove generated artifacts.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		opts := lifecycleOptions{
			action: "clean",
		}
		return runLifecycleCommand(opts)
	},
}

func init() {
	rootCmd.AddCommand(cleanCmd)
}
