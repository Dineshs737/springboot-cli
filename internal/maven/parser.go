package maven

import (
	"fmt"
	"os"
	"strings"

	"github.com/beevik/etree"
)

// MavenDependency represents a Maven dependency.
type MavenDependency struct {
	GroupID    string
	ArtifactID string
	Version    string
	Scope      string
}

// Parser handles reading and writing Maven pom.xml files.
type Parser struct{}

// NewParser creates a new Maven Parser.
func NewParser() *Parser {
	return &Parser{}
}

// ReadPom reads and parses a pom.xml file.
func (p *Parser) ReadPom(path string) (*etree.Document, error) {
	doc := etree.NewDocument()
	if err := doc.ReadFromFile(path); err != nil {
		return nil, fmt.Errorf("reading pom.xml at %s: %w", path, err)
	}
	return doc, nil
}

// WritePom writes the document back to the pom.xml file.
func (p *Parser) WritePom(doc *etree.Document, path string) error {
	doc.Indent(4)
	if err := doc.WriteToFile(path); err != nil {
		return fmt.Errorf("writing pom.xml to %s: %w", path, err)
	}
	return nil
}

// ListDependencies returns all dependencies from the pom.xml.
func (p *Parser) ListDependencies(path string) ([]MavenDependency, error) {
	doc, err := p.ReadPom(path)
	if err != nil {
		return nil, err
	}

	project := doc.Root()
	if project == nil {
		return nil, fmt.Errorf("no root element in pom.xml")
	}

	deps := project.FindElement("dependencies")
	if deps == nil {
		return nil, nil
	}

	var result []MavenDependency
	for _, dep := range deps.SelectElements("dependency") {
		md := MavenDependency{}
		if g := dep.FindElement("groupId"); g != nil {
			md.GroupID = g.Text()
		}
		if a := dep.FindElement("artifactId"); a != nil {
			md.ArtifactID = a.Text()
		}
		if v := dep.FindElement("version"); v != nil {
			md.Version = v.Text()
		}
		if s := dep.FindElement("scope"); s != nil {
			md.Scope = s.Text()
		}
		result = append(result, md)
	}

	return result, nil
}

// AddDependency adds a dependency to the pom.xml. Returns false if it already exists.
func (p *Parser) AddDependency(path string, dep MavenDependency) (bool, error) {
	doc, err := p.ReadPom(path)
	if err != nil {
		return false, err
	}

	project := doc.Root()
	if project == nil {
		return false, fmt.Errorf("no root element in pom.xml")
	}

	// Find or create <dependencies> element.
	deps := project.FindElement("dependencies")
	if deps == nil {
		deps = project.CreateElement("dependencies")
	}

	// Check if dependency already exists.
	for _, existing := range deps.SelectElements("dependency") {
		existingGroup := ""
		existingArtifact := ""
		if g := existing.FindElement("groupId"); g != nil {
			existingGroup = g.Text()
		}
		if a := existing.FindElement("artifactId"); a != nil {
			existingArtifact = a.Text()
		}
		if existingGroup == dep.GroupID && existingArtifact == dep.ArtifactID {
			return false, nil
		}
	}

	// Add the dependency element.
	depElem := deps.CreateElement("dependency")
	depElem.CreateElement("groupId").SetText(dep.GroupID)
	depElem.CreateElement("artifactId").SetText(dep.ArtifactID)
	if dep.Version != "" {
		depElem.CreateElement("version").SetText(dep.Version)
	}
	if dep.Scope != "" {
		depElem.CreateElement("scope").SetText(dep.Scope)
	}

	return true, p.WritePom(doc, path)
}

// RemoveDependency removes a dependency from the pom.xml by artifactId.
// It tries matching "spring-boot-starter-<id>" first, then exact artifactId.
// Returns true if the dependency was found and removed.
func (p *Parser) RemoveDependency(path string, artifactIDOrShort string) (bool, error) {
	doc, err := p.ReadPom(path)
	if err != nil {
		return false, err
	}

	project := doc.Root()
	if project == nil {
		return false, fmt.Errorf("no root element in pom.xml")
	}

	deps := project.FindElement("dependencies")
	if deps == nil {
		return false, nil
	}

	// Try multiple match patterns.
	candidates := []string{
		"spring-boot-starter-" + artifactIDOrShort,
		artifactIDOrShort,
	}

	for _, depElem := range deps.SelectElements("dependency") {
		a := depElem.FindElement("artifactId")
		if a == nil {
			continue
		}
		artifactText := a.Text()
		for _, candidate := range candidates {
			if strings.EqualFold(artifactText, candidate) {
				deps.RemoveChild(depElem)
				return true, p.WritePom(doc, path)
			}
		}
	}

	return false, nil
}

// HasDependency checks if a dependency exists in the pom.xml.
func (p *Parser) HasDependency(path string, artifactID string) (bool, error) {
	allDeps, err := p.ListDependencies(path)
	if err != nil {
		return false, err
	}

	for _, dep := range allDeps {
		if strings.EqualFold(dep.ArtifactID, artifactID) {
			return true, nil
		}
	}
	return false, nil
}

// ReadPomFromString reads a pom.xml from a raw string and returns the document.
func (p *Parser) ReadPomFromString(content string) (*etree.Document, error) {
	doc := etree.NewDocument()
	if err := doc.ReadFromString(content); err != nil {
		return nil, fmt.Errorf("parsing pom.xml string: %w", err)
	}
	return doc, nil
}

// WritePomToString writes the document to a string.
func (p *Parser) WritePomToString(doc *etree.Document) (string, error) {
	doc.Indent(4)
	s, err := doc.WriteToString()
	if err != nil {
		return "", fmt.Errorf("writing pom.xml to string: %w", err)
	}
	return s, nil
}

// AddDependencyToFile reads the pom.xml from the file, adds the dependency, and writes back.
func (p *Parser) AddDependencyToFile(path string, groupID, artifactID, version, scope string) (bool, error) {
	dep := MavenDependency{
		GroupID:    groupID,
		ArtifactID: artifactID,
		Version:    version,
		Scope:      scope,
	}
	return p.AddDependency(path, dep)
}

// DetectBasePackage reads a pom.xml and returns the base package name (groupId.artifactId with hyphens removed).
func (p *Parser) DetectBasePackage(path string) (string, error) {
	doc, err := p.ReadPom(path)
	if err != nil {
		return "", err
	}

	project := doc.Root()
	if project == nil {
		return "", fmt.Errorf("no root element in pom.xml")
	}

	groupID := ""
	artifactID := ""
	if g := project.FindElement("groupId"); g != nil {
		groupID = g.Text()
	}
	if a := project.FindElement("artifactId"); a != nil {
		artifactID = a.Text()
	}

	if groupID == "" {
		return "", fmt.Errorf("groupId not found in pom.xml")
	}

	// Remove hyphens from artifactId for package name.
	cleanArtifact := strings.ReplaceAll(artifactID, "-", "")
	return groupID + "." + cleanArtifact, nil
}

// GetPomContent reads the raw content of a pom.xml file.
func (p *Parser) GetPomContent(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("reading pom.xml: %w", err)
	}
	return string(data), nil
}
