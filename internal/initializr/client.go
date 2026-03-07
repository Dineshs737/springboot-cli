package initializr

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

const (
	defaultBaseURL = "https://start.spring.io"
	httpTimeout    = 30 * time.Second
)

// Client is the interface for interacting with the Spring Initializr API.
type Client interface {
	FetchMetadata(ctx context.Context) (*MetadataResponse, error)
	DownloadProject(ctx context.Context, config ProjectDownloadConfig, destDir string) error
	ResolveDependency(ctx context.Context, id string) (*ResolvedDependency, error)
	SearchDependencies(ctx context.Context, query string) ([]ResolvedDependency, error)
	ListAllDependencies(ctx context.Context) ([]DependencyCategory, error)
}

// ProjectDownloadConfig holds parameters for downloading a project ZIP.
type ProjectDownloadConfig struct {
	Type         string
	Language     string
	BootVersion  string
	BaseDir      string
	GroupID      string
	ArtifactID   string
	Name         string
	Description  string
	PackageName  string
	Packaging    string
	JavaVersion  string
	Dependencies []string
}

// HTTPClient is the production implementation of Client.
type HTTPClient struct {
	baseURL    string
	httpClient *http.Client
	metadata   *MetadataResponse
	mu         sync.Mutex
}

// NewClient creates a new HTTPClient with the default base URL.
func NewClient() *HTTPClient {
	return &HTTPClient{
		baseURL: defaultBaseURL,
		httpClient: &http.Client{
			Timeout: httpTimeout,
		},
	}
}

// NewClientWithBaseURL creates an HTTPClient with a custom base URL (for testing).
func NewClientWithBaseURL(baseURL string) *HTTPClient {
	return &HTTPClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: httpTimeout,
		},
	}
}

// FetchMetadata fetches and caches the Spring Initializr metadata.
func (c *HTTPClient) FetchMetadata(ctx context.Context) (*MetadataResponse, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.metadata != nil {
		return c.metadata, nil
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/metadata/client", nil)
	if err != nil {
		return nil, fmt.Errorf("creating metadata request: %w", err)
	}
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetching metadata: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("metadata endpoint returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading metadata response: %w", err)
	}

	var meta MetadataResponse
	if err := json.Unmarshal(body, &meta); err != nil {
		return nil, fmt.Errorf("parsing metadata JSON: %w", err)
	}

	c.metadata = &meta
	return &meta, nil
}

// DownloadProject downloads and extracts a Spring Boot project ZIP.
func (c *HTTPClient) DownloadProject(ctx context.Context, config ProjectDownloadConfig, destDir string) error {
	// Build query parameters.
	params := url.Values{}
	params.Set("type", mapBuildType(config.Type))
	params.Set("language", config.Language)
	params.Set("bootVersion", config.BootVersion)
	params.Set("baseDir", config.BaseDir)
	params.Set("groupId", config.GroupID)
	params.Set("artifactId", config.ArtifactID)
	params.Set("name", config.Name)
	params.Set("description", config.Description)
	params.Set("packageName", config.PackageName)
	params.Set("packaging", config.Packaging)
	params.Set("javaVersion", config.JavaVersion)

	if len(config.Dependencies) > 0 {
		params.Set("dependencies", strings.Join(config.Dependencies, ","))
	}

	reqURL := c.baseURL + "/starter.zip?" + params.Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return fmt.Errorf("creating download request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("downloading project: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("download endpoint returned status %d: %s", resp.StatusCode, string(body))
	}

	// Read the entire ZIP into memory for extraction.
	zipData, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("reading ZIP response: %w", err)
	}

	reader, err := zip.NewReader(bytes.NewReader(zipData), int64(len(zipData)))
	if err != nil {
		return fmt.Errorf("opening ZIP archive: %w", err)
	}

	return extractZip(reader, destDir)
}

// ResolveDependency resolves a dependency ID to its full coordinates.
func (c *HTTPClient) ResolveDependency(ctx context.Context, id string) (*ResolvedDependency, error) {
	meta, err := c.FetchMetadata(ctx)
	if err != nil {
		return nil, fmt.Errorf("fetching metadata for resolution: %w", err)
	}

	lowerID := strings.ToLower(id)
	for _, cat := range meta.Dependencies.Values {
		for _, dep := range cat.Values {
			if strings.ToLower(dep.ID) == lowerID {
				return &ResolvedDependency{
					ID:         dep.ID,
					Name:       dep.Name,
					GroupID:    resolveGroupID(dep),
					ArtifactID: resolveArtifactID(dep),
					Version:    dep.Version,
					Scope:      dep.Scope,
					Category:   cat.Name,
					Starter:    dep.Starter,
				}, nil
			}
		}
	}

	return nil, fmt.Errorf("dependency '%s' not found", id)
}

// SearchDependencies searches for dependencies matching a query string.
func (c *HTTPClient) SearchDependencies(ctx context.Context, query string) ([]ResolvedDependency, error) {
	meta, err := c.FetchMetadata(ctx)
	if err != nil {
		return nil, fmt.Errorf("fetching metadata for search: %w", err)
	}

	lowerQuery := strings.ToLower(query)
	var results []ResolvedDependency

	for _, cat := range meta.Dependencies.Values {
		for _, dep := range cat.Values {
			if strings.Contains(strings.ToLower(dep.ID), lowerQuery) ||
				strings.Contains(strings.ToLower(dep.Name), lowerQuery) ||
				strings.Contains(strings.ToLower(dep.Description), lowerQuery) {
				results = append(results, ResolvedDependency{
					ID:         dep.ID,
					Name:       dep.Name,
					GroupID:    resolveGroupID(dep),
					ArtifactID: resolveArtifactID(dep),
					Version:    dep.Version,
					Scope:      dep.Scope,
					Category:   cat.Name,
					Starter:    dep.Starter,
				})
			}
		}
	}

	return results, nil
}

// ListAllDependencies returns all available dependency categories.
func (c *HTTPClient) ListAllDependencies(ctx context.Context) ([]DependencyCategory, error) {
	meta, err := c.FetchMetadata(ctx)
	if err != nil {
		return nil, fmt.Errorf("fetching metadata for listing: %w", err)
	}
	return meta.Dependencies.Values, nil
}

// knownDependencyCoords maps Spring Initializr dependency IDs to their correct
// Maven coordinates for dependencies that do NOT follow the standard
// "org.springframework.boot:spring-boot-starter-<id>" naming convention.
var knownDependencyCoords = map[string]struct {
	GroupID    string
	ArtifactID string
}{
	// Databases
	"mysql":      {"com.mysql", "mysql-connector-j"},
	"postgresql": {"org.postgresql", "postgresql"},
	"h2":         {"com.h2database", "h2"},
	"mariadb":    {"org.mariadb.jdbc", "mariadb-java-client"},
	"sqlserver":  {"com.microsoft.sqlserver", "mssql-jdbc"},
	"oracle":     {"com.oracle.database.jdbc", "ojdbc11"},

	// Developer tools & utilities
	"lombok":                  {"org.projectlombok", "lombok"},
	"devtools":                {"org.springframework.boot", "spring-boot-devtools"},
	"configuration-processor": {"org.springframework.boot", "spring-boot-configuration-processor"},
	"docker-compose":          {"org.springframework.boot", "spring-boot-docker-compose"},

	// Observability
	"prometheus": {"io.micrometer", "micrometer-registry-prometheus"},

	// Testing
	"testcontainers": {"org.testcontainers", "testcontainers"},

	// Messaging
	"kafka": {"org.springframework.kafka", "spring-kafka"},
	"amqp":  {"org.springframework.boot", "spring-boot-starter-amqp"},

	// Caching & NoSQL
	"data-redis":         {"org.springframework.boot", "spring-boot-starter-data-redis"},
	"data-mongodb":       {"org.springframework.boot", "spring-boot-starter-data-mongodb"},
	"data-elasticsearch": {"org.springframework.boot", "spring-boot-starter-data-elasticsearch"},

	// Security
	"oauth2-client":          {"org.springframework.boot", "spring-boot-starter-oauth2-client"},
	"oauth2-resource-server": {"org.springframework.boot", "spring-boot-starter-oauth2-resource-server"},

	// Misc
	"flyway":     {"org.flywaydb", "flyway-core"},
	"liquibase":  {"org.liquibase", "liquibase-core"},
	"validation": {"org.springframework.boot", "spring-boot-starter-validation"},
	"mail":       {"org.springframework.boot", "spring-boot-starter-mail"},
	"websocket":  {"org.springframework.boot", "spring-boot-starter-websocket"},
	"graphql":    {"org.springframework.boot", "spring-boot-starter-graphql"},
	"batch":      {"org.springframework.boot", "spring-boot-starter-batch"},
	"quartz":     {"org.springframework.boot", "spring-boot-starter-quartz"},
}

func resolveGroupID(dep Dependency) string {
	if dep.GroupID != "" {
		return dep.GroupID
	}
	if coords, ok := knownDependencyCoords[strings.ToLower(dep.ID)]; ok {
		return coords.GroupID
	}
	return "org.springframework.boot"
}

func resolveArtifactID(dep Dependency) string {
	if dep.ArtifactID != "" {
		return dep.ArtifactID
	}
	if coords, ok := knownDependencyCoords[strings.ToLower(dep.ID)]; ok {
		return coords.ArtifactID
	}
	return "spring-boot-starter-" + dep.ID
}

func mapBuildType(t string) string {
	switch t {
	case "maven":
		return "maven-project"
	case "gradle":
		return "gradle-project"
	case "gradle-kotlin":
		return "gradle-project"
	default:
		return "maven-project"
	}
}

func extractZip(reader *zip.Reader, destDir string) error {
	for _, file := range reader.File {
		// ZIP files always use forward slashes. Standardize them to the OS separator.
		cleanName := filepath.FromSlash(file.Name)
		fpath := filepath.Join(destDir, cleanName)

		// Security check: prevent path traversal (ZipSlip vulnerability).
		absDestDir, err := filepath.Abs(destDir)
		if err != nil {
			return fmt.Errorf("resolving absolute path for destination: %w", err)
		}

		absFpath, err := filepath.Abs(fpath)
		if err != nil {
			return fmt.Errorf("resolving absolute path for extracted file: %w", err)
		}

		if !strings.HasPrefix(absFpath, absDestDir+string(os.PathSeparator)) {
			// Allow exact match of destDir itself (if a folder matching the root is in the zip)
			if absFpath != absDestDir {
				return fmt.Errorf("illegal file path in ZIP: %s", file.Name)
			}
		}

		if file.FileInfo().IsDir() {
			if err := os.MkdirAll(fpath, os.ModePerm); err != nil {
				return fmt.Errorf("creating directory %s: %w", fpath, err)
			}
			continue
		}

		if err := os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
			return fmt.Errorf("creating parent directory for %s: %w", fpath, err)
		}

		outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, file.Mode())
		if err != nil {
			return fmt.Errorf("creating file %s: %w", fpath, err)
		}

		rc, err := file.Open()
		if err != nil {
			outFile.Close()
			return fmt.Errorf("opening ZIP entry %s: %w", file.Name, err)
		}

		_, err = io.Copy(outFile, rc)
		rc.Close()
		outFile.Close()
		if err != nil {
			return fmt.Errorf("extracting %s: %w", file.Name, err)
		}
	}

	return nil
}
