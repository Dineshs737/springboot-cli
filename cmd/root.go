package cmd

import (
	"github.com/spf13/cobra"
	"github.com/springcli/springcli/internal/ui"
)

var (
	version = "dev"
	printer = ui.NewPrinter()
	banner  = ui.NewBannerPrinter()
)

// SetVersion sets the version string (injected from main).
func SetVersion(v string) {
	version = v
}

var rootCmd = &cobra.Command{
	Use:   "springcli",
	Short: "Spring Boot CLI — Create and manage Spring Boot projects from the terminal",
	Long: `SpringCLI is a production-grade CLI tool that lets developers
create and manage Spring Boot projects from the terminal,
like create-next-app but for Spring Boot.`,
	Run: func(cmd *cobra.Command, args []string) {
		banner.Print(version)
		_ = cmd.Help()
	},
}

func init() {
	rootCmd.AddCommand(newCmd)
	rootCmd.AddCommand(addCmd)
	rootCmd.AddCommand(removeCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(securityCmd)
}

// Execute runs the root command.
func Execute() error {
	return rootCmd.Execute()
}
