package app

import (
	"fmt"
	"strings"

	"github.com/gaurv-m/keywi/internal/statistics"
	"github.com/gaurv-m/keywi/internal/theme"
	"github.com/gaurv-m/keywi/internal/ui"
)

func (m Model) resultsView(width int, p theme.Palette) string {
	r := m.result
	panelWidth := min(60, width)
	center := func(s string) string { return ui.CenterLine(s, width) }
	heading, score := "Casual Results", fmt.Sprintf("%.1f WPM", r.Metrics.WPM)
	setup := titleLabel(r.Config.ShortLabel())
	if r.Config.Text == "prose" {
		setup += " · Paragraphs"
	}
	if r.Mode == "textattack" {
		heading = "Blitz Results"
		score = fmt.Sprintf("%d %s · Level %d", r.Sentences, plural(r.Sentences, "Sentence"), r.Level)
		setup = displayDuration(r.Metrics.Duration) + " Survived"
	}
	rows := []string{center(p.Accent(heading)), center(p.Muted.Render(setup)), ""}
	if ui.Width(score) > width {
		rows = append(rows, center(p.Primary.Render(fmt.Sprintf("%d %s", r.Sentences, plural(r.Sentences, "Sentence")))), center(p.Primary.Render(fmt.Sprintf("Level %d", r.Level))))
	} else {
		rows = append(rows, center(p.Primary.Render(score)))
	}
	accuracy := fmt.Sprintf("Accuracy %.1f%%", r.Metrics.Accuracy)
	if r.Mode == "textattack" {
		wpm := fmt.Sprintf("%.1f WPM", r.Metrics.WPM)
		if ui.Width(wpm+" · "+accuracy) <= width {
			accuracy = wpm + " · " + accuracy
		} else {
			rows = append(rows, center(p.Correct.Render(wpm)))
		}
	}
	rows = append(rows, center(p.Correct.Render(accuracy)))
	previous, hasPrevious := previousResult(m.Store.Data.Results, r)
	if hasPrevious && m.Height >= 16 {
		comparison := fmt.Sprintf("%+.1f WPM Vs Last Run", r.Metrics.WPM-previous.Metrics.WPM)
		if r.Mode == "textattack" {
			difference := r.Sentences - previous.Sentences
			comparison = fmt.Sprintf("%+d %s Vs Last Run", difference, plural(difference, "Sentence"))
		}
		if m.Height >= 22 && width >= 44 {
			best, _ := bestBeforeResult(m.Store.Data.Results, r)
			panel := comparisonTable(resultScore(previous), resultScore(best), panelWidth, p)
			rows = append(rows, "")
			rows = append(rows, strings.Split(ui.CenterBlock(panel, width), "\n")...)
		}
		if ui.Width(comparison) > width {
			comparison = strings.ReplaceAll(comparison, " Vs Last Run", " Vs Last")
		}
		if ui.Width(comparison) <= width {
			rows = append(rows, center(p.Secondary.Render(comparison)))
		}
	}
	if m.personalBest && m.Height >= 14 {
		rows = append(rows, center(p.Bonus.Render("New Personal Best")))
	}
	chartHeight := min(10, m.Height-3-len(rows)-2)
	if m.Height >= 18 && width >= 32 && chartHeight >= 4 && len(r.Metrics.Samples) > 0 {
		rows = append(rows, "", ui.CenterBlock(speedHeading(r.Metrics, panelWidth, p), width), ui.CenterBlock(m.speedGraph(r.Metrics, panelWidth, chartHeight), width))
	}
	return strings.Join(rows, "\n")
}

func resultScore(r statistics.Result) string {
	if r.Mode == "textattack" {
		return fmt.Sprintf("%d %s · Level %d", r.Sentences, plural(r.Sentences, "Sentence"), r.Level)
	}
	return fmt.Sprintf("%.1f WPM", r.Metrics.WPM)
}

func bestBeforeResult(results []statistics.Result, current statistics.Result) (statistics.Result, bool) {
	var best statistics.Result
	found := false
	for _, r := range results {
		if r.ID == current.ID || r.Mode != current.Mode || r.At.After(current.At) {
			continue
		}
		if r.Mode == "normal" && r.Config.Key() != current.Config.Key() {
			continue
		}
		if !found || (r.Mode == "normal" && r.Metrics.WPM > best.Metrics.WPM) || (r.Mode == "textattack" && (r.Sentences > best.Sentences || (r.Sentences == best.Sentences && r.Metrics.WPM > best.Metrics.WPM))) {
			best, found = r, true
		}
	}
	return best, found
}

func previousResult(results []statistics.Result, current statistics.Result) (statistics.Result, bool) {
	var previous statistics.Result
	found := false
	for _, r := range results {
		if r.ID == current.ID || r.Mode != current.Mode || r.At.After(current.At) {
			continue
		}
		if r.Mode == "normal" && r.Config.Key() != current.Config.Key() {
			continue
		}
		if !found || r.At.After(previous.At) {
			previous, found = r, true
		}
	}
	return previous, found
}

func (m Model) detailView(width int, p theme.Palette) string {
	r, mt := m.result, m.result.Metrics
	center := func(s string) string { return ui.CenterLine(s, width) }
	heading := "Casual Results"
	if r.Mode == "textattack" {
		heading = "Blitz Results"
	}
	rows := []string{center(p.Accent(heading)), "", center(p.Primary.Render(fmt.Sprintf("%.1f WPM", mt.WPM))), center(fmt.Sprintf("Accuracy %.1f%%", mt.Accuracy))}
	if r.Mode == "textattack" {
		rows = append(rows, center(fmt.Sprintf("%d %s · Level %d", r.Sentences, plural(r.Sentences, "Sentence"), r.Level)))
	}
	if m.Height >= 14 {
		rows = append(rows, "", center(p.Muted.Render(fmt.Sprintf("Raw %.1f WPM · Consistency %.1f%%", mt.RawWPM, mt.Consistency))), center(p.Muted.Render(fmt.Sprintf("%d Correct · %d Incorrect", mt.Correct, mt.Incorrect))), center(p.Muted.Render(fmt.Sprintf("%d %s · %s", mt.Errors, plural(mt.Errors, "Error"), displayDuration(mt.Duration)))))
		if r.Mode == "normal" && width >= 42 {
			rows = append(rows, center(p.Muted.Render(titleLabel(r.Config.Label()))))
		}
	}
	chartHeight := min(10, m.Height-3-len(rows)-2)
	if m.Height >= 24 && width >= 32 && chartHeight >= 4 {
		rows = append(rows, "", speedHeading(mt, width, p), m.speedGraph(mt, width, chartHeight))
	}
	return strings.Join(rows, "\n")
}

func plural(count int, noun string) string {
	if count == 1 || count == -1 {
		return noun
	}
	return noun + "s"
}
