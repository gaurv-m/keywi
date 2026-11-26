package app

import (
	"fmt"
	"github.com/charmbracelet/x/ansi"
	"github.com/gaurv-m/keywi/internal/theme"
	"github.com/gaurv-m/keywi/internal/ui"
	"strings"
)

func (m Model) configurationView(w int, p theme.Palette) string {
	line := func(s string) string { return ui.CenterLine(s, w) }
	title := func(s string) string { return line(p.Accent(s)) }
	hint := func(s string) string { return line(p.Muted.Render(ansi.Truncate(s, w, "…"))) }
	var body string
	choice := func(label string, selected, focused, reserve bool) string {
		if selected {
			return p.Primary.Underline(focused).Render("[" + label + "]")
		}
		if reserve {
			label = " " + label + " "
		}
		return p.Muted.Render(label)
	}
	timeChoice := choice("Time", m.config.Kind == "time", m.configRow == 0, true)
	wordsChoice := choice("Words", m.config.Kind == "words", m.configRow == 0, true)
	kindGap := strings.Repeat(" ", min(8, max(1, w-ui.Width(timeChoice)-ui.Width(wordsChoice))))
	kinds := timeChoice + kindGap + wordsChoice
	values := []int{15, 30, 60, 120}
	description := ""
	if m.config.Kind == "words" {
		values = []int{10, 25, 50, 100}
		description = fmt.Sprintf("Finish %d words.", m.config.Amount)
	}
	presets := make([]string, len(values))
	presetWidth := 0
	for i, n := range values {
		label := fmt.Sprint(n)
		if m.config.Kind == "time" {
			label += "s"
		}
		presets[i] = choice(label, n == m.config.Amount, m.configRow == 1, w >= 24)
		presetWidth += ui.Width(presets[i])
	}
	gap := strings.Repeat(" ", min(6, max(1, (w-presetWidth)/3)))
	sectionGap := "\n\n"
	if m.Height >= 18 {
		sectionGap = "\n\n\n"
	}
	if m.Height < 12 {
		sectionGap = "\n"
	}
	body = strings.Join([]string{title("Casual"), line(kinds), line(strings.Join(presets, gap))}, sectionGap) + "\n\n" + hint(description)

	return body
}
