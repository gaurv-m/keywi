package app

import (
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"github.com/gaurv-m/keywi/internal/casual"
	"github.com/gaurv-m/keywi/internal/graph"
	"github.com/gaurv-m/keywi/internal/statistics"
	"github.com/gaurv-m/keywi/internal/theme"
	"github.com/gaurv-m/keywi/internal/ui"

	"github.com/charmbracelet/x/ansi"
)

func displayDuration(seconds float64) string {
	n := max(0, int(math.Round(seconds)))
	if n >= 3600 {
		return fmt.Sprintf("%dh %02dm", n/3600, n%3600/60)
	}
	if n >= 60 {
		return fmt.Sprintf("%dm %02ds", n/60, n%60)
	}
	return fmt.Sprintf("%ds", n)
}

// statPair anchors values to a common right edge, with a readable label column.
func statPair(label, value string, width int, p theme.Palette, record bool) string {
	value = ansi.Truncate(value, max(1, width-3), "…")
	label = ansi.Truncate(label, max(1, width-ui.Width(value)-2), "…")
	style := p.Correct
	if record {
		style = p.Primary
	}
	return p.Muted.Render(label) + strings.Repeat(" ", max(2, width-ui.Width(label)-ui.Width(value))) + style.Render(value)
}

type personalBestRow struct {
	config casual.Config
	wpm    float64
}

func casualBestRows(results []statistics.Result) []personalBestRow {
	records := map[string]personalBestRow{}
	for _, r := range results {
		if r.Mode != "normal" {
			continue
		}
		key := r.Config.Key()
		old, ok := records[key]
		if !ok || r.Metrics.WPM > old.wpm {
			records[key] = personalBestRow{r.Config, r.Metrics.WPM}
		}
	}
	out := make([]personalBestRow, 0, len(records))
	for _, row := range records {
		out = append(out, row)
	}
	sort.Slice(out, func(i, j int) bool {
		a, b := out[i].config, out[j].config
		if a.Kind != b.Kind {
			return a.Kind == "time"
		}
		if a.Amount != b.Amount {
			return a.Amount < b.Amount
		}
		if a.Text != b.Text {
			return a.Text == "prose"
		}
		return a.Key() < b.Key()
	})
	return out
}

func (m Model) statisticsView(width int, p theme.Palette) string {
	center := func(text string) string { return ui.CenterLine(text, width) }
	casual, attack := "Casual", "Blitz"
	if m.statsTab == 0 {
		casual = p.Primary.Render("[Casual]")
		attack = p.Muted.Render(attack)
	} else {
		casual = p.Muted.Render(casual)
		attack = p.Primary.Render("[Blitz]")
	}
	header := ui.Join(center(p.Accent("Statistics")), "", center(casual+"     "+attack), "")
	panelWidth := min(68, width)
	available := max(1, m.Height-7)
	mode := "normal"
	if m.statsTab == 1 {
		mode = "textattack"
	}
	summary := statistics.Aggregate(m.Store.Data.Results, mode)
	lines := []string{sectionHeading("Overview", panelWidth, p)}
	if summary.Count == 0 {
		lines = append(lines, p.Muted.Render("No completed runs yet."))
		if available >= 4 {
			lines = append(lines, "", p.Muted.Render("Complete a run to start your statistics."))
		}
		return header + "\n" + ui.CenterBlock(boundStats(lines, panelWidth, available), width)
	}
	type metric struct{ label, value string }
	var left, right []metric
	if mode == "normal" {
		left = []metric{{"Average Speed", fmt.Sprintf("%.1f WPM", summary.AverageWPM)}, {"Accuracy", fmt.Sprintf("%.1f%%", summary.AverageAccuracy)}, {"Consistency", fmt.Sprintf("%.1f%%", summary.AverageConsistency)}}
		right = []metric{{"Best Speed", fmt.Sprintf("%.1f WPM", summary.BestWPM)}, {"Completed", fmt.Sprint(summary.Count)}, {"Typing Time", displayDuration(summary.TotalTime)}}
	} else {
		last := time.Time{}
		for _, r := range m.Store.Data.Results {
			if r.Mode == mode && r.At.After(last) {
				last = r.At
			}
		}
		played := "—"
		if !last.IsZero() {
			played = last.Local().Format("Jan 02")
		}
		left = []metric{{"Average Score", fmt.Sprintf("%.1f Sentences", summary.AverageSentences)}, {"Average Speed", fmt.Sprintf("%.1f WPM", summary.AverageWPM)}, {"Accuracy", fmt.Sprintf("%.1f%%", summary.AverageAccuracy)}}
		right = []metric{{"Completed", fmt.Sprint(summary.Count)}, {"Typing Time", displayDuration(summary.TotalTime)}, {"Last Played", played}}
	}
	if panelWidth >= 52 {
		leftRows, rightRows := make([][]string, len(left)), make([][]string, len(right))
		for i := range left {
			leftRows[i] = []string{left[i].label, left[i].value}
			rightRows[i] = []string{right[i].label, right[i].value}
		}
		lines = strings.Split(overviewTable(leftRows, rightRows, panelWidth, p), "\n")
	} else {
		for i := range left {
			if len(lines) < available {
				lines = append(lines, statPair(left[i].label, left[i].value, panelWidth, p, false))
			}
			if len(lines) < available {
				lines = append(lines, statPair(right[i].label, right[i].value, panelWidth, p, false))
			}
		}
	}
	if mode == "normal" {
		bests := casualBestRows(m.Store.Data.Results)
		if len(bests) > 0 && available-len(lines) >= 4 {
			visible := casualBestVisible(len(bests), width, m.Height)
			start := min(m.bestOffset, max(0, len(bests)-visible))
			heading := "Personal Bests"
			if len(bests) > visible {
				heading += fmt.Sprintf("  %d–%d of %d", start+1, start+visible, len(bests))
			}
			lines = append(lines, "", sectionHeading(heading, panelWidth, p), statPair("Test", "WPM", panelWidth, p, false))
			for _, best := range bests[start : start+visible] {
				lines = append(lines, statPair(titleLabel(best.config.Label()), fmt.Sprintf("%.1f", best.wpm), panelWidth, p, true))
			}
		}
	} else if available-len(lines) >= 5 {
		lines = append(lines, "", sectionHeading("Personal Bests", panelWidth, p), statPair("Sentence Score", fmt.Sprintf("%d %s", summary.BestSentences, plural(summary.BestSentences, "Sentence")), panelWidth, p, true), statPair("Highest Level", fmt.Sprint(summary.HighestLevel), panelWidth, p, true), statPair("Longest Run", displayDuration(summary.LongestSurvival), panelWidth, p, true))
	}
	var recent []statistics.Result
	for _, r := range m.Store.Data.Results {
		if r.Mode == mode {
			recent = append(recent, r)
		}
	}
	if len(recent) > 20 {
		recent = recent[len(recent)-20:]
	}
	remaining := available - len(lines)
	if len(recent) >= 2 && remaining >= 6 && panelWidth >= 32 {
		plotHeight := min(9, remaining-2)
		// Leave one quiet line for the averaging note when there is room.
		if mode == "normal" && plotHeight > 4 {
			plotHeight--
		}
		title := fmt.Sprintf("Recent Speed · WPM · %d Tests", len(recent))
		if mode == "textattack" {
			title = fmt.Sprintf("Recent Score · Sentences · %d Runs", len(recent))
		}
		points := make([]graph.Point, len(recent))
		for i, r := range recent {
			value := r.Metrics.WPM
			if mode == "textattack" {
				value = float64(r.Sentences)
			}
			points[i] = graph.Point{X: float64(i + 1), Y: value}
		}
		average := 0.0
		for _, point := range points {
			average += point.Y
		}
		average /= float64(len(points))
		lines = append(lines, "", sectionHeading(title, panelWidth, p))
		options := graph.Options{XMin: 1, XUnit: "run", Average: &average}
		lines = append(lines, strings.Split(graph.Draw(points, panelWidth, plotHeight, options, graphStyles(p)), "\n")...)
	} else if len(recent) == 1 && remaining >= 2 {
		lines = append(lines, "", p.Muted.Render("Complete another run to see a trend."))
	}
	if mode == "normal" && available > len(lines) {
		lines = append(lines, p.Muted.Render("Averages: All Casual Tests · Bests: By Setup"))
	}
	return header + "\n" + ui.CenterBlock(boundStats(lines, panelWidth, available), width)
}

func casualBestVisible(count, width, height int) int {
	available := max(1, height-7)
	used := 4
	if min(68, width) < 52 {
		used = min(7, available)
	}
	if available-used < 4 {
		return 0
	}
	return min(count, max(1, min(4, height-22)), available-used-3)
}

func boundStats(lines []string, width, height int) string {
	if len(lines) > height {
		lines = lines[:height]
	}
	for i, line := range lines {
		line = ansi.Truncate(line, width, "…")
		lines[i] = line + strings.Repeat(" ", max(0, width-ui.Width(line)))
	}
	return strings.Join(lines, "\n")
}

func (m Model) historyView(width int, p theme.Palette) string {
	name := "Casual"
	if m.statsTab == 1 {
		name = "Blitz"
	}
	list := m.filtered()
	panelWidth := min(68, width)
	lines := []string{}
	if len(list) == 0 {
		lines = append(lines, p.Muted.Render("No completed runs yet."))
	} else {
		wide := panelWidth >= 54
		heading := "Date       Score   Acc"
		if wide {
			heading = fmt.Sprintf("  %-14s %-14s %9s %9s", "Date", "Test", "WPM", "Accuracy")
			if m.statsTab == 1 {
				heading = fmt.Sprintf("  %-14s %-14s %9s %9s", "Date", "Sentences", "Level", "Duration")
			}
		}
		lines = append(lines, p.Muted.Render(heading))
		visible := min(10, max(1, m.Height-9))
		start := max(0, min(m.historyIndex-visible/2, len(list)-visible))
		for i := start; i < min(len(list), start+visible); i++ {
			r := list[i]
			text := ""
			if wide {
				if m.statsTab == 0 {
					text = fmt.Sprintf("%-14s %-14s %9.1f %8.1f%%", r.At.Local().Format("Jan 02 15:04"), titleLabel(r.Config.ShortLabel()), r.Metrics.WPM, r.Metrics.Accuracy)
				} else {
					text = fmt.Sprintf("%-14s %-14d %9d %9s", r.At.Local().Format("Jan 02 15:04"), r.Sentences, r.Level, displayDuration(r.Metrics.Duration))
				}
			} else {
				score := fmt.Sprintf("%.0f WPM", r.Metrics.WPM)
				if m.statsTab == 1 {
					score = fmt.Sprintf("%d Sent", r.Sentences)
				}
				text = fmt.Sprintf("%s  %s  %.0f%%", r.At.Local().Format("01/02"), score, r.Metrics.Accuracy)
			}
			if i == m.historyIndex {
				text = p.Primary.Render("› " + text)
			} else {
				text = p.Muted.Render("  " + text)
			}
			lines = append(lines, text)
		}
	}
	return ui.Join(ui.CenterLine(p.Accent(name+" History"), width), "", ui.CenterBlock(boundStats(lines, panelWidth, max(1, m.Height-5)), width))
}
