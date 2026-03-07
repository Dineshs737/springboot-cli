package initializr

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDependencyStruct(t *testing.T) {
	dep := Dependency{
		ID:          "web",
		Name:        "Spring Web",
		Description: "Build web applications",
		GroupID:     "org.springframework.boot",
		ArtifactID:  "spring-boot-starter-web",
		Scope:       "compile",
		Starter:     true,
	}
	assert.Equal(t, "web", dep.ID)
	assert.Equal(t, "Spring Web", dep.Name)
	assert.Equal(t, "org.springframework.boot", dep.GroupID)
	assert.True(t, dep.Starter)
}

func TestResolvedDependencyStruct(t *testing.T) {
	rd := ResolvedDependency{
		ID:         "web",
		Name:       "Spring Web",
		GroupID:    "org.springframework.boot",
		ArtifactID: "spring-boot-starter-web",
		Category:   "Web",
		Starter:    true,
	}
	assert.Equal(t, "web", rd.ID)
	assert.Equal(t, "Web", rd.Category)
	assert.Empty(t, rd.Version)
}

func TestDependencyCategoryStruct(t *testing.T) {
	cat := DependencyCategory{
		Name: "Web",
		Values: []Dependency{
			{ID: "web", Name: "Spring Web"},
			{ID: "webflux", Name: "Spring Reactive Web"},
		},
	}
	assert.Equal(t, "Web", cat.Name)
	assert.Len(t, cat.Values, 2)
}

func TestMetadataResponseStruct(t *testing.T) {
	meta := MetadataResponse{
		Dependencies: DependencyGroup{
			Type: "hierarchical-multi-select",
			Values: []DependencyCategory{
				{
					Name: "Web",
					Values: []Dependency{
						{ID: "web", Name: "Spring Web"},
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
	assert.Equal(t, "17", meta.JavaVersion.Default)
	assert.Len(t, meta.Dependencies.Values, 1)
	assert.Len(t, meta.Dependencies.Values[0].Values, 1)
}

func TestValueStruct(t *testing.T) {
	v := Value{
		ID:   "java",
		Name: "Java",
	}
	assert.Equal(t, "java", v.ID)
	assert.Equal(t, "Java", v.Name)
}
