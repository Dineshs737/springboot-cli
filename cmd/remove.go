package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/springcli/springcli/internal/gradle"
	"github.com/springcli/springcli/internal/maven"
	"github.com/springcli/springcli/internal/project"
)

var removeCmd = &cobra.Command{
	Use:     "remove <dependency> [dependency2 ...]",
	Aliases: []string{"rm"},
	Short:   "Remove dependencies from the current Spring Boot project",
	Long: `Remove one or more dependencies from the current project's build file.
Accepts both short form (e.g., "web") and full artifact ID (e.g., "spring-boot-starter-web").`,
	Args: cobra.MinimumNArgs(1),
	RunE: runRemove,
}

var removeDir string

func init() {
	removeCmd.Flags().StringVarP(&removeDir, "dir", "d", ".", "Target project directory")
}

func runRemove(cmd *cobra.Command, args []string) error {
	detector := project.NewDetector()
	buildPath, buildType, err := detector.GetBuildFilePath(removeDir)
	if err != nil {
		return fmt.Errorf("detecting project: %w", err)
	}

	printer.Info("Detected %s project at %s", buildType, buildPath)

	for _, depID := range args {
		removed, removeErr := removeDependencyFromProject(buildPath, buildType, depID)
		if removeErr != nil {
			printer.Error("Failed to remove '%s': %v", depID, removeErr)
			continue
		}

		if !removed {
			printer.Warn("Dependency '%s' not found in project", depID)
		} else {
			printer.Success("Removed '%s'", depID)
		}
	}

	return nil
}

func removeDependencyFromProject(buildPath string, buildType project.BuildType, dep string) (bool, error) {
	switch buildType {
	case project.BuildTypeMaven:
		mp := maven.NewParser()
		return mp.RemoveDependency(buildPath, dep)
	case project.BuildTypeGradleGroovy, project.BuildTypeGradleKotlin:
		gp := gradle.NewParser()
		return gp.RemoveDependency(buildPath, dep)
	default:
		return false, fmt.Errorf("unsupported build type: %s", buildType)
	}
}
