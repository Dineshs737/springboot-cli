package project

import (
	"fmt"
	"os"
	"path/filepath"
)

// Detector detects the build system used in a given directory.
type Detector struct{}

// NewDetector creates a new Detector.
func NewDetector() *Detector {
	return &Detector{}
}

// Detect returns the BuildType for the given directory by checking for
// pom.xml, build.gradle.kts, or build.gradle.
func (d *Detector) Detect(dir string) (BuildType, error) {
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return BuildTypeUnknown, fmt.Errorf("resolving directory path: %w", err)
	}

	info, err := os.Stat(absDir)
	if err != nil {
		return BuildTypeUnknown, fmt.Errorf("accessing directory %s: %w", absDir, err)
	}
	if !info.IsDir() {
		return BuildTypeUnknown, fmt.Errorf("%s is not a directory", absDir)
	}

	// Check for pom.xml first (Maven takes priority).
	if fileExists(filepath.Join(absDir, "pom.xml")) {
		return BuildTypeMaven, nil
	}

	// Check for build.gradle.kts (Kotlin DSL).
	if fileExists(filepath.Join(absDir, "build.gradle.kts")) {
		return BuildTypeGradleKotlin, nil
	}

	// Check for build.gradle (Groovy DSL).
	if fileExists(filepath.Join(absDir, "build.gradle")) {
		return BuildTypeGradleGroovy, nil
	}

	return BuildTypeUnknown, nil
}

// GetBuildFilePath returns the path to the build file for the given directory.
func (d *Detector) GetBuildFilePath(dir string) (string, BuildType, error) {
	buildType, err := d.Detect(dir)
	if err != nil {
		return "", BuildTypeUnknown, fmt.Errorf("detecting build type: %w", err)
	}

	absDir, err := filepath.Abs(dir)
	if err != nil {
		return "", BuildTypeUnknown, fmt.Errorf("resolving directory path: %w", err)
	}

	switch buildType {
	case BuildTypeMaven:
		return filepath.Join(absDir, "pom.xml"), buildType, nil
	case BuildTypeGradleKotlin:
		return filepath.Join(absDir, "build.gradle.kts"), buildType, nil
	case BuildTypeGradleGroovy:
		return filepath.Join(absDir, "build.gradle"), buildType, nil
	default:
		return "", BuildTypeUnknown, fmt.Errorf("no build file found in %s", absDir)
	}
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return !info.IsDir()
}
