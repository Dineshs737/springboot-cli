package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/springcli/springcli/internal/project"
)

var doctorDir string

var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Check project for common Spring Boot issues",
	Long: `Scans your local Spring Boot project for misconfigurations,
missing annotations, dangerous field injections, and Java version mismatches.`,
	RunE: runDoctor,
}

func init() {
	doctorCmd.Flags().StringVar(&doctorDir, "dir", ".", "Target project directory")
	rootCmd.AddCommand(doctorCmd)
}

func runDoctor(cmd *cobra.Command, args []string) error {
	detector := project.NewDetector()
	buildPath, buildType, err := detector.GetBuildFilePath(doctorDir)
	if err != nil {
		return fmt.Errorf("detecting project: %w", err)
	}

	printer.Info("Running SpringCLI Doctor on %s project", buildType)
	fmt.Println()

	var issuesFound int

	// 1. Check Java Version
	if !checkJavaVersion(buildPath, buildType) {
		issuesFound++
	}

	// 2. Check @SpringBootApplication presence
	if !checkSpringBootApplication(doctorDir) {
		issuesFound++
	}

	// 3. Check for @Autowired field injection
	if !checkFieldInjection(doctorDir) {
		issuesFound++
	}

	fmt.Println()
	if issuesFound == 0 {
		printer.Success("✅ Doctor says: Your project looks perfectly healthy!")
	} else {
		printer.Warn("⚠️ Doctor says: Found %d potential issue(s) to address.", issuesFound)
	}

	return nil
}

// checkJavaVersion extracts project java version and compares it to `java -version`.
// Returns true if healthy, false if issue found.
func checkJavaVersion(buildPath string, buildType project.BuildType) bool {
	printRunningCheck("Checking Java version alignment")

	// Get local java version
	localCmd := exec.Command("java", "-version")
	out, err := localCmd.CombinedOutput()
	if err != nil {
		printFail("Could not execute 'java -version'. Is Java installed and on your PATH?")
		return false
	}

	localVersionStr := extractLocalJavaVersion(string(out))
	if localVersionStr == "" {
		printFail("Could not parse local Java version from output.")
		return false
	}

	localMajor := parseMajorVersion(localVersionStr)

	// Get project java version
	var projectVersionStr string
	contentBytes, err := os.ReadFile(buildPath)
	if err != nil {
		printFail("Could not read build file: %v", err)
		return false
	}
	content := string(contentBytes)

	if buildType == project.BuildTypeMaven {
		// Try to find <java.version>
		re := regexp.MustCompile(`(?i)<java\.version>(.+?)</java\.version>`)
		matches := re.FindStringSubmatch(content)
		if len(matches) > 1 {
			projectVersionStr = matches[1]
		}
	} else {
		// Gradle: try sourceCompatibility or toolchain
		reSource := regexp.MustCompile(`sourceCompatibility\s*=\s*['"]?([^'"\s]+)['"]?`)
		matches := reSource.FindStringSubmatch(content)
		if len(matches) > 1 {
			projectVersionStr = matches[1]
		} else {
			reToolchain := regexp.MustCompile(`languageVersion\s*=\s*JavaLanguageVersion\.of\(\s*(\d+)\s*\)`)
			matches = reToolchain.FindStringSubmatch(content)
			if len(matches) > 1 {
				projectVersionStr = matches[1]
			}
		}
	}

	if projectVersionStr == "" {
		printWarn("Could not find explicit Java version in build file. Assuming 17+")
		return true // not strictly a failure, just missing
	}

	projectMajor := parseMajorVersion(projectVersionStr)

	if localMajor != projectMajor {
		printFail("Java version mismatch! Project requests Java %d, but your environment is running Java %d.", projectMajor, localMajor)
		printer.Dim("   Local: %s", localVersionStr)
		printer.Dim("   Project: %s", projectVersionStr)
		return false
	}

	printPass("Java version %d aligns nicely.", localMajor)
	return true
}

func extractLocalJavaVersion(output string) string {
	// e.g. openjdk version "21.0.1" 2023-10-17
	// e.g. java version "1.8.0_391"
	re := regexp.MustCompile(`version "([^"]+)"`)
	matches := re.FindStringSubmatch(output)
	if len(matches) > 1 {
		return matches[1]
	}
	return ""
}

func parseMajorVersion(version string) int {
	v := strings.TrimSpace(version)
	v = strings.TrimPrefix(v, "1.") // Java 1.8 -> 8
	parts := strings.Split(v, ".")
	if len(parts) > 0 {
		var majorStr string
		if strings.Contains(parts[0], "_") {
			majorStr = strings.Split(parts[0], "_")[0]
		} else {
			majorStr = parts[0]
		}

		val, err := strconv.Atoi(majorStr)
		if err == nil {
			return val
		}
	}
	return 0
}

// checkSpringBootApplication scans the src tree for exactly ONE @SpringBootApplication
func checkSpringBootApplication(dir string) bool {
	printRunningCheck("Checking @SpringBootApplication entry point")

	srcDirs := []string{
		filepath.Join(dir, "src", "main", "java"),
		filepath.Join(dir, "src", "main", "kotlin"),
	}

	count := 0
	for _, srcDir := range srcDirs {
		_ = filepath.Walk(srcDir, func(path string, info os.FileInfo, err error) error {
			if err != nil || info.IsDir() {
				return nil
			}
			if strings.HasSuffix(path, ".java") || strings.HasSuffix(path, ".kt") {
				content, err := os.ReadFile(path)
				if err == nil {
					if strings.Contains(string(content), "@SpringBootApplication") {
						count++
					}
				}
			}
			return nil
		})
	}

	if count == 0 {
		printFail("Missing @SpringBootApplication annotation.")
		printer.Dim("   No class found with this annotation. Your application probably won't start.")
		return false
	} else if count > 1 {
		printFail("Found %d classes annotated with @SpringBootApplication.", count)
		printer.Dim("   Multiple entry points can cause Spring context conflicts. Keep only one.")
		return false
	}

	printPass("Found exactly 1 @SpringBootApplication class.")
	return true
}

// checkFieldInjection looks for @Autowired on fields instead of constructors
func checkFieldInjection(dir string) bool {
	printRunningCheck("Checking for dangerous field injections")

	srcDirs := []string{
		filepath.Join(dir, "src", "main", "java"),
		filepath.Join(dir, "src", "main", "kotlin"),
	}

	// Regex looks for @Autowired followed by an access modifier and type on the next line or same line, NOT followed by a parenthesis (method)
	// Rough approximation, works for most generic field injections
	reFieldJava := regexp.MustCompile(`@Autowired[\s\n]+(?:private|protected|public|[\w<>]+)\s+[\w<>]+\s+\w+\s*;`)
	reFieldKotlin := regexp.MustCompile(`@Autowired[\s\n]+(?:lateinit\s+)?var\s+\w+\s*:`)

	var flaggedFiles []string

	for _, srcDir := range srcDirs {
		_ = filepath.Walk(srcDir, func(path string, info os.FileInfo, err error) error {
			if err != nil || info.IsDir() {
				return nil
			}
			isJava := strings.HasSuffix(path, ".java")
			isKotlin := strings.HasSuffix(path, ".kt")

			if isJava || isKotlin {
				contentBytes, err := os.ReadFile(path)
				if err == nil {
					content := string(contentBytes)
					hasFieldInjection := false

					if isJava {
						if reFieldJava.MatchString(content) {
							hasFieldInjection = true
						}
					} else if isKotlin {
						if reFieldKotlin.MatchString(content) {
							hasFieldInjection = true
						}
					}

					if hasFieldInjection {
						// Record relative path
						rel, _ := filepath.Rel(dir, path)
						flaggedFiles = append(flaggedFiles, rel)
					}
				}
			}
			return nil
		})
	}

	if len(flaggedFiles) > 0 {
		printFail("Found field injection (@Autowired) in %d file(s).", len(flaggedFiles))
		printer.Dim("   Field injection makes testing difficult and hides circular dependencies.")
		printer.Dim("   Consider using Constructor Injection instead.")
		for _, f := range flaggedFiles {
			printer.Dim("   - %s", f)
		}
		return false
	}

	printPass("No field injections detected. Good job using constructors!")
	return true
}

// Helpers to format output consistently
func printRunningCheck(msg string) {
	fmt.Printf("🔍 %s...\n", msg)
}

func printPass(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	fmt.Printf("   ✔️  %s\n", msg)
}

func printFail(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	fmt.Printf("   ❌ %s\n", msg)
}

func printWarn(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	fmt.Printf("   ⚠️  %s\n", msg)
}
