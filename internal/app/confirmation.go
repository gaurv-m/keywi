package app

import (
	"github.com/gaurv-m/keywi/internal/theme"
	"github.com/gaurv-m/keywi/internal/ui"

	"github.com/charmbracelet/x/ansi"
)

func confirmationChoices(no, yes string, selectedYes bool, width int, p theme.Palette) string {
	compact := ui.Width(no)+ui.Width(yes)+13 > width
	gap := "     "
	if compact {
		gap = "   "
	}
	choice := func(label string, selected bool) string {
		left, right := "  ", "  "
		if compact {
			left, right = " ", " "
		}
		if selected {
			left, right = "[ ", " ]"
			if compact {
				left, right = "[", "]"
			}
			left, right = p.Primary.Render(left), p.Primary.Render(right)
		}
		return left + p.Correct.Render(label) + right
	}
	return choice(no, !selectedYes) + gap + choice(yes, selectedYes)
}

func (m Model) confirmationView(width int, p theme.Palette) string {
	label, help := "Abandon this run?", "The timer keeps running."
	if m.mode == "normal" && m.config.Kind == "words" {
		help = "Your elapsed time keeps running."
	}
	if m.screen == play && m.tracker().Started.IsZero() {
		help = "The run hasn't started yet."
	}
	no, yes := "Keep", "Abandon"
	switch m.confirm {
	case "quit":
		label, help, no, yes = "Exit KeyWi?", "", "Continue", "Exit"
	case "reset":
		label, help, no, yes = "Reset all statistics?", "A local backup will be preserved.", "Cancel", "Reset"
	case "restart":
		label, no, yes = "Restart this test?", "Cancel", "Restart"
	}
	rows := []string{ui.CenterLine(p.Accent(label), width), ""}
	if help != "" {
		rows = append(rows, ui.CenterLine(p.Muted.Render(ansi.Truncate(help, width, "…")), width))
	}
	rows = append(rows, "", ui.CenterLine(confirmationChoices(no, yes, m.confirmYes, width, p), width))
	return ui.Join(rows...)
}
