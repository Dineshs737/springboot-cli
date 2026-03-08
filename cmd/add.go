package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/springcli/springcli/internal/gradle"
	"github.com/springcli/springcli/internal/initializr"
	"github.com/springcli/springcli/internal/maven"
	"github.com/springcli/springcli/internal/project"
)

var addCmd = &cobra.Command{
	Use:   "add <dependency> [dependency2 ...]",
	Short: "Add dependencies to the current Spring Boot project",
	Long: `Add one or more Spring Boot dependencies to the current project.
Dependencies are resolved from the Spring Initializr metadata and added
to the appropriate build file (pom.xml or build.gradle).`,
	Args: cobra.MinimumNArgs(1),
	RunE: runAdd,
}

var addDir string

func init() {
	addCmd.Flags().StringVarP(&addDir, "dir", "d", ".", "Target project directory")
}

func runAdd(cmd *cobra.Command, args []string) error {
	detector := project.NewDetector()
	buildPath, buildType, err := detector.GetBuildFilePath(addDir)
	if err != nil {
		return fmt.Errorf("detecting project: %w", err)
	}

	printer.Info("Detected %s project at %s", buildType, buildPath)

	// Fetch metadata once for all deps.
	client := initializr.NewClient()
	ctx := context.Background()

	for _, depID := range args {
		// Parse optional @version suffix
		var explicitVersion string
		rawDepID := depID
		if parts := strings.SplitN(depID, "@", 2); len(parts) == 2 {
			rawDepID = parts[0]
			explicitVersion = parts[1]
		}

		resolved, err := client.ResolveDependency(ctx, rawDepID)
		if err != nil {
			// Try fuzzy search for suggestions.
			suggestions, searchErr := client.SearchDependencies(ctx, rawDepID)
			if searchErr == nil && len(suggestions) > 0 {
				printer.Error("Dependency '%s' not found. Did you mean:", rawDepID)
				for _, s := range suggestions {
					printer.Dim("  - %s (%s)", s.ID, s.Name)
				}
			} else {
				printer.Error("Dependency '%s' not found", rawDepID)
			}
			continue
		}

		if explicitVersion != "" {
			resolved.Version = explicitVersion
			// If an explicit version is specified, it's not managed exclusively by the Spring Boot bom
			resolved.Starter = false
		}

		added, addErr := addDependencyToProject(buildPath, buildType, resolved)
		if addErr != nil {
			printer.Error("Failed to add '%s': %v", depID, addErr)
			continue
		}

		if !added {
			printer.Warn("Dependency '%s' (%s) is already present", depID, resolved.ArtifactID)
		} else {
			printer.Success("Added %s (%s:%s)", resolved.Name, resolved.GroupID, resolved.ArtifactID)
		}
	}

	return nil
}

func addDependencyToProject(buildPath string, buildType project.BuildType, dep *initializr.ResolvedDependency) (bool, error) {
	// For Spring Boot managed deps (starters), don't add version.
	version := dep.Version
	if dep.Starter {
		version = ""
	}

	switch buildType {
	case project.BuildTypeMaven:
		mp := maven.NewParser()
		return mp.AddDependency(buildPath, maven.MavenDependency{
			GroupID:    dep.GroupID,
			ArtifactID: dep.ArtifactID,
			Version:    version,
			Scope:      dep.Scope,
		})
	case project.BuildTypeGradleGroovy, project.BuildTypeGradleKotlin:
		gp := gradle.NewParser()
		return gp.AddDependency(buildPath, dep.GroupID, dep.ArtifactID, version, dep.Scope)
	default:
		return false, fmt.Errorf("unsupported build type: %s", buildType)
	}
}
