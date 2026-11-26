package app

import (
	"fmt"
	"strings"

	"github.com/gaurv-m/keywi/internal/graph"
	"github.com/gaurv-m/keywi/internal/theme"
	"github.com/gaurv-m/keywi/internal/typing"
	"github.com/gaurv-m/keywi/internal/ui"
)

func graphStyles(p theme.Palette) graph.Styles {
	return graph.Styles{Line: p.Primary.Bold(true), Labels: p.Correct, Grid: p.Muted.Faint(true), Average: p.Secondary}
}

func speedWindow(mt typing.Metrics) int {
	duration := mt.Duration
	for _, sample := range mt.Samples {
		if sample.Seconds > duration {
			duration = sample.Seconds
		}
	}
	if duration > 30 {
		return 5
	}
	return 3
}

func speedHeading(mt typing.Metrics, width int, p theme.Palette) string {
	label, detail := "Speed · WPM Over Time", fmt.Sprintf("%ds Average", speedWindow(mt))
	if ui.Width(label)+len(detail)+3 > width {
		label = "Speed · WPM"
	}
	gap := max(1, width-ui.Width(label)-len(detail)-2)
	return p.Secondary.Render(label) + p.Muted.Faint(true).Render(" "+strings.Repeat("─", gap)+" ") + p.Correct.Render(detail)
}

func (m Model) speedGraph(mt typing.Metrics, w, h int) string {
	points := make([]graph.Point, len(mt.Samples))
	for i, sample := range mt.Samples {
		points[i] = graph.Point{X: sample.Seconds, Y: sample.WPM}
	}
	points = graph.Smooth(points, float64(speedWindow(mt)))
	return graph.Draw(points, w, max(4, h), graph.Options{XMax: mt.Duration, XUnit: "s", Average: &mt.WPM}, graphStyles(theme.New(m.Store.Preferences.Theme)))
}
