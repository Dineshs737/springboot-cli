package cmd

import (
	"github.com/spf13/cobra"
)

var packageCmd = &cobra.Command{
	Use:   "package",
	Short: "Create executable JAR/WAR",
	Long:  `Packages the application into an executable JAR or WAR file.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		opts := lifecycleOptions{
			action: "package",
		}
		if packageSkipTests {
			opts.extraArgs = append(opts.extraArgs, "SKIP_TESTS_MARKER")
		}
		return runLifecycleCommand(opts)
	},
}

var packageSkipTests bool

func init() {
	packageCmd.Flags().BoolVar(&packageSkipTests, "skip-tests", true, "Skip running tests during package")
	rootCmd.AddCommand(packageCmd)
}
