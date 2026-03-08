package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/spf13/cobra"
)

var uninstallCmd = &cobra.Command{
	Use:   "uninstall",
	Short: "Uninstall SpringCLI",
	Long:  `Permanently removes the SpringCLI executable from your system.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		execPath, err := os.Executable()
		if err != nil {
			return fmt.Errorf("finding current executable: %w", err)
		}

		// Resolve symlinks
		execPath, err = filepath.EvalSymlinks(execPath)
		if err != nil {
			return fmt.Errorf("resolving executable path: %w", err)
		}

		printer.Info("Uninstalling SpringCLI from: %s", execPath)

		if runtime.GOOS == "windows" {
			// On Windows, an executable cannot delete itself while running.
			// Spawn a detached cmd process that waits a moment then deletes the file.
			cmdArgs := []string{"/c", "timeout", "/t", "2", ">nul", "&", "del", execPath}
			c := exec.Command("cmd.exe", cmdArgs...)
			if err := c.Start(); err != nil {
				return fmt.Errorf("scheduling uninstallation: %w", err)
			}

			printer.Success("SpringCLI has been scheduled for uninstallation.")
			printer.Dim("It will be removed automatically right after this command finishes.")
			return nil
		}

		// Non-Windows: we can just delete the file directly.
		if err := os.Remove(execPath); err != nil {
			return fmt.Errorf("removing executable: %w", err)
		}

		printer.Success("SpringCLI has been successfully uninstalled. Goodbye! 👋")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(uninstallCmd)
}
