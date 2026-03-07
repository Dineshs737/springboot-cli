package cmd

import (
	"github.com/spf13/cobra"
)

var testCmd = &cobra.Command{
	Use:   "test",
	Short: "Run the Spring Boot tests",
	Long:  `Detects whether the project uses Maven or Gradle and executes the applicable test command.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		opts := lifecycleOptions{
			action: "test",
		}
		if testClass != "" {
			opts.extraArgs = append(opts.extraArgs, "TEST_CLASS_MARKER="+testClass)
		}
		if testVerbose {
			opts.extraArgs = append(opts.extraArgs, "TEST_VERBOSE_MARKER")
		}
		return runLifecycleCommand(opts)
	},
}

var (
	testClass   string
	testVerbose bool
)

func init() {
	testCmd.Flags().StringVar(&testClass, "class", "", "Run a specific test class (e.g. MyServiceTest)")
	testCmd.Flags().BoolVar(&testVerbose, "verbose", false, "Enable verbose test output")
	rootCmd.AddCommand(testCmd)
}
