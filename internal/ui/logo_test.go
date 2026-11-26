package ui

import (
	"os"
	"testing"

	"github.com/gaurv-m/keywi/assets"
	"github.com/gaurv-m/keywi/internal/theme"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
)

func logoColors(t testing.TB) {
	t.Helper()
	t.Setenv("TERM", "xterm-256color")
	old, present := os.LookupEnv("NO_COLOR")
	os.Unsetenv("NO_COLOR")
	t.Cleanup(func() {
		if present {
			os.Setenv("NO_COLOR", old)
		}
	})
	renderer := lipgloss.DefaultRenderer()
	previous := renderer.ColorProfile()
	renderer.SetColorProfile(termenv.TrueColor)
	t.Cleanup(func() { renderer.SetColorProfile(previous) })
}

func TestTitleEffectsPreserveArtworkAndStayBounded(t *testing.T) {
	logoColors(t)
	renderer := lipgloss.DefaultRenderer()
	for _, profile := range []termenv.Profile{termenv.TrueColor, termenv.ANSI256, termenv.ANSI} {
		renderer.SetColorProfile(profile)
		for _, name := range theme.Names() {
			p := theme.New(name)
			for frame := 0; frame < 240; frame += 3 {
				got := LogoFrame(assets.Logo(), p, frame)
				if Plain(got) != assets.Logo() || Width(got) != Width(assets.Logo()) {
					t.Fatalf("%s frame %d changed the artwork", name, frame)
				}
				if len(got) > 8192 {
					t.Fatalf("%s frame %d expanded to %d bytes", name, frame, len(got))
				}
			}
		}
	}
}

func TestTitleEffectsChangeOnlyWhenColourIsAllowed(t *testing.T) {
	logoColors(t)
	for _, name := range theme.Names() {
		p := theme.New(name)
		if LogoFrame(assets.Logo(), p, 12) == LogoFrame(assets.Logo(), p, 28) {
			t.Fatalf("%s title colours did not move", name)
		}
	}
	t.Setenv("NO_COLOR", "")
	for _, name := range theme.Names() {
		p := theme.New(name)
		if LogoFrame(assets.Logo(), p, 12) != Logo(assets.Logo(), p) {
			t.Fatalf("%s ignored NO_COLOR", name)
		}
	}
}

func TestTitleShadesFormFastIndependentPatches(t *testing.T) {
	const shades = 24
	increased, decreased, unchanged, fast := false, false, false, false
	vertical := map[int]bool{}
	for frame := 0; frame < 80; frame++ {
		for y := 0; y < 9; y++ {
			for x := 0; x < 42; x++ {
				value := titleShade(x, y, frame, shades)
				next := titleShade(x, y, frame+1, shades)
				if value < 0 || value >= shades || value != titleShade(x, y, frame, shades) {
					t.Fatal("unstable or out-of-palette title shade")
				}
				delta := next - value
				if delta < -6 || delta > 6 {
					t.Fatal("title shades jumped outside the short color transition")
				}
				fast = fast || delta > 1 || delta < -1
				increased = increased || delta > 0
				decreased = decreased || delta < 0
				unchanged = unchanged || delta == 0
				if frame == 12 && x == 12 {
					vertical[value] = true
				}
			}
		}
	}
	if !fast || !increased || !decreased || !unchanged || len(vertical) < 2 {
		t.Fatal("title colours changed as a uniform band rather than independent patches")
	}
}

func BenchmarkTitleEffects(b *testing.B) {
	logoColors(b)
	for _, name := range theme.Names() {
		b.Run(name, func(b *testing.B) {
			p := theme.New(name)
			art := assets.Logo()
			b.ReportAllocs()
			b.ResetTimer()
			for frame := 0; frame < b.N; frame++ {
				LogoFrame(art, p, frame)
			}
		})
	}
}
