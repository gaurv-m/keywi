package theme

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
)

func TestNoColorKeepsCursorAttributes(t *testing.T) {
	t.Setenv("NO_COLOR", "")
	t.Setenv("TERM", "xterm-256color")
	renderer := lipgloss.DefaultRenderer()
	previous := renderer.ColorProfile()
	t.Cleanup(func() { renderer.SetColorProfile(previous) })
	renderer.SetColorProfile(termenv.Ascii)
	ConfigureRenderer()
	p := New("fire")
	cursor := p.Primary.Reverse(true).Render("x")
	if !strings.Contains(cursor, ";7m") && !strings.Contains(cursor, "[7m") {
		t.Fatal("NO_COLOR hid the reverse cursor", cursor)
	}
	if !p.Plain || len(p.LogoShades) != 0 || p.Primary.GetForeground() != (lipgloss.NoColor{}) {
		t.Fatal("NO_COLOR reenabled palette colors")
	}
}
