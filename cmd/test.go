package cmd

import (
	"github.com/spf13/cobra"
)

var testCmd = &cobra.Command{
	Use:   "test",
	Short: "Run the Spring Boot tests",
	Long:  `Detects whether the project uses Maven or Gradle and executes the applicable test command.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runLifecycleCommand("test")
	},
}

func init() {
	rootCmd.AddCommand(testCmd)
}
