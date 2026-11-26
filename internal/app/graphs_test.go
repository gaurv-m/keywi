package app

import (
	"reflect"
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
	"github.com/gaurv-m/keywi/internal/theme"
	"github.com/gaurv-m/keywi/internal/typing"
)

func TestSpeedChartsLabelSmoothingAndRetainRecordedMetrics(t *testing.T) {
	m := model(t)
	for _, seconds := range []int{15, 30, 60, 120} {
		mt := typing.Metrics{Duration: float64(seconds), WPM: 52.1}
		for second := 1; second <= seconds; second++ {
			mt.Samples = append(mt.Samples, typing.Sample{Seconds: float64(second), WPM: float64((second % 3) * 48)})
		}
		before := mt
		before.Samples = append([]typing.Sample(nil), mt.Samples...)
		window := "3s Average"
		if seconds > 30 {
			window = "5s Average"
		}
		heading := ansi.Strip(speedHeading(mt, 60, theme.New("kiwi")))
		plot := ansi.Strip(m.speedGraph(mt, 60, 10))
		if !strings.Contains(heading, window) || !strings.Contains(plot, "Avg 52.1") {
			t.Fatal("speed chart did not explain the averaged line", heading, plot)
		}
		if !reflect.DeepEqual(mt, before) {
			t.Fatal("rendering changed WPM or recorded samples")
		}
	}
}

func TestStatisticsReferenceUsesOnlyVisibleRecentRuns(t *testing.T) {
	m := statsFixture(t)
	m.Width, m.Height = 100, 30
	m.Store.Data.Results = m.Store.Data.Results[:1]
	base := m.Store.Data.Results[0]
	for i := 0; i < 20; i++ {
		r := base
		r.Metrics.WPM = 80
		m.Store.Data.Results = append(m.Store.Data.Results, r)
	}
	m.Store.Data.Results[0].Metrics.WPM = 1000
	view := ansi.Strip(m.View())
	if !strings.Contains(view, "Avg 80.0") {
		t.Fatal("recent chart mixed in older runs", view)
	}
}

func TestHistoryDetailsUseTheSameReadableSpeedChart(t *testing.T) {
	for _, mode := range []string{"normal", "textattack"} {
		m := model(t)
		m.Width, m.Height, m.screen = 80, 24, detail
		m.result.Mode = mode
		m.result.Config.Kind, m.result.Config.Amount, m.result.Config.Text = "time", 120, "prose"
		m.result.Metrics = typing.Metrics{WPM: 52.1, Accuracy: 91.4, Duration: 120, Samples: []typing.Sample{{Seconds: 1, WPM: 48}, {Seconds: 120, WPM: 54}}}
		view := ansi.Strip(m.View())
		for _, label := range []string{"5s Average", "Avg 52.1", "120s", "Speed · WPM Over Time", "enter / esc history"} {
			if !strings.Contains(view, label) {
				t.Fatal("history chart lost its labels or navigation", mode, label, view)
			}
		}
	}
}
