package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/spf13/cobra"
)

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the Spring Boot development server",
	Long:  `Detects whether the project uses Maven or Gradle and executes the appropriate spring-boot:run or bootRun command.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runLifecycleCommand("start")
	},
}

func init() {
	rootCmd.AddCommand(startCmd)
}

func runLifecycleCommand(action string) error {
	var exePath string
	var cmdArgs []string

	// Find absolute path of the current directory to avoid "cannot run executable found relative to current directory" errors.
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("could not get current working directory: %w", err)
	}

	if _, err := os.Stat("pom.xml"); err == nil {
		// Maven project
		if os.PathSeparator == '\\' {
			exePath = filepath.Join(cwd, "mvnw.cmd")
		} else {
			exePath = filepath.Join(cwd, "mvnw")
		}

		switch action {
		case "start":
			cmdArgs = []string{"spring-boot:run"}
		case "test":
			cmdArgs = []string{"test"}
		case "build":
			cmdArgs = []string{"clean", "package"}
		}
	} else {
		_, errGradle := os.Stat("build.gradle")
		_, errKotlin := os.Stat("build.gradle.kts")
		if errGradle == nil || errKotlin == nil {
			// Gradle project
			if os.PathSeparator == '\\' {
				exePath = filepath.Join(cwd, "gradlew.bat")
			} else {
				exePath = filepath.Join(cwd, "gradlew")
			}

			switch action {
			case "start":
				cmdArgs = []string{"bootRun"}
			case "test":
				cmdArgs = []string{"test"}
			case "build":
				cmdArgs = []string{"clean", "build"}
			}
		} else {
			return fmt.Errorf("could not detect pom.xml or build.gradle in the current directory")
		}
	}

	fmt.Printf("Running: %s %v\n\n", exePath, cmdArgs)

	cmd := exec.Command(exePath, cmdArgs...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("command execution failed: %w", err)
	}

	return nil
}
