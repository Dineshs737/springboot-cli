package ui

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSuccess(t *testing.T) {
	var buf bytes.Buffer
	p := NewPrinterWithWriter(&buf)
	p.Success("Project created: %s", "my-api")
	out := buf.String()
	assert.Contains(t, out, "✔")
	assert.Contains(t, out, "Project created: my-api")
}

func TestError(t *testing.T) {
	var buf bytes.Buffer
	p := NewPrinterWithWriter(&buf)
	p.Error("Failed to download: %s", "timeout")
	out := buf.String()
	assert.Contains(t, out, "✖")
	assert.Contains(t, out, "Failed to download: timeout")
}

func TestWarn(t *testing.T) {
	var buf bytes.Buffer
	p := NewPrinterWithWriter(&buf)
	p.Warn("Dependency already present: %s", "web")
	out := buf.String()
	assert.Contains(t, out, "⚠")
	assert.Contains(t, out, "Dependency already present: web")
}

func TestInfo(t *testing.T) {
	var buf bytes.Buffer
	p := NewPrinterWithWriter(&buf)
	p.Info("Using Java %d", 17)
	out := buf.String()
	assert.Contains(t, out, "ℹ")
	assert.Contains(t, out, "Using Java 17")
}

func TestStep(t *testing.T) {
	var buf bytes.Buffer
	p := NewPrinterWithWriter(&buf)
	p.Step("Downloading from %s", "start.spring.io")
	out := buf.String()
	assert.Contains(t, out, "▸")
	assert.Contains(t, out, "Downloading from start.spring.io")
}

func TestDim(t *testing.T) {
	var buf bytes.Buffer
	p := NewPrinterWithWriter(&buf)
	p.Dim("some dimmed text")
	out := buf.String()
	assert.Contains(t, out, "some dimmed text")
}

func TestHeader(t *testing.T) {
	var buf bytes.Buffer
	p := NewPrinterWithWriter(&buf)
	p.Header("Test Header")
	out := buf.String()
	assert.Contains(t, out, "┌")
	assert.Contains(t, out, "Test Header")
	assert.Contains(t, out, "└")
}

func TestTableRendersCorrectly(t *testing.T) {
	var buf bytes.Buffer
	p := NewPrinterWithWriter(&buf)
	headers := []string{"ID", "Name", "Description"}
	rows := [][]string{
		{"web", "Spring Web", "Build web apps"},
		{"jpa", "Spring Data JPA", "Persist data"},
	}
	p.Table(headers, rows)
	out := buf.String()
	assert.Contains(t, out, "ID")
	assert.Contains(t, out, "Name")
	assert.Contains(t, out, "Description")
	assert.Contains(t, out, "web")
	assert.Contains(t, out, "Spring Web")
	assert.Contains(t, out, "Spring Data JPA")
	// Check box-drawing characters.
	assert.Contains(t, out, "┌")
	assert.Contains(t, out, "┘")
}

func TestTableEmptyHeaders(t *testing.T) {
	var buf bytes.Buffer
	p := NewPrinterWithWriter(&buf)
	p.Table([]string{}, nil)
	assert.Empty(t, buf.String())
}

func TestTruncate(t *testing.T) {
	assert.Equal(t, "hello", truncate("hello", 10))
	assert.Equal(t, "hel...", truncate("hello world", 6))
	assert.Equal(t, "hel", truncate("hello", 3))
}

func TestBanner(t *testing.T) {
	var buf bytes.Buffer
	p := NewPrinterWithWriter(&buf)
	p.Banner("v1.0.0")
	out := buf.String()
	assert.Contains(t, out, "Spring Boot CLI Tool")
	assert.Contains(t, out, "v1.0.0")
}

func TestTableColumnWidths(t *testing.T) {
	var buf bytes.Buffer
	p := NewPrinterWithWriter(&buf)
	headers := []string{"A", "B"}
	rows := [][]string{
		{"short", "also short"},
		{"a longer value here", "x"},
	}
	p.Table(headers, rows)
	out := buf.String()
	// Verify the table has correct structure.
	assert.Contains(t, out, "A")
	assert.Contains(t, out, "B")
	assert.Contains(t, out, "short")
	assert.Contains(t, out, "a longer value here")
	// Verify all border characters are present.
	assert.Contains(t, out, "┌")
	assert.Contains(t, out, "┘")
	assert.Contains(t, out, "├")
	assert.Contains(t, out, "┤")
}
