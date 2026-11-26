package app

import (
	"fmt"
	"strings"

	"github.com/gaurv-m/keywi/internal/theme"
	"github.com/gaurv-m/keywi/internal/ui"

	"github.com/charmbracelet/x/ansi"
)

// IDs retain the original preference actions; visual order groups related items.
var settingsOrder = [...]int{0, 1, 2, 3, 7, 6, 4, 5, 8}

func nextSetting(current, delta int) int {
	for i, id := range settingsOrder {
		if id == current {
			return settingsOrder[(i+delta+len(settingsOrder))%len(settingsOrder)]
		}
	}
	return settingsOrder[0]
}
func (m Model) settingsPanel(width int, p theme.Palette) string {
	prefs := m.Store.Preferences
	onoff := func(v bool) string {
		if v {
			return "on"
		}
		return "off"
	}
	type item struct {
		id           int
		label, value string
	}
	groups := []struct {
		name  string
		items []item
	}{
		{"Appearance", []item{{0, "Theme", theme.Label(prefs.Theme)}, {1, "Cursor", prefs.Cursor}, {2, "Animations", onoff(prefs.Animations)}}},
		{"Feedback", []item{{3, "Live WPM", onoff(prefs.LiveWPM)}, {7, "Sound", onoff(prefs.Sound)}}},
		{"Word tests", []item{{6, "Capitals", onoff(prefs.Capitalization)}, {4, "Punctuation", onoff(prefs.Punctuation)}, {5, "Numbers", onoff(prefs.Numbers)}}},
		{"Data", []item{{8, "Reset statistics", ""}}},
	}
	type row struct {
		id   int
		text string
	}
	var rows []row
	selected := 0
	panelWidth := min(33, width)
	labelWidth := min(11, max(7, panelWidth-14))
	columnGap := 8
	if panelWidth < 31 {
		columnGap = 5
	}
	if panelWidth < 28 {
		columnGap = 2
	}
	for _, group := range groups {
		rows = append(rows, row{-1, p.Secondary.Render("  " + group.name)})
		for _, option := range group.items {
			text := option.label
			if option.value != "" {
				label := ansi.Truncate(option.label, labelWidth, "…")
				value := ansi.Truncate(option.value, max(1, panelWidth-labelWidth-columnGap-2), "…")
				text = fmt.Sprintf("%-*s%s%s", labelWidth, label, strings.Repeat(" ", columnGap), value)
			}
			if option.id == m.settingRow {
				selected = len(rows)
				text = p.Primary.Render("› " + text)
			} else {
				text = p.Muted.Render("  " + text)
			}
			rows = append(rows, row{option.id, text})
		}
	}
	visible := min(len(rows), max(1, m.Height-9))
	start := max(0, min(selected-visible/2, len(rows)-visible))
	out := make([]string, 0, visible)
	for _, entry := range rows[start : start+visible] {
		text := ansi.Truncate(entry.text, panelWidth, "…")
		out = append(out, text+strings.Repeat(" ", max(0, panelWidth-ui.Width(text))))
	}
	return ui.CenterBlock(strings.Join(out, "\n"), width)
}
func (m Model) settingHelp() string {
	switch m.settingRow {
	case 0:
		return "Foreground colours; your terminal background stays native."
	case 1:
		return "Choose how the typing cursor looks."
	case 2:
		return "Title colour effects and Blitz sentence shattering."
	case 3:
		return "Show speed while typing in Casual."
	case 7:
		if !m.Audio.Available() {
			return "Needs afplay, paplay or aplay on this system."
		}
		return "A short 8-bit shatter when you finish a Blitz sentence."
	case 6, 4, 5:
		return "Word tests only; Time keeps the paragraph's original text."
	case 8:
		return "Requires confirmation and preserves a local backup."
	}
	return ""
}
