package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/springcli/springcli/internal/gradle"
	"github.com/springcli/springcli/internal/maven"
	"github.com/springcli/springcli/internal/project"
)

var auditCmd = &cobra.Command{
	Use:   "audit",
	Short: "Audit project dependencies for known vulnerabilities",
	Long: `Scans your local build tool file (pom.xml or build.gradle)
for direct explicit dependencies and queries the OSV (Open Source Vulnerability)
database API to find known vulnerabilities. Optionally applies patch updates.`,
	RunE: runAudit,
}

var (
	auditUpdate bool
	auditDir    string
)

func init() {
	auditCmd.Flags().BoolVarP(&auditUpdate, "update", "u", false, "Automatically update vulnerable dependencies to fixed versions")
	auditCmd.Flags().StringVar(&auditDir, "dir", ".", "Target project directory")
	rootCmd.AddCommand(auditCmd)
}

// OSV Query Data structures
type osvQuery struct {
	Package struct {
		Name      string `json:"name"`
		Ecosystem string `json:"ecosystem"`
	} `json:"package"`
	Version string `json:"version"`
}

type osvBatchQuery struct {
	Queries []osvQuery `json:"queries"`
}

type osvBatchResponse struct {
	Results []osvResult `json:"results"`
}

type osvResult struct {
	Vulns []osvVuln `json:"vulns,omitempty"`
}

type osvVuln struct {
	ID       string `json:"id"`
	Details  string `json:"details"`
	Affected []struct {
		Ranges []struct {
			Type   string `json:"type"`
			Events []struct {
				Introduced string `json:"introduced,omitempty"`
				Fixed      string `json:"fixed,omitempty"`
			} `json:"events"`
		} `json:"ranges"`
	} `json:"affected"`
}

// Internal structure to link parsed deps with OSV results
type auditDependency struct {
	GroupID    string
	ArtifactID string
	Version    string
	OSVPkgName string
}

func runAudit(cmd *cobra.Command, args []string) error {
	detector := project.NewDetector()
	buildPath, buildType, err := detector.GetBuildFilePath(auditDir)
	if err != nil {
		return fmt.Errorf("detecting project: %w", err)
	}

	printer.Info("Detecting explicit dependencies from %s", buildPath)

	deps, err := extractDependencies(buildPath, buildType)
	if err != nil {
		return fmt.Errorf("parsing dependencies: %w", err)
	}

	if len(deps) == 0 {
		printer.Warn("No explicit dependencies with versions found in build file to audit.")
		return nil
	}

	printer.Step("Found %d explicit versioned dependencies. Querying OSV database...", len(deps))

	// Construct batch query
	batch := osvBatchQuery{}
	for _, dep := range deps {
		q := osvQuery{}
		q.Package.Ecosystem = "Maven"
		q.Package.Name = dep.OSVPkgName
		q.Version = dep.Version
		batch.Queries = append(batch.Queries, q)
	}

	payload, err := json.Marshal(batch)
	if err != nil {
		return fmt.Errorf("marshaling OSV query: %w", err)
	}

	client := &http.Client{Timeout: 15 * time.Second}
	req, err := http.NewRequest(http.MethodPost, "https://api.osv.dev/v1/querybatch", bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("creating OSV request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("querying OSV API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("OSV API returned status %d: %s", resp.StatusCode, string(body))
	}

	var batchResp osvBatchResponse
	if err := json.NewDecoder(resp.Body).Decode(&batchResp); err != nil {
		return fmt.Errorf("decoding OSV response: %w", err)
	}

	foundVulns := 0
	for i, result := range batchResp.Results {
		if len(result.Vulns) > 0 {
			foundVulns++
			dep := deps[i]
			printer.Error("❌ Vulnerability found in %s@%s", dep.OSVPkgName, dep.Version)

			// Find highest fixed version among vulns
			var bestFix string
			for _, vuln := range result.Vulns {
				printer.Dim("  - CVE/ID: %s", vuln.ID)

				// Try to extract fixed version
				for _, affected := range vuln.Affected {
					for _, r := range affected.Ranges {
						for _, event := range r.Events {
							if event.Fixed != "" {
								// Basic logic: grab the first available fixed string we see
								if bestFix == "" || event.Fixed > bestFix { // extremely naive string comparison for semver
									bestFix = event.Fixed
								}
							}
						}
					}
				}
			}

			if bestFix != "" {
				printer.Success("  + A fix is available: version %s", bestFix)
				if auditUpdate {
					printer.Step("  > Updating build file to version %s...", bestFix)
					if err := updateDependencyVersion(buildPath, buildType, dep, bestFix); err != nil {
						printer.Error("    Failed to update: %v", err)
					} else {
						printer.Success("    Update applied successfully.")
					}
				} else {
					printer.Dim("  Run `springcli audit --update` to automatically fix.")
				}
			} else {
				printer.Warn("  - No fixed version reported by OSV for this vulnerability.")
			}
			fmt.Println()
		}
	}

	if foundVulns == 0 {
		printer.Success("✅ No vulnerabilities found in your explicitly versioned dependencies.")
	} else {
		printer.Warn("Found %d vulnerable dependencies.", foundVulns)
	}

	return nil
}

func extractDependencies(buildPath string, buildType project.BuildType) ([]auditDependency, error) {
	var deps []auditDependency

	switch buildType {
	case project.BuildTypeMaven:
		mp := maven.NewParser()
		mavenDeps, err := mp.ListDependencies(buildPath)
		if err != nil {
			return nil, err
		}
		for _, md := range mavenDeps {
			// Skip managed deps without versions or variables like ${version}
			if md.Version != "" && !strings.HasPrefix(md.Version, "$") {
				deps = append(deps, auditDependency{
					GroupID:    md.GroupID,
					ArtifactID: md.ArtifactID,
					Version:    md.Version,
					OSVPkgName: md.GroupID + ":" + md.ArtifactID,
				})
			}
		}

	case project.BuildTypeGradleGroovy, project.BuildTypeGradleKotlin:
		gp := gradle.NewParser()
		gradleDeps, err := gp.ListDependencies(buildPath)
		if err != nil {
			return nil, err
		}
		for _, gd := range gradleDeps {
			if gd.Version != "" && !strings.HasPrefix(gd.Version, "$") && !strings.Contains(gd.Version, "ext.") {
				deps = append(deps, auditDependency{
					GroupID:    gd.Group,
					ArtifactID: gd.Artifact,
					Version:    gd.Version,
					OSVPkgName: gd.Group + ":" + gd.Artifact,
				})
			}
		}
	}

	return deps, nil
}

func updateDependencyVersion(buildPath string, buildType project.BuildType, dep auditDependency, newVersion string) error {
	// For Maven, we need an edit parser method. Let's do a basic find-replace approach as a fallback if the parser doesn't natively support modify
	contentBytes, err := os.ReadFile(buildPath)
	if err != nil {
		return err
	}
	content := string(contentBytes)

	if buildType == project.BuildTypeMaven {
		// Replacing `<version>old</version>` inside the specific dependency block is complex via string replace.
		// Use etree for safe manipulation.
		mp := maven.NewParser()
		doc, err := mp.ReadPom(buildPath)
		if err != nil {
			return err
		}
		project := doc.Root()
		if project != nil {
			depsElem := project.FindElement("dependencies")
			if depsElem != nil {
				for _, existing := range depsElem.SelectElements("dependency") {
					var g, a string
					if gEl := existing.FindElement("groupId"); gEl != nil {
						g = gEl.Text()
					}
					if aEl := existing.FindElement("artifactId"); aEl != nil {
						a = aEl.Text()
					}

					if g == dep.GroupID && a == dep.ArtifactID {
						if vEl := existing.FindElement("version"); vEl != nil {
							vEl.SetText(newVersion)
							return mp.WritePom(doc, buildPath)
						}
					}
				}
			}
		}
		return fmt.Errorf("dependency not found or no version tag present")
	} else {
		// Gradle string replace
		// The `gradleDeps` parser gives us gd.RawLine. E.g. `implementation 'org.foo:bar:1.0'`
		// We have to recreate the raw line and replace it
		oldCoords := fmt.Sprintf("%s:%s:%s", dep.GroupID, dep.ArtifactID, dep.Version)
		newCoords := fmt.Sprintf("%s:%s:%s", dep.GroupID, dep.ArtifactID, newVersion)

		if !strings.Contains(content, oldCoords) {
			return fmt.Errorf("exact coordinates %s not found in gradle file", oldCoords)
		}

		newContent := strings.Replace(content, oldCoords, newCoords, 1) // only replace the first occurrence to be safe
		return os.WriteFile(buildPath, []byte(newContent), 0o644)
	}
}
