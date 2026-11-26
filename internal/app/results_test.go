package app

import (
	"strings"
	"testing"
	"time"

	"github.com/gaurv-m/keywi/internal/casual"
	"github.com/gaurv-m/keywi/internal/statistics"
	"github.com/gaurv-m/keywi/internal/typing"

	"github.com/charmbracelet/x/ansi"
)

func TestResultsSeparatePreviousRunAndEarlierRecord(t *testing.T) {
	for _, mode := range []string{"normal", "textattack"} {
		m := model(t)
		m.Width, m.Height, m.screen = 80, 24, results
		now := time.Now()
		m.result = statistics.Result{ID: "current", At: now, Mode: mode, Config: casual.DefaultConfig(), Sentences: 8, Level: 2,
			Metrics: typing.Metrics{WPM: 80, Accuracy: 97, Duration: 45, Samples: []typing.Sample{{Seconds: 1, WPM: 60}, {Seconds: 45, WPM: 80}}}}
		record := m.result
		record.ID, record.At, record.Sentences, record.Level, record.Metrics.WPM = "record", now.Add(-time.Hour), 15, 4, 100
		previous := record
		previous.ID, previous.At, previous.Sentences, previous.Level, previous.Metrics.WPM = "previous", now.Add(-time.Minute), 6, 2, 72
		future := record
		future.ID, future.At, future.Sentences, future.Metrics.WPM = "future", now.Add(time.Minute), 99, 999
		other := future
		other.ID, other.At, other.Config.Amount = "other setup", now.Add(-time.Second), 15
		m.Store.Data.Results = []statistics.Result{record, previous, m.result, future}
		if mode == "normal" {
			m.Store.Data.Results = append(m.Store.Data.Results, other)
		}
		best, ok := bestBeforeResult(m.Store.Data.Results, m.result)
		if !ok || best.ID != "record" {
			t.Fatal("record included another setup, current run, or future run", mode, best.ID)
		}
		view := ansi.Strip(m.View())
		wants := []string{"Previous Run", "Best Before", "Accuracy 97.0%", "80.0 WPM", "Speed · WPM Over Time", "└"}
		if mode == "textattack" {
			wants = append(wants, "8 Sentences · Level 2", "6 Sentences · Level 2", "15 Sentences · Level 4", "+2 Sentences Vs Last Run", "45s Survived")
		} else {
			wants = append(wants, "72.0 WPM", "100.0 WPM", "+8.0 WPM Vs Last Run")
		}
		for _, want := range wants {
			if !strings.Contains(view, want) {
				t.Fatal("missing labelled result", mode, want, view)
			}
		}
		m.personalBest = true
		for height := 18; height <= 30; height++ {
			m.Height = height
			view = ansi.Strip(m.View())
			if !strings.Contains(view, "New Personal Best") || !strings.Contains(view, "└") {
				t.Fatal("comparison or record badge crowded out the chart", mode, height)
			}
		}
		m.Store.Data.Results = []statistics.Result{m.result, future}
		if _, ok := bestBeforeResult(m.Store.Data.Results, m.result); ok {
			t.Fatal("first run has an earlier record")
		}
		view = ansi.Strip(m.View())
		if strings.Contains(view, "Previous Run") || strings.Contains(view, "Best Before") {
			t.Fatal("first run showed an invented comparison")
		}
	}
}
