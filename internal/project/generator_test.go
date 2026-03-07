package project

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateEnvFileBasic(t *testing.T) {
	dir := t.TempDir()
	g := NewGenerator()

	err := g.GenerateEnvFile(dir, []string{})
	require.NoError(t, err)

	content, err := os.ReadFile(filepath.Join(dir, ".env.example"))
	require.NoError(t, err)

	s := string(content)
	assert.Contains(t, s, "SERVER_PORT=8080")
	assert.Contains(t, s, "SPRING_PROFILES_ACTIVE=dev")
	assert.NotContains(t, s, "DB_HOST") // No DB deps.
}

func TestGenerateEnvFileWithJPA(t *testing.T) {
	dir := t.TempDir()
	g := NewGenerator()

	err := g.GenerateEnvFile(dir, []string{"jpa", "web"})
	require.NoError(t, err)

	content, err := os.ReadFile(filepath.Join(dir, ".env.example"))
	require.NoError(t, err)

	s := string(content)
	assert.Contains(t, s, "DB_HOST=localhost")
	assert.Contains(t, s, "DB_PORT=5432")
	assert.Contains(t, s, "SPRING_DATASOURCE_URL")
}

func TestGenerateEnvFileWithRedisAndKafka(t *testing.T) {
	dir := t.TempDir()
	g := NewGenerator()

	err := g.GenerateEnvFile(dir, []string{"data-redis", "kafka"})
	require.NoError(t, err)

	content, err := os.ReadFile(filepath.Join(dir, ".env.example"))
	require.NoError(t, err)

	s := string(content)
	assert.Contains(t, s, "REDIS_HOST=localhost")
	assert.Contains(t, s, "REDIS_PORT=6379")
	assert.Contains(t, s, "KAFKA_BOOTSTRAP_SERVERS=localhost:9092")
}

func TestGenerateEnvFileWithSecurity(t *testing.T) {
	dir := t.TempDir()
	g := NewGenerator()

	err := g.GenerateEnvFile(dir, []string{"security"})
	require.NoError(t, err)

	content, err := os.ReadFile(filepath.Join(dir, ".env.example"))
	require.NoError(t, err)

	s := string(content)
	assert.Contains(t, s, "JWT_SECRET")
	assert.Contains(t, s, "JWT_EXPIRATION")
}

func TestGenerateDockerfileMaven(t *testing.T) {
	dir := t.TempDir()
	g := NewGenerator()

	err := g.generateDockerfile(dir, "maven", "21")
	require.NoError(t, err)

	content, err := os.ReadFile(filepath.Join(dir, "Dockerfile"))
	require.NoError(t, err)

	s := string(content)
	assert.Contains(t, s, "FROM eclipse-temurin:21-jdk-alpine AS builder")
	assert.Contains(t, s, "COPY .mvn/")
	assert.Contains(t, s, "mvnw")
	assert.Contains(t, s, "target/*.jar")
	assert.Contains(t, s, "EXPOSE 8080")
}

func TestGenerateDockerfileGradle(t *testing.T) {
	dir := t.TempDir()
	g := NewGenerator()

	err := g.generateDockerfile(dir, "gradle", "17")
	require.NoError(t, err)

	content, err := os.ReadFile(filepath.Join(dir, "Dockerfile"))
	require.NoError(t, err)

	s := string(content)
	assert.Contains(t, s, "FROM eclipse-temurin:17-jdk-alpine AS builder")
	assert.Contains(t, s, "gradlew")
	assert.Contains(t, s, "build/libs/*.jar")
}

func TestGenerateDockerComposeWithPostgres(t *testing.T) {
	dir := t.TempDir()
	g := NewGenerator()

	err := g.generateDockerCompose(dir, []string{"jpa"})
	require.NoError(t, err)

	content, err := os.ReadFile(filepath.Join(dir, "docker-compose.yml"))
	require.NoError(t, err)

	s := string(content)
	assert.Contains(t, s, "postgres:")
	assert.Contains(t, s, "postgres:16-alpine")
	assert.Contains(t, s, "POSTGRES_DB: mydb")
	assert.Contains(t, s, "postgres_data")
}

func TestGenerateDockerComposeWithKafka(t *testing.T) {
	dir := t.TempDir()
	g := NewGenerator()

	err := g.generateDockerCompose(dir, []string{"kafka"})
	require.NoError(t, err)

	content, err := os.ReadFile(filepath.Join(dir, "docker-compose.yml"))
	require.NoError(t, err)

	s := string(content)
	assert.Contains(t, s, "zookeeper:")
	assert.Contains(t, s, "kafka:")
	assert.Contains(t, s, "KAFKA_BROKER_ID")
}

func TestGenerateDockerComposeFull(t *testing.T) {
	dir := t.TempDir()
	g := NewGenerator()

	err := g.generateDockerCompose(dir, []string{"jpa", "data-redis", "kafka", "data-mongodb"})
	require.NoError(t, err)

	content, err := os.ReadFile(filepath.Join(dir, "docker-compose.yml"))
	require.NoError(t, err)

	s := string(content)
	assert.Contains(t, s, "postgres:")
	assert.Contains(t, s, "redis:")
	assert.Contains(t, s, "mongodb:")
	assert.Contains(t, s, "kafka:")
	assert.Contains(t, s, "zookeeper:")
}

func TestGenerateDockerIgnore(t *testing.T) {
	dir := t.TempDir()
	g := NewGenerator()

	err := g.generateDockerIgnore(dir)
	require.NoError(t, err)

	content, err := os.ReadFile(filepath.Join(dir, ".dockerignore"))
	require.NoError(t, err)

	s := string(content)
	assert.Contains(t, s, ".git")
	assert.Contains(t, s, "target/")
	assert.Contains(t, s, "build/")
}

func TestGenerateDockerFilesAll(t *testing.T) {
	dir := t.TempDir()
	g := NewGenerator()

	err := g.GenerateDockerFiles(dir, "maven", "21", []string{"jpa"})
	require.NoError(t, err)

	// All 3 files should exist.
	_, err = os.Stat(filepath.Join(dir, "Dockerfile"))
	assert.NoError(t, err)
	_, err = os.Stat(filepath.Join(dir, "docker-compose.yml"))
	assert.NoError(t, err)
	_, err = os.Stat(filepath.Join(dir, ".dockerignore"))
	assert.NoError(t, err)
}

func TestInitGitInvalidDir(t *testing.T) {
	g := NewGenerator()
	err := g.InitGit("/nonexistent/dir/that/does/not/exist")
	assert.Error(t, err)
}

func TestGenerateDockerComposeMinimal(t *testing.T) {
	dir := t.TempDir()
	g := NewGenerator()

	// No deps -> only app service.
	err := g.generateDockerCompose(dir, []string{})
	require.NoError(t, err)

	content, err := os.ReadFile(filepath.Join(dir, "docker-compose.yml"))
	require.NoError(t, err)

	s := string(content)
	assert.Contains(t, s, "app:")
	assert.NotContains(t, s, "postgres:")
	assert.NotContains(t, s, "redis:")
	lines := strings.Split(s, "\n")
	assert.Greater(t, len(lines), 5, "compose file should have content")
}
