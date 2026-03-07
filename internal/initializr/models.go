package initializr

// MetadataResponse represents the top-level response from the Spring Initializr metadata endpoint.
type MetadataResponse struct {
	Dependencies  DependencyGroup   `json:"dependencies"`
	Type          ValueGroup        `json:"type"`
	Packaging     ValueGroup        `json:"packaging"`
	JavaVersion   ValueGroup        `json:"javaVersion"`
	Language      ValueGroup        `json:"language"`
	BootVersion   ValueGroup        `json:"bootVersion"`
}

// DependencyGroup holds dependency categories.
type DependencyGroup struct {
	Type   string              `json:"type"`
	Values []DependencyCategory `json:"values"`
}

// DependencyCategory is a named group of dependencies (e.g., "Web", "Security").
type DependencyCategory struct {
	Name   string       `json:"name"`
	Values []Dependency `json:"values"`
}

// Dependency represents a single Spring Boot dependency.
type Dependency struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	GroupID     string `json:"groupId"`
	ArtifactID  string `json:"artifactId"`
	Version     string `json:"version"`
	Scope       string `json:"scope"`
	Starter     bool   `json:"starter"`
}

// ValueGroup holds a group of selectable values (e.g., Java versions, languages).
type ValueGroup struct {
	Type    string  `json:"type"`
	Default string  `json:"default"`
	Values  []Value `json:"values"`
}

// Value represents a single selectable option.
type Value struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// ResolvedDependency holds the resolved coordinates of a dependency.
type ResolvedDependency struct {
	ID         string
	Name       string
	GroupID    string
	ArtifactID string
	Version    string
	Scope      string
	Category   string
	Starter    bool
}
