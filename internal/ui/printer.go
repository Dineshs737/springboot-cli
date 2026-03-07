package ui

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/briandowns/spinner"
	"github.com/fatih/color"
)

// Printer handles all styled terminal output.
type Printer struct {
	out io.Writer
}

// NewPrinter creates a Printer that writes to stdout.
func NewPrinter() *Printer {
	return &Printer{out: os.Stdout}
}

// NewPrinterWithWriter creates a Printer with a custom writer (useful for testing).
func NewPrinterWithWriter(w io.Writer) *Printer {
	return &Printer{out: w}
}

// Success prints a green checkmark message.
func (p *Printer) Success(format string, a ...any) {
	msg := fmt.Sprintf(format, a...)
	green := color.New(color.FgGreen, color.Bold)
	fmt.Fprintf(p.out, "%s %s\n", green.Sprint("✔"), msg)
}

// Error prints a red cross message.
func (p *Printer) Error(format string, a ...any) {
	msg := fmt.Sprintf(format, a...)
	red := color.New(color.FgRed, color.Bold)
	fmt.Fprintf(p.out, "%s %s\n", red.Sprint("✖"), msg)
}

// Warn prints a yellow warning message.
func (p *Printer) Warn(format string, a ...any) {
	msg := fmt.Sprintf(format, a...)
	yellow := color.New(color.FgYellow, color.Bold)
	fmt.Fprintf(p.out, "%s %s\n", yellow.Sprint("⚠"), msg)
}

// Info prints a cyan info message.
func (p *Printer) Info(format string, a ...any) {
	msg := fmt.Sprintf(format, a...)
	cyan := color.New(color.FgCyan, color.Bold)
	fmt.Fprintf(p.out, "%s %s\n", cyan.Sprint("ℹ"), msg)
}

// Step prints a blue step indicator.
func (p *Printer) Step(format string, a ...any) {
	msg := fmt.Sprintf(format, a...)
	blue := color.New(color.FgBlue, color.Bold)
	fmt.Fprintf(p.out, "%s %s\n", blue.Sprint("▸"), msg)
}

// Dim prints dimmed/gray text.
func (p *Printer) Dim(format string, a ...any) {
	msg := fmt.Sprintf(format, a...)
	dim := color.New(color.Faint)
	fmt.Fprintln(p.out, dim.Sprint(msg))
}

// Header prints a boxed title with cyan borders.
func (p *Printer) Header(title string) {
	cyan := color.New(color.FgCyan, color.Bold)
	padding := 2
	width := len(title) + padding*2
	top := "┌" + strings.Repeat("─", width) + "┐"
	mid := "│" + strings.Repeat(" ", padding) + title + strings.Repeat(" ", padding) + "│"
	bot := "└" + strings.Repeat("─", width) + "┘"
	fmt.Fprintln(p.out, cyan.Sprint(top))
	fmt.Fprintln(p.out, cyan.Sprint(mid))
	fmt.Fprintln(p.out, cyan.Sprint(bot))
}

// Table prints a formatted table with unicode box-drawing characters.
func (p *Printer) Table(headers []string, rows [][]string) {
	if len(headers) == 0 {
		return
	}

	// Calculate column widths.
	widths := make([]int, len(headers))
	for i, h := range headers {
		widths[i] = len(h)
	}
	for _, row := range rows {
		for i := 0; i < len(row) && i < len(widths); i++ {
			if len(row[i]) > widths[i] {
				widths[i] = len(row[i])
			}
		}
	}

	// Cap column widths at 60 characters.
	for i := range widths {
		if widths[i] > 60 {
			widths[i] = 60
		}
	}

	cyan := color.New(color.FgCyan)
	bold := color.New(color.Bold)

	// Top border.
	topParts := make([]string, len(widths))
	for i, w := range widths {
		topParts[i] = strings.Repeat("─", w+2)
	}
	fmt.Fprintf(p.out, "%s\n", cyan.Sprintf("┌%s┐", strings.Join(topParts, "┬")))

	// Header row.
	headerCells := make([]string, len(headers))
	for i, h := range headers {
		headerCells[i] = fmt.Sprintf(" %s ", bold.Sprintf("%-*s", widths[i], truncate(h, widths[i])))
	}
	fmt.Fprintf(p.out, "%s%s%s\n", cyan.Sprint("│"), strings.Join(headerCells, cyan.Sprint("│")), cyan.Sprint("│"))

	// Header separator.
	sepParts := make([]string, len(widths))
	for i, w := range widths {
		sepParts[i] = strings.Repeat("─", w+2)
	}
	fmt.Fprintf(p.out, "%s\n", cyan.Sprintf("├%s┤", strings.Join(sepParts, "┼")))

	// Data rows.
	for _, row := range rows {
		cells := make([]string, len(widths))
		for i := 0; i < len(widths); i++ {
			val := ""
			if i < len(row) {
				val = row[i]
			}
			cells[i] = fmt.Sprintf(" %-*s ", widths[i], truncate(val, widths[i]))
		}
		fmt.Fprintf(p.out, "%s%s%s\n", cyan.Sprint("│"), strings.Join(cells, cyan.Sprint("│")), cyan.Sprint("│"))
	}

	// Bottom border.
	botParts := make([]string, len(widths))
	for i, w := range widths {
		botParts[i] = strings.Repeat("─", w+2)
	}
	fmt.Fprintf(p.out, "%s\n", cyan.Sprintf("└%s┘", strings.Join(botParts, "┴")))
}

// StartSpinner creates and starts a terminal spinner with the given message.
func (p *Printer) StartSpinner(msg string) *spinner.Spinner {
	s := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
	s.Suffix = " " + msg
	s.Color("cyan")
	s.Start()
	return s
}

// StopSpinner stops the spinner and prints a completion message.
func (p *Printer) StopSpinner(s *spinner.Spinner, msg string) {
	s.Stop()
	p.Success(msg)
}

// Banner prints the SpringCLI ASCII banner.
func (p *Printer) Banner(version string) {
	cyan := color.New(color.FgCyan, color.Bold)
	dim := color.New(color.Faint)
	fmt.Fprintln(p.out, cyan.Sprint(`
  ____             _              ____ _     ___
 / ___| _ __  _ __(_)_ __   __ _/ ___| |   |_ _|
 \___ \| '_ \| '__| | '_ \ / _`+"`"+` | |   | |    | |
  ___) | |_) | |  | | | | | (_| | |___| |___ | |
 |____/| .__/|_|  |_|_| |_|\__, |\____|_____|___|
       |_|                  |___/`))
	fmt.Fprintln(p.out, dim.Sprintf("  Spring Boot CLI Tool — %s", version))
	fmt.Fprintln(p.out)
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	if max <= 3 {
		return s[:max]
	}
	return s[:max-3] + "..."
}
