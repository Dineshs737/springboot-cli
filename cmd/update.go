package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

const (
	updateRepo    = "Dineshs737/springboot-cli"
	updateTimeout = 30 * time.Second
)

// githubRelease holds the JSON response from the GitHub releases API.
type githubRelease struct {
	TagName string        `json:"tag_name"`
	Assets  []githubAsset `json:"assets"`
}

type githubAsset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
	URL                string `json:"url"`
}

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update SpringCLI to the latest version",
	Long:  `Checks GitHub for the latest release and replaces the current binary if a newer version is available.`,
	RunE:  runUpdate,
}

func init() {
	rootCmd.AddCommand(updateCmd)
}

func runUpdate(cmd *cobra.Command, args []string) error {
	printer.Info("Current version: %s", version)
	printer.Info("Checking for updates...")

	// Build HTTP client and headers.
	client := &http.Client{Timeout: updateTimeout}
	headers := map[string]string{
		"Accept": "application/vnd.github.v3+json",
	}

	token := os.Getenv("GITHUB_TOKEN")
	if token != "" {
		headers["Authorization"] = "token " + token
	}

	// Fetch the latest release from GitHub.
	apiURL := fmt.Sprintf("https://api.github.com/repos/%s/releases/latest", updateRepo)
	req, err := http.NewRequest(http.MethodGet, apiURL, nil)
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("fetching latest release: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("GitHub API returned status %d: %s", resp.StatusCode, string(body))
	}

	var release githubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return fmt.Errorf("parsing release JSON: %w", err)
	}

	latestVersion := strings.TrimPrefix(release.TagName, "v")
	currentVersion := strings.TrimPrefix(version, "v")

	if currentVersion == latestVersion {
		printer.Success("You're already on the latest version (%s). No update needed!", version)
		return nil
	}

	printer.Info("New version available: %s → %s", version, release.TagName)

	// Determine the correct asset name for this platform.
	osName := runtime.GOOS     // "windows", "linux", "darwin"
	archName := runtime.GOARCH // "amd64", "arm64"
	assetName := fmt.Sprintf("springcli-%s-%s", osName, archName)
	if osName == "windows" {
		assetName += ".exe"
	}

	// Find the matching asset in the release.
	var downloadURL string
	var assetAPIURL string
	for _, asset := range release.Assets {
		if asset.Name == assetName {
			downloadURL = asset.BrowserDownloadURL
			assetAPIURL = asset.URL
			break
		}
	}

	if downloadURL == "" {
		return fmt.Errorf("no binary found for %s/%s (looking for %s)", osName, archName, assetName)
	}

	// Download the new binary.
	// For private repos, use the API URL with the Accept: application/octet-stream header.
	actualURL := downloadURL
	downloadHeaders := map[string]string{}
	if token != "" {
		actualURL = assetAPIURL
		downloadHeaders["Accept"] = "application/octet-stream"
		downloadHeaders["Authorization"] = "token " + token
	}

	printer.Info("Downloading %s...", assetName)

	dlReq, err := http.NewRequest(http.MethodGet, actualURL, nil)
	if err != nil {
		return fmt.Errorf("creating download request: %w", err)
	}
	for k, v := range downloadHeaders {
		dlReq.Header.Set(k, v)
	}

	dlResp, err := client.Do(dlReq)
	if err != nil {
		return fmt.Errorf("downloading binary: %w", err)
	}
	defer dlResp.Body.Close()

	if dlResp.StatusCode != http.StatusOK {
		return fmt.Errorf("download returned status %d", dlResp.StatusCode)
	}

	// Write to a temp file first.
	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("finding current executable path: %w", err)
	}
	execPath, err = filepath.EvalSymlinks(execPath)
	if err != nil {
		return fmt.Errorf("resolving symlinks: %w", err)
	}

	tmpFile, err := os.CreateTemp(filepath.Dir(execPath), "springcli-update-*.tmp")
	if err != nil {
		return fmt.Errorf("creating temp file: %w", err)
	}
	tmpPath := tmpFile.Name()

	if _, err := io.Copy(tmpFile, dlResp.Body); err != nil {
		tmpFile.Close()
		os.Remove(tmpPath)
		return fmt.Errorf("writing download to temp file: %w", err)
	}
	tmpFile.Close()

	// Make the temp file executable (no-op on Windows, but useful for Linux/macOS).
	if err := os.Chmod(tmpPath, 0755); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("setting executable permissions: %w", err)
	}

	// Replace the current binary.
	// On Windows, we can't overwrite a running executable directly,
	// so we rename the old one first, then move the new one into place.
	oldPath := execPath + ".old"

	// Clean up any leftover .old file from a previous update.
	os.Remove(oldPath)

	if err := os.Rename(execPath, oldPath); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("renaming current binary: %w", err)
	}

	if err := os.Rename(tmpPath, execPath); err != nil {
		// Try to restore the old binary if the move fails.
		_ = os.Rename(oldPath, execPath)
		os.Remove(tmpPath)
		return fmt.Errorf("moving new binary into place: %w", err)
	}

	// Clean up the old binary (best effort).
	os.Remove(oldPath)

	fmt.Println()
	printer.Success("Successfully updated SpringCLI to %s! 🚀", release.TagName)
	printer.Dim("  Restart your terminal to use the new version.")
	fmt.Println()

	return nil
}
