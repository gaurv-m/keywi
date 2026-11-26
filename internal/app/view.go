package app

import (
	"strings"

	"github.com/gaurv-m/keywi/assets"
	"github.com/gaurv-m/keywi/internal/theme"
	"github.com/gaurv-m/keywi/internal/ui"

	"github.com/charmbracelet/x/ansi"
)

func (m Model) View() string {
	if m.Width < 24 || m.Height < 10 {
		return ui.Place(ui.CenterBlock("Keywi\nResize to at least 24 × 10\nctrl+c exit", m.Width), m.Width, m.Height)
	}
	w := ui.ContentWidth(m.Width)
	p := theme.New(m.Store.Preferences.Theme)
	line := func(s string) string { return ui.CenterLine(s, w) }
	title := func(s string) string { return line(p.Accent(s)) }
	hint := func(s string) string {
		if w < 45 {
			compact := map[string]string{
				"← → select · enter confirm":                      "←→ · enter select",
				"↑ ↓ row · ← → change · enter begin · esc back":   "↑↓ · ←→ · enter · esc",
				"esc leave · tab then enter restart":              "esc · tab,enter retry",
				"enter replay · esc menu":                         "enter replay · esc",
				"← → mode · ↑ ↓ bests · enter history · esc menu": "←→ · ↑↓ · enter · esc",
				"↑ ↓ select · enter details · esc statistics":     "↑↓ · enter view · esc",
				"↑ ↓ select · ← → / enter change · esc menu":      "↑↓ · ←→ · enter · esc",
				"← → select · enter confirm · esc return":         "←→ · enter · esc",
				"← → select · enter confirm · esc cancel":         "←→ · enter · esc",
				"↑ ↓ scroll · esc back":                           "↑↓ scroll · esc back",
			}
			if short, ok := compact[s]; ok {
				s = short
			}
		}
		return line(p.Muted.Render(ansi.Truncate(s, w, "…")))
	}
	var body, footer string
	if m.confirm != "" {
		body = m.confirmationView(w, p)
		footer = hint("← → select · enter confirm · esc cancel")
	} else {
		switch m.screen {
		case menu:
			options := []string{"Casual", "Blitz", "Statistics", "Settings"}
			for i := range options {
				if i == m.selection {
					options[i] = p.Accent("› " + options[i])
				} else {
					options[i] = p.Muted.Render("  " + options[i])
				}
			}
			// Reserve marker space so labels stay still.
			nav := strings.Join(options, "    ")
			horizontal := ui.Width(nav) <= w
			legend := "← → select        enter confirm        ? help"
			if !horizontal {
				nav = ui.Join(options[0], options[1], "", options[2], options[3])
				legend = "↑ ↓ select        enter confirm        ? help"
				if w < 45 {
					legend = "↑↓ select · ↵ · ?"
				}
			}
			// Compensate for an odd outer margin when the content width is capped.
			centerWidth := w + (m.Width-w)%2
			nav = ui.CenterBlock(nav, centerWidth)
			footer = ui.CenterLine(p.Muted.Render(ansi.Truncate(legend, w, "…")), centerWidth)
			logo := assets.Logo()
			paintTitle := func(art string) string {
				if m.Store.Preferences.Animations {
					return ui.LogoFrame(art, p, int(m.now.UnixMilli()/100))
				}
				return ui.Logo(art, p)
			}
			if ui.Width(logo) <= w && m.Height >= 21 {
				painted := ui.CenterBlock(paintTitle(logo), centerWidth)
				body = ui.MenuBody(painted, nav, m.Height-3)
			} else {
				compact := p.Accent("Keywi")
				if m.Store.Preferences.Animations {
					compact = paintTitle("Keywi")
				}
				body = ui.MenuBody(ui.CenterLine(compact, centerWidth), nav, m.Height-3)
			}
			if m.warning != "" {
				body += "\n\n" + hint("Storage recovery notice: see startup output")
			}
		case configuration:
			body = m.configurationView(w, p)
			footer = hint("↑ ↓ row · ← → change · enter begin · esc back")
		case play:
			body = m.playView(w, p)
			footer = hint("esc leave · tab then enter restart")
		case results:
			body = m.resultsView(w, p)
			if m.saveError != nil {
				body += "\n" + hint("Score not saved: "+m.saveError.Error())
				footer = hint("enter retry saving · ctrl+c exit")
			} else {
				footer = hint("enter replay · esc menu")
			}
		case detail:
			body = m.detailView(w, p)
			footer = hint("enter / esc history")
		case stats:
			body = m.statisticsView(w, p)
			footer = hint("← → mode · ↑ ↓ bests · enter history · esc menu")
		case history:
			body = m.historyView(w, p)
			footer = hint("↑ ↓ select · enter details · esc statistics")
		case helpScreen:
			body = m.helpView(w, p)
			footer = ui.Join(line(p.Muted.Render("Made by G")), "", hint("↑ ↓ scroll · esc back"))
		case settings:
			body = ui.Join(title("Settings"), "", m.settingsPanel(w, p), "", hint(ansi.Truncate(m.message, w, "…")))
			if m.Height >= 20 {
				body += "\n" + hint(m.settingHelp())
			}
			footer = hint("↑ ↓ select · ← → / enter change · esc menu")
		}
	}
	// Drop blank rows when space is tight.
	lines := strings.Split(body, "\n")
	footerRows := strings.Count(footer, "\n") + 1
	if len(lines) > m.Height-2-footerRows {
		compact := make([]string, 0, len(lines))
		for _, l := range lines {
			if strings.TrimSpace(ansi.Strip(l)) != "" {
				compact = append(compact, l)
			}
		}
		body = strings.Join(compact, "\n")
	}
	// Keep one fixed inner width for centering.
	lines = strings.Split(body, "\n")
	for i, l := range lines {
		l = ansi.Truncate(l, w, "")
		lines[i] = l + strings.Repeat(" ", max(0, w-ui.Width(l)))
	}
	body = strings.Join(lines, "\n")
	footerLines := strings.Split(footer, "\n")
	for i, l := range footerLines {
		l = ansi.Truncate(l, w, "")
		footerLines[i] = l + strings.Repeat(" ", max(0, w-ui.Width(l)))
	}
	footer = strings.Join(footerLines, "\n")
	return ui.Page(ui.CenterBlock(body, m.Width), ui.CenterBlock(footer, m.Width), m.Width, m.Height)
}
