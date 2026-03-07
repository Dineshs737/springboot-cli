package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/spf13/cobra"
	"github.com/springcli/springcli/internal/gradle"
	"github.com/springcli/springcli/internal/maven"
	"github.com/springcli/springcli/internal/project"
)

var infoCmd = &cobra.Command{
	Use:   "info",
	Short: "Show project metadata",
	Long:  `Parses the current Spring Boot project build file and displays key metadata.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runInfo()
	},
}

func init() {
	rootCmd.AddCommand(infoCmd)
}

func runInfo() error {
	detector := project.NewDetector()
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	buildPath, buildType, err := detector.GetBuildFilePath(cwd)
	if err != nil {
		return fmt.Errorf("detecting project: %w", err)
	}

	var bootVersion string
	var javaVersion string
	var packaging string
	var depsCount int
	var group string
	var artifact string

	if buildType == project.BuildTypeMaven {
		mp := maven.NewParser()
		doc, err := mp.ReadPom(buildPath)
		if err != nil {
			return err
		}

		root := doc.Root()
		if root != nil {
			if p := root.FindElement("parent"); p != nil {
				if v := p.FindElement("version"); v != nil {
					bootVersion = v.Text()
				}
			}

			if props := root.FindElement("properties"); props != nil {
				if jv := props.FindElement("java.version"); jv != nil {
					javaVersion = jv.Text()
				}
			}

			if pkg := root.FindElement("packaging"); pkg != nil {
				packaging = pkg.Text()
			} else {
				packaging = "jar"
			}

			if g := root.FindElement("groupId"); g != nil {
				group = g.Text()
			}
			if a := root.FindElement("artifactId"); a != nil {
				artifact = a.Text()
			}
		}

		deps, _ := mp.ListDependencies(buildPath)
		depsCount = len(deps)

	} else if buildType == project.BuildTypeGradleGroovy || buildType == project.BuildTypeGradleKotlin {
		content, err := os.ReadFile(buildPath)
		if err != nil {
			return err
		}
		text := string(content)

		// Boot version regex
		bootRe := regexp.MustCompile(`id\s*['"]org\.springframework\.boot['"]\s*version\s*['"]([^'"]+)['"]`)
		if m := bootRe.FindStringSubmatch(text); len(m) > 1 {
			bootVersion = m[1]
		}

		// Java version regex
		javaRe := regexp.MustCompile(`JavaLanguageVersion\.of\s*\(\s*['"]?([^'")]+)['"]?\s*\)`)
		if m := javaRe.FindStringSubmatch(text); len(m) > 1 {
			javaVersion = m[1]
		}

		if strings.Contains(text, "war") {
			packaging = "war"
		} else {
			packaging = "jar"
		}

		gp := gradle.NewParser()
		deps, _ := gp.ListDependencies(buildPath)
		depsCount = len(deps)

		// Group and artifact
		groupRe := regexp.MustCompile(`group\s*=\s*['"]([^'"]+)['"]`)
		if m := groupRe.FindStringSubmatch(text); len(m) > 1 {
			group = m[1]
		}

		// artifact is usually project folder name for gradle
		artifact = filepath.Base(cwd)
	}

	if bootVersion == "" {
		bootVersion = "Unknown"
	}
	if javaVersion == "" {
		javaVersion = "Unknown"
	}
	if group == "" {
		group = "Unknown"
	}
	if artifact == "" {
		artifact = "Unknown"
	}

	printer.Header("Project Information")

	headers := []string{"Property", "Value"}
	rows := [][]string{
		{"Build Tool", string(buildType)},
		{"Spring Boot", bootVersion},
		{"Java Version", javaVersion},
		{"Packaging", packaging},
		{"Dependencies", fmt.Sprintf("%d", depsCount)},
		{"Group ID", group},
		{"Artifact ID", artifact},
	}

	printer.Table(headers, rows)
	fmt.Println()

	return nil
}
