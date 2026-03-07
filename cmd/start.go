package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the Spring Boot development server",
	Long:  `Detects whether the project uses Maven or Gradle and executes the appropriate spring-boot:run or bootRun command.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		opts := lifecycleOptions{
			action: "start",
		}
		if startProfile != "" {
			opts.envVars = append(opts.envVars, "SPRING_PROFILES_ACTIVE="+startProfile)
		}
		if startPort != "" {
			opts.envVars = append(opts.envVars, "SERVER_PORT="+startPort)
		}
		if startDebug {
			opts.extraArgs = append(opts.extraArgs, "--debug")
		}
		return runLifecycleCommand(opts)
	},
}

var (
	startProfile string
	startPort    string
	startDebug   bool
)

func init() {
	startCmd.Flags().StringVar(&startProfile, "profile", "", "Spring active profile (e.g. dev)")
	startCmd.Flags().StringVar(&startPort, "port", "", "Server port (e.g. 8081)")
	startCmd.Flags().BoolVar(&startDebug, "debug", false, "Enable debug mode")
	rootCmd.AddCommand(startCmd)
}

type lifecycleOptions struct {
	action    string
	extraArgs []string
	envVars   []string
}

func runLifecycleCommand(opts lifecycleOptions) error {
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

		switch opts.action {
		case "start":
			cmdArgs = []string{"spring-boot:run"}
		case "test":
			cmdArgs = []string{"test"}
		case "build":
			cmdArgs = []string{"clean", "package"}
		case "clean":
			cmdArgs = []string{"clean"}
		case "deps":
			cmdArgs = []string{"dependency:tree"}
		case "package":
			cmdArgs = []string{"package"}
		}

		// Process extraArgs markers for Maven
		var finalExtraArgs []string
		for _, arg := range opts.extraArgs {
			if arg == "SKIP_TESTS_MARKER" {
				finalExtraArgs = append(finalExtraArgs, "-DskipTests")
			} else if strings.HasPrefix(arg, "TEST_CLASS_MARKER=") {
				className := strings.TrimPrefix(arg, "TEST_CLASS_MARKER=")
				finalExtraArgs = append(finalExtraArgs, "-Dtest="+className)
			} else if arg == "TEST_VERBOSE_MARKER" {
				finalExtraArgs = append(finalExtraArgs, "-X")
			} else {
				finalExtraArgs = append(finalExtraArgs, arg)
			}
		}
		opts.extraArgs = finalExtraArgs
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

			switch opts.action {
			case "start":
				cmdArgs = []string{"bootRun"}
			case "test":
				cmdArgs = []string{"test"}
			case "build":
				cmdArgs = []string{"clean", "build"}
			case "clean":
				cmdArgs = []string{"clean"}
			case "deps":
				cmdArgs = []string{"dependencies"}
			case "package":
				cmdArgs = []string{"bootJar"}
			}

			// Process extraArgs markers for Gradle
			var finalExtraArgs []string
			for _, arg := range opts.extraArgs {
				if arg == "SKIP_TESTS_MARKER" {
					finalExtraArgs = append(finalExtraArgs, "-x", "test")
				} else if strings.HasPrefix(arg, "TEST_CLASS_MARKER=") {
					className := strings.TrimPrefix(arg, "TEST_CLASS_MARKER=")
					finalExtraArgs = append(finalExtraArgs, "--tests", className)
				} else if arg == "TEST_VERBOSE_MARKER" {
					finalExtraArgs = append(finalExtraArgs, "--info")
				} else {
					finalExtraArgs = append(finalExtraArgs, arg)
				}
			}
			opts.extraArgs = finalExtraArgs
		} else {
			return fmt.Errorf("could not detect pom.xml or build.gradle in the current directory")
		}
	}

	if len(opts.extraArgs) > 0 {
		cmdArgs = append(cmdArgs, opts.extraArgs...)
	}

	fmt.Printf("Running: %s %v\n\n", exePath, cmdArgs)

	cmd := exec.Command(exePath, cmdArgs...)

	// Ensure we pass through current environment variables, plus any extras
	env := os.Environ()
	if len(opts.envVars) > 0 {
		env = append(env, opts.envVars...)
	}
	cmd.Env = env

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("command execution failed: %w", err)
	}

	return nil
}
