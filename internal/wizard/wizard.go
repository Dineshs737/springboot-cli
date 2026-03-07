package wizard

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"golang.org/x/term"

	"github.com/springcli/springcli/internal/initializr"
	"github.com/springcli/springcli/internal/ui"
)

// WizardResult holds all answers collected from the interactive wizard.
type WizardResult struct {
	ProjectName  string
	GroupID      string
	ArtifactID   string
	Description  string
	PackageName  string
	BuildTool    string // "maven", "gradle", "gradle-kotlin"
	Language     string // "java", "kotlin", "groovy"
	JavaVersion  string // "17", "21"
	Packaging    string // "jar", "war"
	BootVersion  string // e.g. "3.3.0"
	Dependencies []string
	InitGit      bool
	GenerateEnv  bool
	GenerateDocker bool
}

// NonInteractiveConfig holds flag values for non-interactive mode.
type NonInteractiveConfig struct {
	ProjectName  string
	GroupID      string
	ArtifactID   string
	Description  string
	Language     string
	BuildType    string
	BootVersion  string
	JavaVersion  string
	Packaging    string
	Dependencies []string
	InitGit      bool
	GenerateEnv  bool
	GenerateDocker bool
}

// IsInteractive returns true if stdin is a TTY (not piped).
func IsInteractive() bool {
	return term.IsTerminal(int(os.Stdin.Fd()))
}

// RunWizard orchestrates the full interactive wizard flow.
func RunWizard(projectNameArg string, client initializr.Client) (*WizardResult, error) {
	ctx := context.Background()
	printer := ui.NewPrinter()

	result := &WizardResult{
		InitGit:        true,
		GenerateEnv:    true,
		GenerateDocker: true,
	}

	// Step 1-5: Project info text inputs.
	projectInfo := struct {
		ProjectName string `survey:"projectName"`
		GroupID     string `survey:"groupID"`
		ArtifactID  string `survey:"artifactID"`
		Description string `survey:"description"`
		PackageName string `survey:"packageName"`
	}{}

	questions := BuildProjectInfoQuestions(projectNameArg)
	if err := survey.Ask(questions, &projectInfo); err != nil {
		return nil, fmt.Errorf("project info prompt: %w", err)
	}

	result.ProjectName = projectInfo.ProjectName
	result.GroupID = projectInfo.GroupID
	result.ArtifactID = projectInfo.ArtifactID
	result.Description = projectInfo.Description
	result.PackageName = projectInfo.PackageName

	// Step 6: Build tool.
	var buildTool string
	if err := survey.AskOne(BuildToolQuestion(), &buildTool); err != nil {
		return nil, fmt.Errorf("build tool prompt: %w", err)
	}
	result.BuildTool = MapBuildToolToType(buildTool)

	// Step 7: Language.
	var language string
	if err := survey.AskOne(LanguageQuestion(), &language); err != nil {
		return nil, fmt.Errorf("language prompt: %w", err)
	}
	result.Language = MapLanguageToAPI(language)

	// Step 8: Java version.
	var javaVersion string
	if err := survey.AskOne(JavaVersionQuestion(), &javaVersion); err != nil {
		return nil, fmt.Errorf("java version prompt: %w", err)
	}
	result.JavaVersion = javaVersion

	// Step 9: Packaging.
	var packaging string
	if err := survey.AskOne(PackagingQuestion(), &packaging); err != nil {
		return nil, fmt.Errorf("packaging prompt: %w", err)
	}
	result.Packaging = strings.ToLower(packaging)

	// Step 10: Spring Boot version (fetched live).
	spinner := printer.StartSpinner("Fetching latest dependency catalog...")
	bootQ, err := BuildBootVersionQuestion(ctx, client)
	printer.StopSpinner(spinner, "Dependency catalog loaded!")
	if err != nil {
		printer.Warn("Could not fetch boot versions, using defaults")
		result.BootVersion = "3.2.3"
	} else {
		var bootVersion string
		if err := survey.AskOne(bootQ, &bootVersion); err != nil {
			return nil, fmt.Errorf("boot version prompt: %w", err)
		}
		result.BootVersion = CleanBootVersion(bootVersion)
	}

	// Step 11: Dependencies (fetched live, grouped by category).
	depQ, depIDs, err := BuildDependencyQuestion(ctx, client)
	if err != nil {
		printer.Warn("Could not fetch dependencies from start.spring.io")
	} else {
		var selectedLabels []string
		if err := survey.AskOne(depQ, &selectedLabels); err != nil {
			return nil, fmt.Errorf("dependencies prompt: %w", err)
		}

		// Map selected labels back to dep IDs, filtering out dividers.
		result.Dependencies = filterSelectedDeps(selectedLabels, depQ.Options, depIDs)
	}

	// Step 12: Git init.
	if err := survey.AskOne(GitInitQuestion(), &result.InitGit); err != nil {
		return nil, fmt.Errorf("git init prompt: %w", err)
	}

	// Step 13: .env.example.
	if err := survey.AskOne(EnvFileQuestion(), &result.GenerateEnv); err != nil {
		return nil, fmt.Errorf("env file prompt: %w", err)
	}

	// Step 14: Docker files.
	if err := survey.AskOne(DockerQuestion(), &result.GenerateDocker); err != nil {
		return nil, fmt.Errorf("docker prompt: %w", err)
	}

	return result, nil
}

// FromNonInteractive creates a WizardResult from non-interactive flag values.
func FromNonInteractive(cfg NonInteractiveConfig) *WizardResult {
	artifactID := cfg.ArtifactID
	if artifactID == "" {
		artifactID = cfg.ProjectName
	}
	packageName := DerivePackageName(cfg.GroupID, artifactID)
	description := cfg.Description
	if description == "" {
		description = "Generated by SpringCLI"
	}
	bootVersion := cfg.BootVersion
	if bootVersion == "" {
		bootVersion = "3.2.3"
	}
	javaVersion := cfg.JavaVersion
	if javaVersion == "" {
		javaVersion = "21"
	}

	return &WizardResult{
		ProjectName:    cfg.ProjectName,
		GroupID:        cfg.GroupID,
		ArtifactID:     artifactID,
		Description:    description,
		PackageName:    packageName,
		BuildTool:      cfg.BuildType,
		Language:       cfg.Language,
		JavaVersion:    javaVersion,
		Packaging:      cfg.Packaging,
		BootVersion:    bootVersion,
		Dependencies:   cfg.Dependencies,
		InitGit:        cfg.InitGit,
		GenerateEnv:    cfg.GenerateEnv,
		GenerateDocker: cfg.GenerateDocker,
	}
}

// filterSelectedDeps maps selected option labels back to dependency IDs.
func filterSelectedDeps(selectedLabels []string, allOptions []string, depIDs []string) []string {
	selectedSet := make(map[string]bool)
	for _, label := range selectedLabels {
		selectedSet[label] = true
	}

	var result []string
	for i, opt := range allOptions {
		if selectedSet[opt] && i < len(depIDs) && depIDs[i] != "" {
			result = append(result, depIDs[i])
		}
	}
	return result
}

// RunSecurityWizard runs the mini-wizard for security configuration.
func RunSecurityWizard() (secType string, generateConfig bool, err error) {
	if err := survey.AskOne(SecurityTypeQuestion(), &secType); err != nil {
		return "", false, fmt.Errorf("security type prompt: %w", err)
	}

	if err := survey.AskOne(SecurityConfigQuestion(), &generateConfig); err != nil {
		return "", false, fmt.Errorf("security config prompt: %w", err)
	}

	return secType, generateConfig, nil
}
