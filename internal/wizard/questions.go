package wizard

import (
	"context"
	"fmt"

	"github.com/AlecAivazis/survey/v2"
	"github.com/springcli/springcli/internal/initializr"
)

// BuildProjectInfoQuestions creates the text input questions for project info.
func BuildProjectInfoQuestions(projectName string) []*survey.Question {
	defaultArtifact := projectName
	defaultPackage := DerivePackageName("com.example", projectName)

	return []*survey.Question{
		{
			Name: "projectName",
			Prompt: &survey.Input{
				Message: "Project name:",
				Default: projectName,
			},
			Validate: ValidateProjectName,
		},
		{
			Name: "groupID",
			Prompt: &survey.Input{
				Message: "Group ID:",
				Default: "com.example",
			},
			Validate: ValidateGroupID,
		},
		{
			Name: "artifactID",
			Prompt: &survey.Input{
				Message: "Artifact ID:",
				Default: defaultArtifact,
			},
			Validate: ValidateArtifactID,
		},
		{
			Name: "description",
			Prompt: &survey.Input{
				Message: "Project description:",
				Default: "My awesome Spring Boot app",
			},
			Validate: ValidateNotEmpty,
		},
		{
			Name: "packageName",
			Prompt: &survey.Input{
				Message: "Package name:",
				Default: defaultPackage,
			},
			Validate: ValidatePackageName,
		},
	}
}

// BuildToolQuestion creates the build tool selection prompt.
func BuildToolQuestion() *survey.Select {
	return &survey.Select{
		Message: "Select build tool:",
		Options: []string{"Maven", "Gradle (Groovy DSL)", "Gradle (Kotlin DSL)"},
		Default: "Maven",
	}
}

// LanguageQuestion creates the language selection prompt.
func LanguageQuestion() *survey.Select {
	return &survey.Select{
		Message: "Select language:",
		Options: []string{"Java", "Kotlin", "Groovy"},
		Default: "Java",
	}
}

// JavaVersionQuestion creates the Java version selection prompt.
func JavaVersionQuestion() *survey.Select {
	return &survey.Select{
		Message: "Select Java version:",
		Options: []string{"21", "17"},
		Default: "21",
	}
}

// PackagingQuestion creates the packaging selection prompt.
func PackagingQuestion() *survey.Select {
	return &survey.Select{
		Message: "Select packaging:",
		Options: []string{"JAR", "WAR"},
		Default: "JAR",
	}
}

// BuildBootVersionQuestion fetches available boot versions from the API and builds a select.
func BuildBootVersionQuestion(ctx context.Context, client initializr.Client) (*survey.Select, error) {
	meta, err := client.FetchMetadata(ctx)
	if err != nil {
		return nil, fmt.Errorf("fetching boot versions: %w", err)
	}

	var options []string
	defaultVer := meta.BootVersion.Default

	for i, v := range meta.BootVersion.Values {
		label := v.Name
		if i == 0 {
			label += " (latest)"
		}
		options = append(options, label)
	}

	if len(options) == 0 {
		options = []string{"3.3.0 (latest)", "3.2.3", "3.1.10"}
		defaultVer = "3.3.0 (latest)"
	} else {
		// Mark the default.
		for i, opt := range options {
			if meta.BootVersion.Values[i].ID == defaultVer {
				defaultVer = opt
				break
			}
		}
	}

	return &survey.Select{
		Message: "Spring Boot version:",
		Options: options,
		Default: defaultVer,
	}, nil
}

// BuildDependencyQuestion fetches the live dependency catalog and builds a multi-select
// with category group dividers.
func BuildDependencyQuestion(ctx context.Context, client initializr.Client) (*survey.MultiSelect, []string, error) {
	categories, err := client.ListAllDependencies(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("fetching dependencies: %w", err)
	}

	var options []string
	var depIDs []string // maps option index -> actual dep ID (empty for dividers)

	// Common defaults to pre-select.
	preSelected := map[string]bool{
		"web":      true,
		"actuator": true,
		"lombok":   true,
	}

	var defaults []string

	for _, cat := range categories {
		divider := fmt.Sprintf("──── %s ────", cat.Name)
		options = append(options, divider)
		depIDs = append(depIDs, "") // divider, no ID

		for _, dep := range cat.Values {
			label := fmt.Sprintf("%s — %s", dep.Name, dep.Description)
			if len(label) > 70 {
				label = label[:67] + "..."
			}
			options = append(options, label)
			depIDs = append(depIDs, dep.ID)

			if preSelected[dep.ID] {
				defaults = append(defaults, label)
			}
		}
	}

	return &survey.MultiSelect{
		Message:  "Select dependencies (space to select, enter to confirm):",
		Options:  options,
		Default:  defaults,
		PageSize: 20,
	}, depIDs, nil
}

// GitInitQuestion creates the git init confirmation prompt.
func GitInitQuestion() *survey.Confirm {
	return &survey.Confirm{
		Message: "Initialize Git repository?",
		Default: true,
	}
}

// EnvFileQuestion creates the .env.example generation confirmation prompt.
func EnvFileQuestion() *survey.Confirm {
	return &survey.Confirm{
		Message: "Generate .env.example file?",
		Default: true,
	}
}

// DockerQuestion creates the Docker files generation confirmation prompt.
func DockerQuestion() *survey.Confirm {
	return &survey.Confirm{
		Message: "Generate Docker files?",
		Default: true,
	}
}

// SecurityTypeQuestion creates the security mini-wizard prompt.
func SecurityTypeQuestion() *survey.Select {
	return &survey.Select{
		Message: "Security setup type:",
		Options: []string{
			"Basic HTTP Security",
			"JWT (Stateless API)",
			"OAuth2 Resource Server",
			"JWT + OAuth2",
		},
		Default: "Basic HTTP Security",
	}
}

// SecurityConfigQuestion creates the SecurityConfig generation prompt.
func SecurityConfigQuestion() *survey.Confirm {
	return &survey.Confirm{
		Message: "Generate SecurityConfig class?",
		Default: true,
	}
}

// MapBuildToolToType converts the UI label back to the API type string.
func MapBuildToolToType(label string) string {
	switch label {
	case "Gradle (Groovy DSL)":
		return "gradle"
	case "Gradle (Kotlin DSL)":
		return "gradle-kotlin"
	default:
		return "maven"
	}
}

// MapLanguageToAPI converts the UI label back to the API language string.
func MapLanguageToAPI(label string) string {
	switch label {
	case "Kotlin":
		return "kotlin"
	case "Groovy":
		return "groovy"
	default:
		return "java"
	}
}

// CleanBootVersion strips the "(latest)" suffix from a version label.
func CleanBootVersion(label string) string {
	// Trim any suffix like " (latest)".
	for i, ch := range label {
		if ch == ' ' {
			return label[:i]
		}
	}
	return label
}
