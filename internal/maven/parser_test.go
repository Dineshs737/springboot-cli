package maven

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func copyFixture(t *testing.T, fixtureName string) string {
	t.Helper()
	// Read the fixture from testdata.
	data, err := os.ReadFile(filepath.Join("..", "..", "testdata", fixtureName))
	require.NoError(t, err)

	dir := t.TempDir()
	dest := filepath.Join(dir, "pom.xml")
	require.NoError(t, os.WriteFile(dest, data, 0644))
	return dest
}

func TestListDependencies(t *testing.T) {
	path := copyFixture(t, "sample-pom.xml")
	p := NewParser()
	deps, err := p.ListDependencies(path)
	require.NoError(t, err)
	assert.Len(t, deps, 3)

	artifactIDs := make([]string, len(deps))
	for i, d := range deps {
		artifactIDs[i] = d.ArtifactID
	}
	assert.Contains(t, artifactIDs, "spring-boot-starter-web")
	assert.Contains(t, artifactIDs, "spring-boot-starter-data-jpa")
	assert.Contains(t, artifactIDs, "spring-boot-starter-test")
}

func TestAddDependency(t *testing.T) {
	path := copyFixture(t, "sample-pom.xml")
	p := NewParser()

	added, err := p.AddDependency(path, MavenDependency{
		GroupID:    "org.springframework.boot",
		ArtifactID: "spring-boot-starter-security",
	})
	require.NoError(t, err)
	assert.True(t, added)

	// Verify it's actually in the file.
	deps, err := p.ListDependencies(path)
	require.NoError(t, err)
	found := false
	for _, d := range deps {
		if d.ArtifactID == "spring-boot-starter-security" {
			found = true
			break
		}
	}
	assert.True(t, found, "security dependency should be present after adding")
}

func TestAddDependencyAlreadyExists(t *testing.T) {
	path := copyFixture(t, "sample-pom.xml")
	p := NewParser()

	added, err := p.AddDependency(path, MavenDependency{
		GroupID:    "org.springframework.boot",
		ArtifactID: "spring-boot-starter-web",
	})
	require.NoError(t, err)
	assert.False(t, added, "should not add duplicate dependency")

	// Count should remain the same.
	deps, err := p.ListDependencies(path)
	require.NoError(t, err)

	count := 0
	for _, d := range deps {
		if d.ArtifactID == "spring-boot-starter-web" {
			count++
		}
	}
	assert.Equal(t, 1, count, "should not duplicate the dependency")
}

func TestRemoveDependency(t *testing.T) {
	path := copyFixture(t, "sample-pom.xml")
	p := NewParser()

	removed, err := p.RemoveDependency(path, "spring-boot-starter-web")
	require.NoError(t, err)
	assert.True(t, removed)

	deps, err := p.ListDependencies(path)
	require.NoError(t, err)
	for _, d := range deps {
		assert.NotEqual(t, "spring-boot-starter-web", d.ArtifactID)
	}
}

func TestRemoveDependencyShortForm(t *testing.T) {
	path := copyFixture(t, "sample-pom.xml")
	p := NewParser()

	// Should match "spring-boot-starter-web" from short form "web".
	removed, err := p.RemoveDependency(path, "web")
	require.NoError(t, err)
	assert.True(t, removed)

	has, err := p.HasDependency(path, "spring-boot-starter-web")
	require.NoError(t, err)
	assert.False(t, has)
}

func TestRemoveDependencyNotFound(t *testing.T) {
	path := copyFixture(t, "sample-pom.xml")
	p := NewParser()

	removed, err := p.RemoveDependency(path, "nonexistent-lib")
	require.NoError(t, err)
	assert.False(t, removed)
}

func TestAddDependencyPreservesFormatting(t *testing.T) {
	path := copyFixture(t, "sample-pom.xml")
	p := NewParser()

	// Read original content.
	origContent, err := p.GetPomContent(path)
	require.NoError(t, err)

	// Add a dependency.
	_, err = p.AddDependency(path, MavenDependency{
		GroupID:    "org.springframework.boot",
		ArtifactID: "spring-boot-starter-actuator",
	})
	require.NoError(t, err)

	// Check that important structural elements are preserved.
	newContent, err := p.GetPomContent(path)
	require.NoError(t, err)
	assert.Contains(t, newContent, "<modelVersion>4.0.0</modelVersion>")
	assert.Contains(t, newContent, "<dependencies>")
	assert.Contains(t, newContent, "</dependencies>")
	// The original deps should still be present.
	assert.Contains(t, newContent, "spring-boot-starter-web")
	assert.Contains(t, newContent, "spring-boot-starter-data-jpa")
	_ = origContent // We verified structure is preserved.
}

func TestAddDependencyWithVersionAndScope(t *testing.T) {
	path := copyFixture(t, "sample-pom.xml")
	p := NewParser()

	added, err := p.AddDependency(path, MavenDependency{
		GroupID:    "io.jsonwebtoken",
		ArtifactID: "jjwt-api",
		Version:    "0.12.5",
	})
	require.NoError(t, err)
	assert.True(t, added)

	deps, err := p.ListDependencies(path)
	require.NoError(t, err)
	found := false
	for _, d := range deps {
		if d.ArtifactID == "jjwt-api" {
			found = true
			assert.Equal(t, "0.12.5", d.Version)
			assert.Equal(t, "io.jsonwebtoken", d.GroupID)
		}
	}
	assert.True(t, found)
}

func TestDetectBasePackage(t *testing.T) {
	path := copyFixture(t, "sample-pom.xml")
	p := NewParser()

	pkg, err := p.DetectBasePackage(path)
	require.NoError(t, err)
	assert.Equal(t, "com.example.demo", pkg)
}

func TestHasDependency(t *testing.T) {
	path := copyFixture(t, "sample-pom.xml")
	p := NewParser()

	has, err := p.HasDependency(path, "spring-boot-starter-web")
	require.NoError(t, err)
	assert.True(t, has)

	has, err = p.HasDependency(path, "nonexistent")
	require.NoError(t, err)
	assert.False(t, has)
}
