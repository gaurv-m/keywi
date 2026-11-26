package app

import (
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/gaurv-m/keywi/internal/casual"
	"github.com/gaurv-m/keywi/internal/statistics"
	"github.com/gaurv-m/keywi/internal/typing"

	"github.com/charmbracelet/x/ansi"
)

func statsFixture(t *testing.T) Model {
	t.Helper()
	m := model(t)
	m.screen = stats
	for i, amount := range []int{120, 15, 60, 30} {
		m.Store.Data.Results = append(m.Store.Data.Results, statistics.Result{ID: string(rune('a' + i)), At: time.Date(2026, 11, 24, 12, i, 0, 0, time.UTC), Mode: "normal", Config: casual.Config{Kind: "time", Amount: amount, Text: "prose", Punctuation: true, Capitalization: true}, Metrics: typing.Metrics{WPM: float64(70 + i), Accuracy: 96, Consistency: 80, Duration: float64(amount)}})
	}
	for i, score := range []int{4, 9} {
		m.Store.Data.Results = append(m.Store.Data.Results, statistics.Result{ID: string(rune('x' + i)), At: time.Date(2026, 11, 24, 13, i, 0, 0, time.UTC), Mode: "textattack", Sentences: score, Level: 1 + i, Metrics: typing.Metrics{WPM: 80, Accuracy: 98, Duration: float64(45 + i*15)}})
	}
	return m
}
func TestStatisticsSectionsAndNumericPresetOrder(t *testing.T) {
	m := statsFixture(t)
	before := append([]statistics.Result(nil), m.Store.Data.Results...)
	rows := casualBestRows(m.Store.Data.Results)
	for i, want := range []int{15, 30, 60, 120} {
		if rows[i].config.Amount != want {
			t.Fatal("presets not sorted numerically")
		}
	}
	view := ansi.Strip(m.View())
	for _, label := range []string{"Overview", "Activity", "Average Speed", "71.5 WPM", "96.0%", "Consistency", "80.0%", "Completed", "Typing Time", "Personal Bests", "Recent Speed", "Averages: All Casual Tests"} {
		if !strings.Contains(view, label) {
			t.Fatal("missing statistics section", label)
		}
	}
	m.statsTab = 1
	view = ansi.Strip(m.View())
	for _, label := range []string{"Average Score", "6.5 Sentences", "80.0 WPM", "98.0%", "Last Played", "Sentence Score", "9 Sentences", "Highest Level", "Longest Run", "Recent Score"} {
		if !strings.Contains(view, label) {
			t.Fatal("missing survival statistic", label)
		}
	}
	if !reflect.DeepEqual(before, m.Store.Data.Results) {
		t.Fatal("presentation mutated scores")
	}
}
func TestStatisticsAndHistoryFitPopulatedSmallScreens(t *testing.T) {
	m := statsFixture(t)
	for _, size := range [][2]int{{24, 10}, {40, 12}, {60, 20}, {80, 24}, {100, 30}} {
		m.Width, m.Height = size[0], size[1]
		for _, tab := range []int{0, 1} {
			m.statsTab = tab
			for _, screen := range []screen{stats, history} {
				m.screen = screen
				lines := strings.Split(m.View(), "\n")
				if len(lines) != m.Height {
					t.Fatal("screen height changed")
				}
				for _, line := range lines {
					if ansi.StringWidth(line) > m.Width {
						t.Fatal("statistics clipped")
					}
				}
				if strings.TrimSpace(ansi.Strip(lines[m.Height-2])) == "" {
					t.Fatal("footer disappeared")
				}
			}
		}
	}
}
func TestSingleRunHasNoDecorativeTrend(t *testing.T) {
	m := statsFixture(t)
	m.Store.Data.Results = m.Store.Data.Results[:1]
	view := ansi.Strip(m.View())
	if strings.Contains(view, "Recent Speed") || strings.ContainsAny(view, "⠂⠄⣿") {
		t.Fatal("trend shown before a second result")
	}
	if !strings.Contains(view, "Complete another run") {
		t.Fatal("missing empty-trend explanation")
	}
	m.Store.Data.Results = nil
	view = ansi.Strip(m.View())
	if !strings.Contains(view, "No completed runs yet.") {
		t.Fatal("missing empty state")
	}
}
