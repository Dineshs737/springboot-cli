package cmd

import (
	"github.com/spf13/cobra"
)

var depsCmd = &cobra.Command{
	Use:   "deps",
	Short: "Show project dependency tree",
	Long:  `Displays the dependency tree of the project using the underlying build tool (mvnw dependency:tree or gradlew dependencies).`,
	RunE: func(cmd *cobra.Command, args []string) error {
		opts := lifecycleOptions{
			action: "deps",
		}
		return runLifecycleCommand(opts)
	},
}

func init() {
	rootCmd.AddCommand(depsCmd)
}
