package ui

import (
	"fmt"
	"io"
	"os"

	"github.com/fatih/color"
)

// BannerPrinter renders the SpringCLI ASCII banner.
type BannerPrinter struct {
	out io.Writer
}

// NewBannerPrinter creates a BannerPrinter writing to stdout.
func NewBannerPrinter() *BannerPrinter {
	return &BannerPrinter{out: os.Stdout}
}

// NewBannerPrinterWithWriter creates a BannerPrinter with a custom writer.
func NewBannerPrinterWithWriter(w io.Writer) *BannerPrinter {
	return &BannerPrinter{out: w}
}

// Print renders the full ASCII banner with tagline.
func (b *BannerPrinter) Print(version string) {
	green := color.New(color.FgGreen, color.Bold)
	cyan := color.New(color.FgCyan)
	dim := color.New(color.Faint)

	banner := `
 ██████  ██████  ██████  ██ ███    ██  ██████   ██████ ██      ██
 ██      ██   ██ ██   ██ ██ ████   ██ ██       ██      ██      ██
 ╚█████  ██████  ██████  ██ ██ ██  ██ ██   ███ ██      ██      ██
      ██ ██      ██   ██ ██ ██  ██ ██ ██    ██ ██      ██      ██
 ██████  ██      ██   ██ ██ ██   ████  ██████   ██████ ███████ ██`

	fmt.Fprintln(b.out, green.Sprint(banner))
	fmt.Fprintln(b.out)
	fmt.Fprintf(b.out, " %s\n", cyan.Sprint("The Spring Boot CLI — like create-next-app, but for Java/Kotlin"))
	fmt.Fprintf(b.out, " %s\n", dim.Sprint("─────────────────────────────────────────────────────────────────"))
	fmt.Fprintln(b.out)
}

// PrintCompact renders a smaller banner for sub-commands.
func (b *BannerPrinter) PrintCompact(version string) {
	cyan := color.New(color.FgCyan, color.Bold)
	dim := color.New(color.Faint)
	fmt.Fprintf(b.out, " %s %s\n", cyan.Sprint("🌱 SpringCLI"), dim.Sprintf("(%s)", version))
	fmt.Fprintln(b.out)
}
