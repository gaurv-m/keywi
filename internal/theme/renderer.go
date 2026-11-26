package theme

import (
	"os"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
)

// Keep cursor and weight attributes when the palette disables colors.
func ConfigureRenderer() {
	_, noColor := os.LookupEnv("NO_COLOR")
	if noColor || os.Getenv("TERM") == "dumb" {
		lipgloss.SetColorProfile(termenv.ANSI)
	}
}
