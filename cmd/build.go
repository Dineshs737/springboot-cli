package cmd

import (
	"github.com/spf13/cobra"
)

var buildCmd = &cobra.Command{
	Use:   "build",
	Short: "Build the Spring Boot production artifacts",
	Long:  `Detects whether the project uses Maven or Gradle and executes clean and package/build steps.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		opts := lifecycleOptions{
			action: "build",
		}
		if buildSkipTests {
			// handled in runLifecycleCommand logic
			opts.extraArgs = append(opts.extraArgs, "SKIP_TESTS_MARKER")
		}
		return runLifecycleCommand(opts)
	},
}

var buildSkipTests bool

func init() {
	buildCmd.Flags().BoolVar(&buildSkipTests, "skip-tests", false, "Skip running tests during build")
	rootCmd.AddCommand(buildCmd)
}
