package wizard

import (
	"fmt"
	"regexp"
	"strings"
)

var (
	projectNameRegex = regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9-]*$`)
	groupIDRegex     = regexp.MustCompile(`^[a-z][a-z0-9]*(\.[a-z][a-z0-9]*)+$`)
	artifactIDRegex  = regexp.MustCompile(`^[a-z][a-z0-9-]*$`)
	packageNameRegex = regexp.MustCompile(`^[a-z][a-z0-9]*(\.[a-z][a-z0-9]*)+$`)
)

// ValidateProjectName ensures the name contains only letters, numbers, and hyphens.
func ValidateProjectName(val interface{}) error {
	s, ok := val.(string)
	if !ok {
		return fmt.Errorf("invalid input type")
	}
	s = strings.TrimSpace(s)
	if len(s) < 2 {
		return fmt.Errorf("project name must be at least 2 characters")
	}
	if len(s) > 64 {
		return fmt.Errorf("project name must be at most 64 characters")
	}
	if !projectNameRegex.MatchString(s) {
		return fmt.Errorf("project name can only contain letters, numbers, and hyphens (must start with a letter)")
	}
	return nil
}

// ValidateGroupID ensures the value is a valid Java package format (e.g., com.example).
func ValidateGroupID(val interface{}) error {
	s, ok := val.(string)
	if !ok {
		return fmt.Errorf("invalid input type")
	}
	s = strings.TrimSpace(s)
	if s == "" {
		return fmt.Errorf("group ID cannot be empty")
	}
	if !groupIDRegex.MatchString(s) {
		return fmt.Errorf("group ID must be in Java package format (e.g., com.example)")
	}
	return nil
}

// ValidateArtifactID ensures the value contains only lowercase letters, numbers, and hyphens.
func ValidateArtifactID(val interface{}) error {
	s, ok := val.(string)
	if !ok {
		return fmt.Errorf("invalid input type")
	}
	s = strings.TrimSpace(s)
	if s == "" {
		return fmt.Errorf("artifact ID cannot be empty")
	}
	if !artifactIDRegex.MatchString(s) {
		return fmt.Errorf("artifact ID can only contain lowercase letters, numbers, and hyphens")
	}
	return nil
}

// ValidatePackageName ensures the value is a valid Java package name (no hyphens).
func ValidatePackageName(val interface{}) error {
	s, ok := val.(string)
	if !ok {
		return fmt.Errorf("invalid input type")
	}
	s = strings.TrimSpace(s)
	if s == "" {
		return fmt.Errorf("package name cannot be empty")
	}
	if !packageNameRegex.MatchString(s) {
		return fmt.Errorf("package name must be in Java format (e.g., com.example.myapp) — no hyphens allowed")
	}
	return nil
}

// ValidateNotEmpty ensures the value is not blank.
func ValidateNotEmpty(val interface{}) error {
	s, ok := val.(string)
	if !ok {
		return fmt.Errorf("invalid input type")
	}
	if strings.TrimSpace(s) == "" {
		return fmt.Errorf("this field cannot be empty")
	}
	return nil
}

// DerivePackageName creates a Java package name from groupID and artifactID.
// Removes hyphens from artifactID.
func DerivePackageName(groupID, artifactID string) string {
	clean := strings.ReplaceAll(artifactID, "-", "")
	return groupID + "." + clean
}
