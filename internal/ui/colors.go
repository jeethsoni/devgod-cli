package ui

import (
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func ColorsEnabled() bool {
	if os.Getenv("NO_COLOR") != "" {
		return false
	}
	fi, err := os.Stdout.Stat()
	if err != nil {
		return true
	}
	// disable when piped
	if (fi.Mode() & os.ModeCharDevice) == 0 {
		return false
	}
	term := strings.ToLower(os.Getenv("TERM"))
	return term != "" && term != "dumb"
}

func render(s lipgloss.Style, text string) string {
	if !ColorsEnabled() {
		return text
	}
	return s.Render(text)
}

var (
	TitleStyle  = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("10")).Border(lipgloss.NormalBorder()).BorderForeground(lipgloss.Color("8")).Padding(0, 1).MarginBottom(1)
	Divider     = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "7", Dark: "245"})
	ValueStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("15"))
	PromptStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("14"))

	GreenStyle        = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("10"))
	YellowStyle       = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("11"))
	RedStyle          = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("9"))
	CyanStyle         = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("14"))
	SectionTitleStyle = lipgloss.NewStyle().
				Bold(true).
				Underline(true).
				Foreground(lipgloss.AdaptiveColor{
			Light: "0",
			Dark:  "15",
		})

	BranchLabelStyle = lipgloss.NewStyle().
				Bold(true).
				Underline(true).
				Foreground(lipgloss.AdaptiveColor{
			Light: "25",
			Dark:  "75",
		})

	IntentLabelStyle = lipgloss.NewStyle().
				Bold(true).
				Underline(true).
				Foreground(lipgloss.AdaptiveColor{
			Light: "90",
			Dark:  "177",
		})

	CommitLabelStyle = lipgloss.NewStyle().
				Bold(true).
				Underline(true).
				Foreground(lipgloss.AdaptiveColor{
			Light: "60",
			Dark:  "109",
		})
)

func Green(s string) string  { return render(GreenStyle, s) }
func Yellow(s string) string { return render(YellowStyle, s) }
func Red(s string) string    { return render(RedStyle, s) }
func Cyan(s string) string   { return render(CyanStyle, s) }

func Bold(s string) string { return render(lipgloss.NewStyle().Bold(true), s) }
func Dim(s string) string  { return render(Divider, s) }

// "M file.py" -> colors only the status letter
func ColorizeStatusLine(line string) string {
	trim := strings.TrimSpace(line)
	if trim == "" {
		return line
	}

	parts := strings.Fields(trim)
	if len(parts) < 2 {
		return trim
	}

	code := parts[0] // A, M, D, R, etc.

	switch code {
	case "A":
		return GreenStyle.Render(trim)
	case "M":
		return YellowStyle.Render(trim)
	case "D":
		return RedStyle.Render(trim)
	case "R":
		return CyanStyle.Render(trim)
	default:
		return ValueStyle.Render(trim)
	}

}
