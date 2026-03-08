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

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "Browse available Spring Boot dependencies",
	Long: `List all available Spring Boot dependencies from start.spring.io.
You can search, filter by category, or view installed dependencies.`,
	RunE: runList,
}

var (
	listSearch    string
	listCategory  string
	listInstalled bool
	listDir       string
)

func init() {
	listCmd.Flags().StringVarP(&listSearch, "search", "s", "", "Search for dependencies by keyword")
	listCmd.Flags().StringVarP(&listCategory, "category", "c", "", "Filter by category name")
	listCmd.Flags().BoolVarP(&listInstalled, "installed", "i", false, "Show installed dependencies only")
	listCmd.Flags().StringVarP(&listDir, "dir", "d", ".", "Project directory (used with --installed)")
}

func runList(cmd *cobra.Command, args []string) error {
	if listInstalled {
		return listInstalledDeps()
	}

	client := initializr.NewClient()
	ctx := context.Background()

	if listSearch != "" {
		return searchDeps(ctx, client)
	}

	if listCategory != "" {
		return listByCategory(ctx, client)
	}

	return listAllDeps(ctx, client)
}

func listAllDeps(ctx context.Context, client *initializr.HTTPClient) error {
	categories, err := client.ListAllDependencies(ctx)
	if err != nil {
		return fmt.Errorf("listing dependencies: %w", err)
	}

	for _, cat := range categories {
		fmt.Println()
		printer.Header(cat.Name)
		headers := []string{"ID", "Name", "Description"}
		var rows [][]string
		for _, dep := range cat.Values {
			rows = append(rows, []string{dep.ID, dep.Name, dep.Description})
		}
		printer.Table(headers, rows)
	}

	return nil
}

func searchDeps(ctx context.Context, client *initializr.HTTPClient) error {
	results, err := client.SearchDependencies(ctx, listSearch)
	if err != nil {
		return fmt.Errorf("searching dependencies: %w", err)
	}

	if len(results) == 0 {
		printer.Warn("No dependencies found matching '%s'", listSearch)
		return nil
	}

	printer.Info("Found %d dependencies matching '%s':", len(results), listSearch)
	fmt.Println()

	headers := []string{"ID", "Name", "Description", "Category"}
	var rows [][]string
	for _, dep := range results {
		rows = append(rows, []string{dep.ID, dep.Name, dep.GroupID + ":" + dep.ArtifactID, dep.Category})
	}
	printer.Table(headers, rows)

	return nil
}

func listByCategory(ctx context.Context, client *initializr.HTTPClient) error {
	categories, err := client.ListAllDependencies(ctx)
	if err != nil {
		return fmt.Errorf("listing dependencies: %w", err)
	}

	found := false
	for _, cat := range categories {
		if strings.EqualFold(cat.Name, listCategory) {
			found = true
			fmt.Println()
			printer.Header(cat.Name)
			headers := []string{"ID", "Name", "Description"}
			var rows [][]string
			for _, dep := range cat.Values {
				rows = append(rows, []string{dep.ID, dep.Name, dep.Description})
			}
			printer.Table(headers, rows)
			break
		}
	}

	if !found {
		printer.Warn("Category '%s' not found", listCategory)
		printer.Info("Available categories:")
		for _, cat := range categories {
			printer.Dim("  - %s", cat.Name)
		}
	}

	return nil
}

func listInstalledDeps() error {
	detector := project.NewDetector()
	buildPath, buildType, err := detector.GetBuildFilePath(listDir)
	if err != nil {
		return fmt.Errorf("detecting project: %w", err)
	}

	printer.Info("Installed dependencies in %s (%s):", buildPath, buildType)
	fmt.Println()

	switch buildType {
	case project.BuildTypeMaven:
		return listMavenDeps(buildPath)
	case project.BuildTypeGradleGroovy, project.BuildTypeGradleKotlin:
		return listGradleDeps(buildPath)
	default:
		return fmt.Errorf("unsupported build type: %s", buildType)
	}
}

func listMavenDeps(path string) error {
	mp := maven.NewParser()
	deps, err := mp.ListDependencies(path)
	if err != nil {
		return fmt.Errorf("listing Maven dependencies: %w", err)
	}

	if len(deps) == 0 {
		printer.Warn("No dependencies found")
		return nil
	}

	headers := []string{"Group ID", "Artifact ID", "Version", "Scope"}
	var rows [][]string
	for _, dep := range deps {
		v := dep.Version
		if v == "" {
			v = "(managed)"
		}
		scope := dep.Scope
		if scope == "" {
			scope = "compile"
		}
		rows = append(rows, []string{dep.GroupID, dep.ArtifactID, v, scope})
	}
	printer.Table(headers, rows)

	return nil
}

func listGradleDeps(path string) error {
	gp := gradle.NewParser()
	deps, err := gp.ListDependencies(path)
	if err != nil {
		return fmt.Errorf("listing Gradle dependencies: %w", err)
	}

	if len(deps) == 0 {
		printer.Warn("No dependencies found")
		return nil
	}

	headers := []string{"Configuration", "Group", "Artifact", "Version"}
	var rows [][]string
	for _, dep := range deps {
		v := dep.Version
		if v == "" {
			v = "(managed)"
		}
		rows = append(rows, []string{dep.Configuration, dep.Group, dep.Artifact, v})
	}
	printer.Table(headers, rows)

	return nil
}
