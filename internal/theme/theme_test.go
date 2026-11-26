package theme

import (
	"os"
	"testing"

	"github.com/charmbracelet/lipgloss"
)

func TestNativeAndNoColor(t *testing.T) {
	t.Setenv("TERM", "xterm-256color")
	if Normalize("monochrome") != "kiwi" {
		t.Fatal("removed theme did not fall back to Kiwi")
	}
	t.Setenv("NO_COLOR", "")
	if !New("tropical").Plain {
		t.Fatal("NO_COLOR empty not respected")
	}
	os.Unsetenv("NO_COLOR")
	t.Setenv("TERM", "dumb")
	if !New("tropical").Plain {
		t.Fatal("dumb")
	}
}

func TestThemePalettes(t *testing.T) {
	// Palette definitions remain inspectable even when output is not a TTY.
	t.Setenv("TERM", "xterm-256color")
	old, wasSet := os.LookupEnv("NO_COLOR")
	os.Unsetenv("NO_COLOR")
	t.Cleanup(func() {
		if wasSet {
			os.Setenv("NO_COLOR", old)
		}
	})
	kiwi := New("kiwi")
	contrast := New("contrast")
	rainbow := New("rainbow")
	if kiwi.Plain || contrast.Plain || rainbow.Plain {
		t.Fatal("coloured theme disabled")
	}
	if kiwi.Primary.GetForeground() == contrast.Primary.GetForeground() {
		t.Fatal("contrast palette unchanged")
	}
	if len(rainbow.Spectrum) != 6 {
		t.Fatal("rainbow missing spectrum")
	}
	if kiwi.LogoI.GetForeground() != lipgloss.Color("#B68A62") {
		t.Fatal("Kiwi i is not brown")
	}
	if Normalize("tropical") != "kiwi" || Label("rainbow") != "Rainbow" {
		t.Fatal("theme names")
	}
	if Normalize("Ember") != "fall" || Label("ember") != "Fall" || !New("fire").Flame {
		t.Fatal("Fall migration or Fire animation missing")
	}
	seen := map[lipgloss.TerminalColor]string{}
	for _, name := range Names() {
		p := New(name)
		if p.Error.GetUnderline() {
			t.Fatal("error colors must not underline text", name)
		}
		color := p.Primary.GetForeground()
		if previous, exists := seen[color]; exists {
			t.Fatalf("%s reused %s's primary palette", name, previous)
		}
		seen[color] = name
		if len(p.LogoShades) != 24 {
			t.Fatalf("%s is missing title shades", name)
		}
	}
	t.Setenv("NO_COLOR", "")
	for _, name := range Names() {
		p := New(name)
		if !p.Plain || len(p.Spectrum) > 0 || len(p.LogoShades) > 0 {
			t.Fatal("NO_COLOR ignored")
		}
		if p.Error.GetUnderline() {
			t.Fatal("plain error feedback gained an underline", name)
		}
	}
}

func TestPlainFeedbackUsesDistinctWeights(t *testing.T) {
	t.Setenv("NO_COLOR", "")
	p := New("kiwi")
	if p.Correct.GetBold() || !p.Error.GetBold() || !p.Muted.GetFaint() || p.Error.GetUnderline() {
		t.Fatal("plain feedback does not distinguish correct and incorrect input cleanly")
	}
}
