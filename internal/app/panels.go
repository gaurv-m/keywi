package app

import (
	"strings"
	"unicode"

	"github.com/gaurv-m/keywi/internal/theme"
	"github.com/gaurv-m/keywi/internal/ui"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
)

// Title-case display labels while keeping numeric units intact.
func titleLabel(label string) string {
	words := strings.Fields(label)
	for i, word := range words {
		runes := []rune(word)
		if unicode.IsLetter(runes[0]) {
			runes[0] = unicode.ToUpper(runes[0])
		}
		words[i] = string(runes)
	}
	return strings.Join(words, " ")
}

func sectionHeading(label string, width int, p theme.Palette) string {
	return p.Secondary.Render(label) + p.Muted.Faint(true).Render("  "+strings.Repeat("─", max(0, width-ui.Width(label)-2)))
}

// comparisonTable shares the available width between two records.
func comparisonTable(previous, best string, width int, p theme.Palette) string {
	column := width / 2
	return table.New().Headers("Previous Run", "Best Before").Row(previous, best).
		BorderTop(false).BorderBottom(false).BorderLeft(false).BorderRight(false).
		BorderColumn(false).BorderStyle(p.Muted.Faint(true)).Width(column * 2).Wrap(false).
		StyleFunc(func(row, col int) lipgloss.Style {
			style := p.Correct
			if row == table.HeaderRow {
				style = p.Muted
			}
			return style.Width(column).Align(lipgloss.Center)
		}).String()
}

func overviewTable(left, right [][]string, width int, p theme.Palette) string {
	column := (width - 4) / 2
	rows := make([][]string, len(left))
	for i := range left {
		rows[i] = []string{statPair(left[i][0], left[i][1], column, p, false) + "    ", statPair(right[i][0], right[i][1], column, p, false)}
	}
	return table.New().Headers(sectionHeading("Overview", column, p)+"    ", sectionHeading("Activity", column, p)).Rows(rows...).Width(width).BorderTop(false).BorderBottom(false).
		BorderLeft(false).BorderRight(false).BorderHeader(false).BorderColumn(false).
		StyleFunc(func(row, col int) lipgloss.Style {
			style := lipgloss.NewStyle()
			if row == table.HeaderRow {
				style = p.Secondary
			}
			if col == 0 {
				return style.Width(column + 4)
			}
			return style.Width(column)
		}).Wrap(false).String()
}
