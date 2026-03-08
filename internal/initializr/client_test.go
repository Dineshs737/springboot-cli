package initializr

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func sampleMetadata() MetadataResponse {
	return MetadataResponse{
		Dependencies: DependencyGroup{
			Type: "hierarchical-multi-select",
			Values: []DependencyCategory{
				{
					Name: "Web",
					Values: []Dependency{
						{
							ID:          "web",
							Name:        "Spring Web",
							Description: "Build web, including RESTful, applications using Spring MVC",
							GroupID:     "org.springframework.boot",
							ArtifactID:  "spring-boot-starter-web",
							Starter:     true,
						},
						{
							ID:          "webflux",
							Name:        "Spring Reactive Web",
							Description: "Build reactive web applications with Spring WebFlux",
							GroupID:     "org.springframework.boot",
							ArtifactID:  "spring-boot-starter-webflux",
							Starter:     true,
						},
					},
				},
				{
					Name: "SQL",
					Values: []Dependency{
						{
							ID:          "jpa",
							Name:        "Spring Data JPA",
							Description: "Persist data in SQL stores with Java Persistence API",
							GroupID:     "org.springframework.boot",
							ArtifactID:  "spring-boot-starter-data-jpa",
							Starter:     true,
						},
						{
							ID:          "postgresql",
							Name:        "PostgreSQL Driver",
							Description: "A JDBC and R2DBC driver",
							GroupID:     "org.postgresql",
							ArtifactID:  "postgresql",
							Scope:       "runtime",
						},
					},
				},
				{
					Name: "Security",
					Values: []Dependency{
						{
							ID:          "security",
							Name:        "Spring Security",
							Description: "Authentication and access-control framework",
							GroupID:     "org.springframework.boot",
							ArtifactID:  "spring-boot-starter-security",
							Starter:     true,
						},
					},
				},
			},
		},
		JavaVersion: ValueGroup{
			Default: "17",
			Values: []Value{
				{ID: "17", Name: "17"},
				{ID: "21", Name: "21"},
			},
		},
	}
}

func newMockServer(t *testing.T, callCount *atomic.Int32) *httptest.Server {
	t.Helper()
	meta := sampleMetadata()
	metaJSON, err := json.Marshal(meta)
	require.NoError(t, err)

	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/metadata/client":
			if callCount != nil {
				callCount.Add(1)
			}
			w.Header().Set("Content-Type", "application/json")
			w.Write(metaJSON)
		case "/starter.zip":
			w.Header().Set("Content-Type", "application/zip")
			writeTestZip(t, w)
		default:
			http.NotFound(w, r)
		}
	}))
}

func writeTestZip(t *testing.T, w http.ResponseWriter) {
	t.Helper()
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)

	f, err := zw.Create("test-project/pom.xml")
	require.NoError(t, err)
	_, err = f.Write([]byte("<project/>"))
	require.NoError(t, err)

	f2, err := zw.Create("test-project/src/main/java/App.java")
	require.NoError(t, err)
	_, err = f2.Write([]byte("public class App {}"))
	require.NoError(t, err)

	require.NoError(t, zw.Close())
	w.Write(buf.Bytes())
}

func TestFetchMetadata(t *testing.T) {
	server := newMockServer(t, nil)
	defer server.Close()

	client := NewClientWithBaseURL(server.URL)
	meta, err := client.FetchMetadata(context.Background())
	require.NoError(t, err)
	assert.NotNil(t, meta)
	assert.NotEmpty(t, meta.Dependencies.Values)
	assert.Equal(t, "Web", meta.Dependencies.Values[0].Name)
}

func TestMetadataCaching(t *testing.T) {
	var callCount atomic.Int32
	server := newMockServer(t, &callCount)
	defer server.Close()

	client := NewClientWithBaseURL(server.URL)
	ctx := context.Background()

	_, err := client.FetchMetadata(ctx)
	require.NoError(t, err)
	_, err = client.FetchMetadata(ctx)
	require.NoError(t, err)
	_, err = client.FetchMetadata(ctx)
	require.NoError(t, err)

	assert.Equal(t, int32(1), callCount.Load(), "metadata should only be fetched once")
}

func TestDownloadProject(t *testing.T) {
	server := newMockServer(t, nil)
	defer server.Close()

	client := NewClientWithBaseURL(server.URL)
	destDir := t.TempDir()

	config := ProjectDownloadConfig{
		Type:        "maven",
		Language:    "java",
		BootVersion: "3.2.3",
		BaseDir:     "test-project",
		GroupID:     "com.example",
		ArtifactID:  "test",
		Name:        "test",
		Description: "Test project",
		PackageName: "com.example.test",
		Packaging:   "jar",
		JavaVersion: "17",
	}

	err := client.DownloadProject(context.Background(), config, destDir)
	require.NoError(t, err)

	// Verify extracted files exist.
	pomPath := filepath.Join(destDir, "test-project", "pom.xml")
	_, err = os.Stat(pomPath)
	assert.NoError(t, err)

	javaPath := filepath.Join(destDir, "test-project", "src", "main", "java", "App.java")
	_, err = os.Stat(javaPath)
	assert.NoError(t, err)
}

func TestResolveDependencyExact(t *testing.T) {
	server := newMockServer(t, nil)
	defer server.Close()

	client := NewClientWithBaseURL(server.URL)
	resolved, err := client.ResolveDependency(context.Background(), "web")
	require.NoError(t, err)
	assert.Equal(t, "web", resolved.ID)
	assert.Equal(t, "Spring Web", resolved.Name)
	assert.Equal(t, "org.springframework.boot", resolved.GroupID)
	assert.Equal(t, "spring-boot-starter-web", resolved.ArtifactID)
}

func TestResolveDependencyCaseInsensitive(t *testing.T) {
	server := newMockServer(t, nil)
	defer server.Close()

	client := NewClientWithBaseURL(server.URL)
	resolved, err := client.ResolveDependency(context.Background(), "WEB")
	require.NoError(t, err)
	assert.Equal(t, "web", resolved.ID)
}

func TestResolveDependencyNotFound(t *testing.T) {
	server := newMockServer(t, nil)
	defer server.Close()

	client := NewClientWithBaseURL(server.URL)
	_, err := client.ResolveDependency(context.Background(), "nonexistent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestSearchDependencies(t *testing.T) {
	server := newMockServer(t, nil)
	defer server.Close()

	client := NewClientWithBaseURL(server.URL)
	results, err := client.SearchDependencies(context.Background(), "web")
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(results), 2) // "web" and "webflux"
}

func TestSearchDependenciesByDescription(t *testing.T) {
	server := newMockServer(t, nil)
	defer server.Close()

	client := NewClientWithBaseURL(server.URL)
	results, err := client.SearchDependencies(context.Background(), "reactive")
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(results), 1)
	assert.Equal(t, "webflux", results[0].ID)
}

func TestSearchDependenciesNoResults(t *testing.T) {
	server := newMockServer(t, nil)
	defer server.Close()

	client := NewClientWithBaseURL(server.URL)
	results, err := client.SearchDependencies(context.Background(), "zzzzNotExist")
	require.NoError(t, err)
	assert.Empty(t, results)
}

func TestListAllDependencies(t *testing.T) {
	server := newMockServer(t, nil)
	defer server.Close()

	client := NewClientWithBaseURL(server.URL)
	categories, err := client.ListAllDependencies(context.Background())
	require.NoError(t, err)
	assert.Len(t, categories, 3)
	assert.Equal(t, "Web", categories[0].Name)
}

func TestMapBuildType(t *testing.T) {
	assert.Equal(t, "maven-project", mapBuildType("maven"))
	assert.Equal(t, "gradle-project", mapBuildType("gradle"))
	assert.Equal(t, "gradle-project", mapBuildType("gradle-kotlin"))
	assert.Equal(t, "maven-project", mapBuildType("unknown"))
}

func TestFetchMetadataServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	client := NewClientWithBaseURL(server.URL)
	_, err := client.FetchMetadata(context.Background())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "status 500")
}
