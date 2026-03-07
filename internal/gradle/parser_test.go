package gradle

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func copyGradleFixture(t *testing.T, fixtureName string) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("..", "..", "testdata", fixtureName))
	require.NoError(t, err)

	dir := t.TempDir()
	dest := filepath.Join(dir, fixtureName)
	require.NoError(t, os.WriteFile(dest, data, 0644))
	return dest
}

func TestGroovyListDependencies(t *testing.T) {
	path := copyGradleFixture(t, "sample-build.gradle")
	p := NewParser()
	deps, err := p.ListDependencies(path)
	require.NoError(t, err)
	assert.Len(t, deps, 4)

	artifacts := make([]string, len(deps))
	for i, d := range deps {
		artifacts[i] = d.Artifact
	}
	assert.Contains(t, artifacts, "spring-boot-starter-web")
	assert.Contains(t, artifacts, "spring-boot-starter-data-jpa")
	assert.Contains(t, artifacts, "postgresql")
	assert.Contains(t, artifacts, "spring-boot-starter-test")
}

func TestKotlinListDependencies(t *testing.T) {
	path := copyGradleFixture(t, "sample-build.gradle.kts")
	p := NewParser()
	deps, err := p.ListDependencies(path)
	require.NoError(t, err)
	assert.Len(t, deps, 4)

	artifacts := make([]string, len(deps))
	for i, d := range deps {
		artifacts[i] = d.Artifact
	}
	assert.Contains(t, artifacts, "spring-boot-starter-web")
	assert.Contains(t, artifacts, "spring-boot-starter-data-jpa")
}

func TestGroovyAddDependency(t *testing.T) {
	path := copyGradleFixture(t, "sample-build.gradle")
	p := NewParser()

	added, err := p.AddDependency(path, "org.springframework.boot", "spring-boot-starter-security", "", "compile")
	require.NoError(t, err)
	assert.True(t, added)

	deps, err := p.ListDependencies(path)
	require.NoError(t, err)

	found := false
	for _, d := range deps {
		if d.Artifact == "spring-boot-starter-security" {
			found = true
			assert.Equal(t, "implementation", d.Configuration)
		}
	}
	assert.True(t, found)

	// Verify Groovy uses single quotes.
	content, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Contains(t, string(content), "'org.springframework.boot:spring-boot-starter-security'")
}

func TestKotlinAddDependency(t *testing.T) {
	path := copyGradleFixture(t, "sample-build.gradle.kts")
	p := NewParser()

	added, err := p.AddDependency(path, "org.springframework.boot", "spring-boot-starter-security", "", "compile")
	require.NoError(t, err)
	assert.True(t, added)

	// Verify Kotlin DSL uses double quotes.
	content, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Contains(t, string(content), "\"org.springframework.boot:spring-boot-starter-security\"")
}

func TestAddDependencyAlreadyExistsGroovy(t *testing.T) {
	path := copyGradleFixture(t, "sample-build.gradle")
	p := NewParser()

	added, err := p.AddDependency(path, "org.springframework.boot", "spring-boot-starter-web", "", "compile")
	require.NoError(t, err)
	assert.False(t, added, "should not add duplicate dependency")
}

func TestAddDependencyAlreadyExistsKotlin(t *testing.T) {
	path := copyGradleFixture(t, "sample-build.gradle.kts")
	p := NewParser()

	added, err := p.AddDependency(path, "org.springframework.boot", "spring-boot-starter-web", "", "compile")
	require.NoError(t, err)
	assert.False(t, added, "should not add duplicate dependency")
}

func TestGroovyRemoveDependency(t *testing.T) {
	path := copyGradleFixture(t, "sample-build.gradle")
	p := NewParser()

	removed, err := p.RemoveDependency(path, "spring-boot-starter-web")
	require.NoError(t, err)
	assert.True(t, removed)

	has, err := p.HasDependency(path, "spring-boot-starter-web")
	require.NoError(t, err)
	assert.False(t, has)
}

func TestGroovyRemoveShortForm(t *testing.T) {
	path := copyGradleFixture(t, "sample-build.gradle")
	p := NewParser()

	removed, err := p.RemoveDependency(path, "web")
	require.NoError(t, err)
	assert.True(t, removed)
}

func TestKotlinRemoveDependency(t *testing.T) {
	path := copyGradleFixture(t, "sample-build.gradle.kts")
	p := NewParser()

	removed, err := p.RemoveDependency(path, "spring-boot-starter-data-jpa")
	require.NoError(t, err)
	assert.True(t, removed)

	has, err := p.HasDependency(path, "spring-boot-starter-data-jpa")
	require.NoError(t, err)
	assert.False(t, has)
}

func TestRemoveDependencyNotFound(t *testing.T) {
	path := copyGradleFixture(t, "sample-build.gradle")
	p := NewParser()

	removed, err := p.RemoveDependency(path, "nonexistent-lib")
	require.NoError(t, err)
	assert.False(t, removed)
}

func TestScopeMapping(t *testing.T) {
	assert.Equal(t, "testImplementation", mapScopeToConfiguration("test"))
	assert.Equal(t, "compileOnly", mapScopeToConfiguration("provided"))
	assert.Equal(t, "runtimeOnly", mapScopeToConfiguration("runtime"))
	assert.Equal(t, "implementation", mapScopeToConfiguration("compile"))
	assert.Equal(t, "implementation", mapScopeToConfiguration(""))
}

func TestGradleGroovyDslFormat(t *testing.T) {
	line := formatDependencyLine("implementation", "com.example", "lib", "", false)
	assert.Contains(t, line, "'com.example:lib'")
	assert.NotContains(t, line, "\"")
}

func TestGradleKotlinDslFormat(t *testing.T) {
	line := formatDependencyLine("implementation", "com.example", "lib", "", true)
	assert.Contains(t, line, "\"com.example:lib\"")
	assert.NotContains(t, line, "'")
}

func TestIsKotlinDSL(t *testing.T) {
	p := NewParser()
	assert.True(t, p.IsKotlinDSL("build.gradle.kts"))
	assert.True(t, p.IsKotlinDSL("/some/path/build.gradle.kts"))
	assert.False(t, p.IsKotlinDSL("build.gradle"))
	assert.False(t, p.IsKotlinDSL("/some/path/build.gradle"))
}

func TestHasDependency(t *testing.T) {
	path := copyGradleFixture(t, "sample-build.gradle")
	p := NewParser()

	has, err := p.HasDependency(path, "spring-boot-starter-web")
	require.NoError(t, err)
	assert.True(t, has)

	has, err = p.HasDependency(path, "nonexistent")
	require.NoError(t, err)
	assert.False(t, has)
}

func TestAddDependencyWithVersion(t *testing.T) {
	path := copyGradleFixture(t, "sample-build.gradle")
	p := NewParser()

	added, err := p.AddDependency(path, "io.jsonwebtoken", "jjwt-api", "0.12.5", "compile")
	require.NoError(t, err)
	assert.True(t, added)

	content, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Contains(t, string(content), "io.jsonwebtoken:jjwt-api:0.12.5")
}
