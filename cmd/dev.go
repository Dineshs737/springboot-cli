package cmd

import (
	"github.com/spf13/cobra"
)

var devCmd = &cobra.Command{
	Use:   "dev",
	Short: "Start the Spring Boot server with the dev profile",
	Long:  `Starts the application with the 'dev' profile activated, useful for local development and live-reloading.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		opts := lifecycleOptions{
			action: "start",
		}
		opts.envVars = append(opts.envVars, "SPRING_PROFILES_ACTIVE=dev")

		if devPort != "" {
			opts.envVars = append(opts.envVars, "SERVER_PORT="+devPort)
		}
		if devDebug {
			opts.extraArgs = append(opts.extraArgs, "--debug")
		}
		return runLifecycleCommand(opts)
	},
}

var (
	devPort  string
	devDebug bool
)

func init() {
	devCmd.Flags().StringVar(&devPort, "port", "", "Server port (e.g. 8081)")
	devCmd.Flags().BoolVar(&devDebug, "debug", false, "Enable debug mode")
	rootCmd.AddCommand(devCmd)
}
