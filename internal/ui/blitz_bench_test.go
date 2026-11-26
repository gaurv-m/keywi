package ui

import (
	"os"
	"testing"
	"time"

	"github.com/gaurv-m/keywi/internal/theme"
	"github.com/gaurv-m/keywi/internal/typing"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
)

func BenchmarkBlitzStageLongSentence(b *testing.B) {
	b.Setenv("TERM", "xterm-256color")
	oldNoColor, wasSet := os.LookupEnv("NO_COLOR")
	os.Unsetenv("NO_COLOR")
	b.Cleanup(func() {
		if wasSet {
			os.Setenv("NO_COLOR", oldNoColor)
		}
	})
	renderer := lipgloss.DefaultRenderer()
	old := renderer.ColorProfile()
	defer renderer.SetColorProfile(old)
	for _, profile := range []struct {
		name  string
		value termenv.Profile
	}{{"TrueColor", termenv.TrueColor}, {"ANSI256", termenv.ANSI256}, {"ANSI16", termenv.ANSI}} {
		b.Run(profile.name, func(b *testing.B) {
			renderer.SetColorProfile(profile.value)
			tracker := typing.New("The afternoon light falls across the pages of an open book beside the narrow kitchen window.")
			p := theme.New("contrast")
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				BlitzStage(&tracker, "A quiet river winds between the hills before it reaches the coast.", 80*time.Millisecond, 76, "block", p, true)
			}
		})
	}
}
