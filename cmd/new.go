package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/springcli/springcli/internal/initializr"
	"github.com/springcli/springcli/internal/project"
	"github.com/springcli/springcli/internal/wizard"
)

var newCmd = &cobra.Command{
	Use:   "new [project-name]",
	Short: "Create a new Spring Boot project with an interactive wizard",
	Long: `Scaffold a new Spring Boot project using a beautiful step-by-step wizard.
All options can also be provided via flags for non-interactive usage (CI/CD).`,
	Args: cobra.MaximumNArgs(1),
	RunE: runNew,
}

var (
	newGroup         string
	newArtifact      string
	newDescription   string
	newLanguage      string
	newType          string
	newBoot          string
	newJava          string
	newPackaging     string
	newDeps          string
	newGit           bool
	newDocker        bool
	newEnv           bool
	newNoInteractive bool
)

func init() {
	newCmd.Flags().StringVarP(&newGroup, "group", "g", "com.example", "Maven group ID")
	newCmd.Flags().StringVarP(&newArtifact, "artifact", "a", "", "Maven artifact ID (defaults to project name)")
	newCmd.Flags().StringVarP(&newDescription, "description", "d", "", "Project description")
	newCmd.Flags().StringVarP(&newLanguage, "language", "l", "java", "Language: java, kotlin, groovy")
	newCmd.Flags().StringVarP(&newType, "type", "t", "maven", "Build type: maven, gradle, gradle-kotlin")
	newCmd.Flags().StringVarP(&newBoot, "boot", "b", "", "Spring Boot version")
	newCmd.Flags().StringVarP(&newJava, "java", "j", "21", "Java version: 17, 21")
	newCmd.Flags().StringVar(&newPackaging, "packaging", "jar", "Packaging: jar, war")
	newCmd.Flags().StringVar(&newDeps, "deps", "", "Comma-separated dependency IDs (e.g., web,jpa,security)")
	newCmd.Flags().BoolVar(&newGit, "git", true, "Initialize git repository")
	newCmd.Flags().BoolVar(&newDocker, "docker", false, "Generate Docker files")
	newCmd.Flags().BoolVar(&newEnv, "env", false, "Generate .env.example file")
	newCmd.Flags().BoolVar(&newNoInteractive, "no-interactive", false, "Skip interactive wizard, use flags only")
}

func runNew(cmd *cobra.Command, args []string) error {
	projectName := ""
	if len(args) > 0 {
		projectName = args[0]
	}

	client := initializr.NewClient()
	var result *wizard.WizardResult

	// Decide interactive vs non-interactive.
	if newNoInteractive || !wizard.IsInteractive() {
		// Non-interactive: use flags + defaults.
		if projectName == "" {
			return fmt.Errorf("project name is required in non-interactive mode")
		}

		var deps []string
		if newDeps != "" {
			deps = strings.Split(newDeps, ",")
			for i := range deps {
				deps[i] = strings.TrimSpace(deps[i])
			}
		}

		result = wizard.FromNonInteractive(wizard.NonInteractiveConfig{
			ProjectName:    projectName,
			GroupID:        newGroup,
			ArtifactID:     newArtifact,
			Description:    newDescription,
			Language:       newLanguage,
			BuildType:      newType,
			BootVersion:    newBoot,
			JavaVersion:    newJava,
			Packaging:      newPackaging,
			Dependencies:   deps,
			InitGit:        newGit,
			GenerateEnv:    newEnv,
			GenerateDocker: newDocker,
		})
	} else {
		// Interactive wizard.
		banner.Print(version)

		var err error
		result, err = wizard.RunWizard(projectName, client)
		if err != nil {
			return fmt.Errorf("wizard: %w", err)
		}
	}

	// Check if target directory already exists.
	targetDir := filepath.Join(".", result.ProjectName)
	if _, err := os.Stat(targetDir); err == nil {
		return fmt.Errorf("directory '%s' already exists", result.ProjectName)
	}

	// Print the creation banner.
	fmt.Println()
	printer.Dim("  ─────────────────────────────────────────────")
	printer.Info("  Creating your Spring Boot project...")
	printer.Dim("  ─────────────────────────────────────────────")
	fmt.Println()

	// Build package name.
	packageName := result.PackageName
	if packageName == "" {
		packageName = wizard.DerivePackageName(result.GroupID, result.ArtifactID)
	}

	// Step 1: Download project.
	ctx := context.Background()
	s := printer.StartSpinner("Downloading project from start.spring.io...")
	config := initializr.ProjectDownloadConfig{
		Type:         result.BuildTool,
		Language:     result.Language,
		BootVersion:  result.BootVersion,
		BaseDir:      result.ProjectName,
		GroupID:      result.GroupID,
		ArtifactID:   result.ArtifactID,
		Name:         result.ProjectName,
		Description:  result.Description,
		PackageName:  packageName,
		Packaging:    result.Packaging,
		JavaVersion:  result.JavaVersion,
		Dependencies: result.Dependencies,
	}

	if err := client.DownloadProject(ctx, config, "."); err != nil {
		printer.StopSpinner(s, "")
		return fmt.Errorf("downloading project: %w", err)
	}
	printer.StopSpinner(s, "Downloaded and extracted!")

	gen := project.NewGenerator()

	// Step 2: Git init.
	if result.InitGit {
		s = printer.StartSpinner("Initializing git repository...")
		if err := gen.InitGit(targetDir); err != nil {
			printer.StopSpinner(s, "")
			printer.Warn("Git init failed: %v (continuing...)", err)
		} else {
			printer.StopSpinner(s, "Git repository initialized!")
		}
	}

	// Step 3: Docker files.
	if result.GenerateDocker {
		s = printer.StartSpinner("Generating Docker files...")
		if err := gen.GenerateDockerFiles(targetDir, result.BuildTool, result.JavaVersion, result.Dependencies); err != nil {
			printer.StopSpinner(s, "")
			printer.Warn("Docker file generation failed: %v", err)
		} else {
			printer.StopSpinner(s, "Docker files generated!")
		}
	}

	// Step 4: .env.example.
	if result.GenerateEnv {
		s = printer.StartSpinner("Generating .env.example...")
		if err := gen.GenerateEnvFile(targetDir, result.Dependencies); err != nil {
			printer.StopSpinner(s, "")
			printer.Warn(".env.example generation failed: %v", err)
		} else {
			printer.StopSpinner(s, ".env.example generated!")
		}
	}

	// Success output.
	fmt.Println()
	printer.Dim("  ─────────────────────────────────────────────")
	fmt.Println()
	printer.Success("Success! Created %s at ./%s", result.ProjectName, result.ProjectName)
	fmt.Println()
	printer.Info("Inside that directory, you can run:")
	fmt.Println()

	switch result.BuildTool {
	case "maven", "gradle", "gradle-kotlin":
		printer.Dim("    springcli start             Start the development server")
		printer.Dim("    springcli dev               Start with 'dev' profile & live-reload")
		printer.Dim("    springcli test              Run your tests")
		printer.Dim("    springcli package           Build executable JAR/WAR for production")
	}

	fmt.Println()
	printer.Info("Get started by typing:")
	fmt.Println()
	printer.Dim("    cd %s", result.ProjectName)
	printer.Dim("    springcli dev")

	fmt.Println()
	printer.Success("Happy coding! 🌱")
	fmt.Println()

	return nil
}
