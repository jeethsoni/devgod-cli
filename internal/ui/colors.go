package ui

import (
	"strings"
)

const (
	colorReset  = "\033[0m"
	colorCyan   = "\033[36m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorRed    = "\033[31m"
)

func Cyan(s string) string {
	return colorCyan + s + colorReset
}

func Green(s string) string {
	return colorGreen + s + colorReset
}

func Yellow(s string) string {
	return colorYellow + s + colorReset
}

func Red(s string) string {
	return colorRed + s + colorReset
}

func ColorizeStatusLine(line string) string {
	trim := strings.TrimSpace(line)
	if len(trim) < 2 {
		return line
	}

	code := trim[:1] // A, M, D, R, etc.

	switch code {
	case "A":
		return Green(trim)
	case "M":
		return Yellow(trim)
	case "D":
		return Red(trim)
	case "R":
		return Cyan(trim)
	default:
		return trim
	}
}
