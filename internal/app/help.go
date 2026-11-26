package app

import (
	"fmt"
	"strings"

	"github.com/gaurv-m/keywi/internal/theme"
	"github.com/gaurv-m/keywi/internal/ui"

	"github.com/charmbracelet/x/ansi"
)

var helpSections = [][2]string{
	{"Casual", "pick a timer or a word count. type to start"},
	{"Blitz", "endless survival. correct typing adds time, up to 30s"},
	{"Levels", "level up every 5 sentences. rewards shrink from 5–10"},
	{"Mistakes", "two mistakes? backspace and fix them. Blitz needs every character right"},
	{"Scoring", "WPM: 5 correct chars = 1 word, spaces included\naccuracy counts fixed mistakes too"},
	{"Controls", "arrows choose · enter start · esc back\ntab,enter restart · ctrl+c quit"},
}

func helpRows(width int, p theme.Palette) []string {
	rows := []string{}
	for i, section := range helpSections {
		if i > 0 {
			rows = append(rows, "")
		}
		if width < 44 {
			rows = append(rows, p.Secondary.Render(section[0]))
			rows = append(rows, strings.Split(ansi.Wordwrap(p.Muted.Render(section[1]), width, ""), "\n")...)
			continue
		}
		lines := strings.Split(ansi.Wordwrap(section[1], width-10, ""), "\n")
		for j, line := range lines {
			label := ""
			if j == 0 {
				label = section[0]
			}
			rows = append(rows, p.Secondary.Render(fmt.Sprintf("%-8s  ", label))+p.Muted.Render(line))
		}
	}
	return rows
}

func (m Model) helpView(width int, p theme.Palette) string {
	panelWidth := min(64, width)
	rows := helpRows(panelWidth, p)
	visible := max(1, m.Height-10)
	start := min(m.helpOffset, max(0, len(rows)-visible))
	heading := "Help"
	if len(rows) > visible {
		heading += " · scroll for more"
	}
	return ui.Join(ui.CenterLine(p.Accent(heading), width), "", ui.CenterBlock(strings.Join(rows[start:min(len(rows), start+visible)], "\n"), width))
}

func (m Model) helpMaxOffset() int {
	width := min(64, ui.ContentWidth(m.Width))
	return max(0, len(helpRows(width, theme.New(m.Store.Preferences.Theme)))-max(1, m.Height-10))
}
