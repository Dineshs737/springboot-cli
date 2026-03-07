package project

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDetectMaven(t *testing.T) {
	dir := t.TempDir()
	err := os.WriteFile(filepath.Join(dir, "pom.xml"), []byte("<project/>"), 0644)
	require.NoError(t, err)

	d := NewDetector()
	bt, err := d.Detect(dir)
	require.NoError(t, err)
	assert.Equal(t, BuildTypeMaven, bt)
}

func TestDetectGradleGroovy(t *testing.T) {
	dir := t.TempDir()
	err := os.WriteFile(filepath.Join(dir, "build.gradle"), []byte("plugins {}"), 0644)
	require.NoError(t, err)

	d := NewDetector()
	bt, err := d.Detect(dir)
	require.NoError(t, err)
	assert.Equal(t, BuildTypeGradleGroovy, bt)
}

func TestDetectGradleKotlin(t *testing.T) {
	dir := t.TempDir()
	err := os.WriteFile(filepath.Join(dir, "build.gradle.kts"), []byte("plugins {}"), 0644)
	require.NoError(t, err)

	d := NewDetector()
	bt, err := d.Detect(dir)
	require.NoError(t, err)
	assert.Equal(t, BuildTypeGradleKotlin, bt)
}

func TestDetectUnknown(t *testing.T) {
	dir := t.TempDir()
	d := NewDetector()
	bt, err := d.Detect(dir)
	require.NoError(t, err)
	assert.Equal(t, BuildTypeUnknown, bt)
}

func TestDetectMavenPriority(t *testing.T) {
	// When both pom.xml and build.gradle exist, Maven takes priority.
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "pom.xml"), []byte("<project/>"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "build.gradle"), []byte("plugins {}"), 0644))

	d := NewDetector()
	bt, err := d.Detect(dir)
	require.NoError(t, err)
	assert.Equal(t, BuildTypeMaven, bt)
}

func TestGetBuildFilePathMaven(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "pom.xml"), []byte("<project/>"), 0644))

	d := NewDetector()
	path, bt, err := d.GetBuildFilePath(dir)
	require.NoError(t, err)
	assert.Equal(t, BuildTypeMaven, bt)
	assert.Equal(t, filepath.Join(dir, "pom.xml"), path)
}

func TestGetBuildFilePathGradleKotlin(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "build.gradle.kts"), []byte("plugins {}"), 0644))

	d := NewDetector()
	path, bt, err := d.GetBuildFilePath(dir)
	require.NoError(t, err)
	assert.Equal(t, BuildTypeGradleKotlin, bt)
	assert.Equal(t, filepath.Join(dir, "build.gradle.kts"), path)
}

func TestGetBuildFilePathUnknown(t *testing.T) {
	dir := t.TempDir()
	d := NewDetector()
	_, _, err := d.GetBuildFilePath(dir)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no build file found")
}

func TestDetectInvalidDir(t *testing.T) {
	d := NewDetector()
	_, err := d.Detect("/nonexistent/path/that/does/not/exist")
	assert.Error(t, err)
}

func TestDetectFileNotDir(t *testing.T) {
	dir := t.TempDir()
	filePath := filepath.Join(dir, "not-a-dir")
	require.NoError(t, os.WriteFile(filePath, []byte("hi"), 0644))

	d := NewDetector()
	_, err := d.Detect(filePath)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not a directory")
}
