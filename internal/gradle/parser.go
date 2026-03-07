package gradle

import (
	"fmt"
	"os"
	"regexp"
	"strings"
)

// GradleDependency represents a dependency in a Gradle build file.
type GradleDependency struct {
	Configuration string // implementation, testImplementation, runtimeOnly, etc.
	Group         string
	Artifact      string
	Version       string
	RawLine       string
}

// Parser handles reading and writing Gradle build files.
type Parser struct{}

// NewParser creates a new Gradle Parser.
func NewParser() *Parser {
	return &Parser{}
}

// IsKotlinDSL detects if the given file path is a Kotlin DSL build file.
func (p *Parser) IsKotlinDSL(path string) bool {
	return strings.HasSuffix(path, ".gradle.kts")
}

// ListDependencies lists all dependencies in a Gradle build file.
func (p *Parser) ListDependencies(path string) ([]GradleDependency, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading gradle file %s: %w", path, err)
	}

	return parseDependencies(string(content), p.IsKotlinDSL(path)), nil
}

// AddDependency adds a dependency to the Gradle build file.
// Returns false if the dependency already exists.
func (p *Parser) AddDependency(path string, group, artifact, version, scope string) (bool, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return false, fmt.Errorf("reading gradle file %s: %w", path, err)
	}

	isKotlin := p.IsKotlinDSL(path)
	lines := string(content)

	// Check if dependency already exists.
	existing := parseDependencies(lines, isKotlin)
	for _, dep := range existing {
		if dep.Group == group && dep.Artifact == artifact {
			return false, nil
		}
	}

	config := mapScopeToConfiguration(scope)
	depLine := formatDependencyLine(config, group, artifact, version, isKotlin)

	// Find the closing brace of the dependencies block and insert before it.
	newContent, inserted := insertDependency(lines, depLine)
	if !inserted {
		return false, fmt.Errorf("could not find dependencies block in %s", path)
	}

	if err := os.WriteFile(path, []byte(newContent), 0644); err != nil {
		return false, fmt.Errorf("writing gradle file %s: %w", path, err)
	}

	return true, nil
}

// RemoveDependency removes a dependency from the Gradle build file.
// Accepts both full artifact ID and short form.
// Returns true if the dependency was found and removed.
func (p *Parser) RemoveDependency(path string, artifactOrShort string) (bool, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return false, fmt.Errorf("reading gradle file %s: %w", path, err)
	}

	lines := strings.Split(string(content), "\n")
	candidates := []string{
		"spring-boot-starter-" + artifactOrShort,
		artifactOrShort,
	}

	found := false
	var newLines []string
	for _, line := range lines {
		shouldRemove := false
		for _, candidate := range candidates {
			if strings.Contains(line, candidate) && isDependencyLine(line) {
				shouldRemove = true
				found = true
				break
			}
		}
		if !shouldRemove {
			newLines = append(newLines, line)
		}
	}

	if !found {
		return false, nil
	}

	if err := os.WriteFile(path, []byte(strings.Join(newLines, "\n")), 0644); err != nil {
		return false, fmt.Errorf("writing gradle file %s: %w", path, err)
	}

	return true, nil
}

// HasDependency checks if a dependency exists in the Gradle build file.
func (p *Parser) HasDependency(path string, artifact string) (bool, error) {
	deps, err := p.ListDependencies(path)
	if err != nil {
		return false, err
	}
	for _, dep := range deps {
		if strings.EqualFold(dep.Artifact, artifact) {
			return true, nil
		}
	}
	return false, nil
}

// parseDependencies extracts dependencies from Gradle build file content.
func parseDependencies(content string, isKotlin bool) []GradleDependency {
	var deps []GradleDependency

	var re *regexp.Regexp
	if isKotlin {
		// Kotlin DSL: implementation("group:artifact:version") or implementation("group:artifact")
		re = regexp.MustCompile(`(\w+)\("([^"]+)"\)`)
	} else {
		// Groovy DSL: implementation 'group:artifact:version' or implementation 'group:artifact'
		re = regexp.MustCompile(`(\w+)\s+'([^']+)'`)
	}

	lines := strings.Split(content, "\n")
	inDeps := false
	braceCount := 0

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		if strings.HasPrefix(trimmed, "dependencies") && strings.Contains(trimmed, "{") {
			inDeps = true
			braceCount = 1
			continue
		}

		if inDeps {
			braceCount += strings.Count(trimmed, "{") - strings.Count(trimmed, "}")
			if braceCount <= 0 {
				inDeps = false
				continue
			}

			matches := re.FindStringSubmatch(trimmed)
			if len(matches) >= 3 {
				config := matches[1]
				coords := matches[2]
				parts := strings.Split(coords, ":")

				dep := GradleDependency{
					Configuration: config,
					RawLine:       trimmed,
				}

				if len(parts) >= 2 {
					dep.Group = parts[0]
					dep.Artifact = parts[1]
				}
				if len(parts) >= 3 {
					dep.Version = parts[2]
				}

				deps = append(deps, dep)
			}
		}
	}

	return deps
}

// mapScopeToConfiguration maps Maven scopes to Gradle configurations.
func mapScopeToConfiguration(scope string) string {
	switch strings.ToLower(scope) {
	case "test":
		return "testImplementation"
	case "provided":
		return "compileOnly"
	case "runtime":
		return "runtimeOnly"
	case "compile", "":
		return "implementation"
	default:
		return "implementation"
	}
}

// formatDependencyLine creates a dependency line string.
func formatDependencyLine(config, group, artifact, version string, isKotlin bool) string {
	coords := group + ":" + artifact
	if version != "" {
		coords += ":" + version
	}

	if isKotlin {
		return fmt.Sprintf("    %s(\"%s\")", config, coords)
	}
	return fmt.Sprintf("    %s '%s'", config, coords)
}

// insertDependency inserts a dependency line into the dependencies block.
func insertDependency(content, depLine string) (string, bool) {
	lines := strings.Split(content, "\n")
	inDeps := false
	braceCount := 0
	insertIdx := -1

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)

		if strings.HasPrefix(trimmed, "dependencies") && strings.Contains(trimmed, "{") {
			inDeps = true
			braceCount = 1
			continue
		}

		if inDeps {
			braceCount += strings.Count(trimmed, "{") - strings.Count(trimmed, "}")
			if braceCount <= 0 {
				insertIdx = i
				break
			}
		}
	}

	if insertIdx == -1 {
		return content, false
	}

	// Insert before the closing brace.
	newLines := make([]string, 0, len(lines)+1)
	newLines = append(newLines, lines[:insertIdx]...)
	newLines = append(newLines, depLine)
	newLines = append(newLines, lines[insertIdx:]...)

	return strings.Join(newLines, "\n"), true
}

// isDependencyLine checks if a line looks like a dependency declaration.
func isDependencyLine(line string) bool {
	trimmed := strings.TrimSpace(line)
	configs := []string{
		"implementation", "testImplementation", "runtimeOnly",
		"compileOnly", "api", "annotationProcessor",
	}
	for _, c := range configs {
		if strings.HasPrefix(trimmed, c) {
			return true
		}
	}
	return false
}
