package themes

import (
	"fmt"
	"math"
	"os"
	"strings"
)

// RenderBanner returns an ASCII banner string based on style.
// Supported styles: "default", "block", "minimal", "gradient", "gemini", "gradient_solid", "solid", "gradient_fill", "gradient_outline_fill", "kawaii"
func RenderBanner(style string) string {
	switch strings.ToLower(style) {
	case "block":
		return blockBanner()
	case "minimal":
		return minimalBanner()
	case "gradient":
		return gradientBanner()
	default:
		return gradientBanner()
	}
}

func defaultBanner() string {
	// Rounded + whale mark
	return ansi256("36", "\n      🐳  Docsh\n  Docker Command Mapping Shell\n\n")
}

func minimalBanner() string {
	return "Docsh (Docker-Only)\n\n"
}

func blockBanner() string {
	ascii := []string{
		`  ____    ___    ____   ____    _   _ `,
		` |  _ \  / _ \  / ___| / ___|  | | | |`,
		` | | | || | | | | |    \___ \  | |_| |`,
		` | |_| || |_| | | |__   ___) | |  _  |`,
		` |____/  \___/   \____||____/  |_| |_|`,
		"            🐳 Docsh                    ",
	}
	return ansi256("36", fmt.Sprintln(strings.Join(ascii, "\n"))) + "\n"
}

// ----- Gradient (truecolor) banner -----

type rgb struct{ R, G, B int }

func gradientBanner() string {
	// Big ASCII for "DOCSH" (5 rows)
	lines := []string{
		`  ____    ___    ____   ____    _   _ `,
		` |  _ \  / _ \  / ___| / ___|  | | | |`,
		` | | | || | | | | |    \___ \  | |_| |`,
		` | |_| || |_| | | |__   ___) | |  _  |`,
		` |____/  \___/   \____||____/  |_| |_|`,
	}

	// Colors: pink -> cyan
	start := rgb{R: 255, G: 102, B: 170} // pink
	end := rgb{R: 0, G: 255, B: 255}     // cyan

	// Fallback to non-truecolor terminals
	if !supportsTrueColor() {
		// Use 256-color approximate: magenta to cyan (5;13 -> 6;14 not exact)
		return ansi256Gradient(lines)
	}

	width := maxWidth(lines)
	out := &strings.Builder{}
	for _, line := range lines {
		for i := 0; i < width; i++ {
			var ch byte = ' '
			if i < len(line) {
				ch = line[i]
			}
			t := float64(i) / math.Max(1, float64(width-1))
			c := lerpColor(start, end, t)
			if ch == ' ' {
				out.WriteByte(' ')
				continue
			}
			out.WriteString(truecolor(c.R, c.G, c.B))
			out.WriteByte(ch)
		}
		out.WriteString("\033[0m\n")
	}
	out.WriteString("\n")
	return out.String()
}

func maxWidth(lines []string) int {
	w := 0
	for _, l := range lines {
		if len(l) > w {
			w = len(l)
		}
	}
	return w
}

func lerpColor(a, b rgb, t float64) rgb {
	cl := func(x int) int {
		if x < 0 {
			return 0
		}
		if x > 255 {
			return 255
		}
		return x
	}
	r := int(float64(a.R) + (float64(b.R)-float64(a.R))*t)
	g := int(float64(a.G) + (float64(b.G)-float64(a.G))*t)
	b2 := int(float64(a.B) + (float64(b.B)-float64(a.B))*t)
	return rgb{R: cl(r), G: cl(g), B: cl(b2)}
}

func truecolor(r, g, b int) string   { return fmt.Sprintf("\033[38;2;%d;%d;%dm", r, g, b) }
func bgTruecolor(r, g, b int) string { return fmt.Sprintf("\033[48;2;%d;%d;%dm", r, g, b) }

func supportsTrueColor() bool {
	if v := os.Getenv("COLORTERM"); strings.Contains(strings.ToLower(v), "truecolor") || strings.Contains(strings.ToLower(v), "24bit") {
		return true
	}
	// Assume modern terminals support it; Windows is enabled elsewhere
	return true
}

func ansi256(code, text string) string { return fmt.Sprintf("\033[%sm%s\033[0m", code, text) }
func bgAnsi256(code string) string     { return "\033[" + code + "m" }

func ansi256Gradient(lines []string) string {
	// Use a small palette from magenta -> purple -> blue -> cyan
	palette := []string{"38;5;201", "38;5;135", "38;5;69", "38;5;51"}
	width := maxWidth(lines)
	out := &strings.Builder{}
	for _, line := range lines {
		for i := 0; i < width; i++ {
			var ch byte = ' '
			if i < len(line) {
				ch = line[i]
			}
			idx := int(float64(i) / math.Max(1, float64(width)) * float64(len(palette)))
			if idx >= len(palette) {
				idx = len(palette) - 1
			}
			if ch == ' ' {
				out.WriteByte(' ')
				continue
			}
			out.WriteString("\033[" + palette[idx] + "m")
			out.WriteByte(ch)
		}
		out.WriteString("\033[0m\n")
	}
	out.WriteString("\n")
	return out.String()
}

// helper to pick an index into the 256-color palette for background gradient
func gradientPaletteIndex(x, width int) string {
	palette := []string{"48;5;201", "48;5;135", "48;5;69", "48;5;51"}
	idx := int(float64(x) / math.Max(1, float64(width)) * float64(len(palette)))
	if idx >= len(palette) {
		idx = len(palette) - 1
	}
	return palette[idx]
}
