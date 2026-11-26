package app

import (
	"strings"
	"testing"
	"time"

	"github.com/gaurv-m/keywi/assets"
	"github.com/gaurv-m/keywi/internal/casual"
	"github.com/gaurv-m/keywi/internal/statistics"
	"github.com/gaurv-m/keywi/internal/storage"
	"github.com/gaurv-m/keywi/internal/typing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/ansi"
)

func TestPreferenceToastLastsTwoSeconds(t *testing.T) {
	m := model(t)
	m.screen = settings
	now := time.Now()
	m.now = now
	m.saveSettings()
	m.advance(now.Add(1999 * time.Millisecond))
	if m.message != "Preferences saved" {
		t.Fatal("toast expired early")
	}
	m.now = now.Add(time.Second)
	m.saveSettings()
	m.advance(now.Add(2 * time.Second))
	if m.message != "Preferences saved" {
		t.Fatal("second save did not refresh the toast")
	}
	m.advance(now.Add(3 * time.Second))
	if m.message != "" {
		t.Fatal("toast exceeded two seconds")
	}
}

func TestCasualSetupSurvivesRelaunch(t *testing.T) {
	m := model(t)
	m.screen = configuration
	m = key(m, tea.KeyMsg{Type: tea.KeyRight})
	m = key(m, tea.KeyMsg{Type: tea.KeyDown})
	m = key(m, tea.KeyMsg{Type: tea.KeyRight})
	reopened, _, err := storage.OpenAt(m.Store.ConfigDir, m.Store.DataDir)
	if err != nil {
		t.Fatal(err)
	}
	next := New(reopened, nil)
	if next.config.Kind != "words" || next.config.Amount != 50 {
		t.Fatal("last setup was lost", next.config)
	}
	reopened.Preferences.CasualKind, reopened.Preferences.CasualAmount = "time", 999
	reopened.Preferences.Normalize()
	if reopened.Preferences.CasualAmount != 30 {
		t.Fatal("invalid preset was not normalized")
	}
}

func TestRestartKeepsRemainingDeck(t *testing.T) {
	m := model(t)
	seen := map[string]bool{}
	for range assets.Sentences()[1] {
		m.start("textattack")
		text := string(m.blitzGame.Tracker.Target)
		if seen[text] {
			t.Fatal("Blitz restart repeated a sentence before exhaustion")
		}
		seen[text] = true
	}
	seen = map[string]bool{}
	for range 8 {
		m.config = casual.DefaultConfig()
		m.start("normal")
		for _, text := range m.paragraphDeck.Pools[1] {
			if strings.Contains(string(m.casualGame.Tracker.Target), text) {
				if seen[text] {
					t.Fatal("Casual restart repeated a paragraph before exhaustion")
				}
				seen[text] = true
			}
		}
	}
	if len(seen) != 16 {
		t.Fatal("paragraph batches changed", len(seen))
	}
}

func TestHelpReturnsAndDoesNotStealTyping(t *testing.T) {
	m := model(t)
	m.screen = settings
	question := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}}
	m = key(m, question)
	if m.screen != helpScreen {
		t.Fatal("help did not open")
	}
	if !strings.Contains(ansi.Strip(m.View()), "Made by G") {
		t.Fatal("missing credit")
	}
	for range 100 {
		m = key(m, tea.KeyMsg{Type: tea.KeyDown})
	}
	if m.helpOffset != m.helpMaxOffset() {
		t.Fatal("help scroll escaped its bounds")
	}
	if !strings.Contains(ansi.Strip(m.View()), "ctrl+c") {
		t.Fatal("last help section is unreachable")
	}
	m = key(m, tea.KeyMsg{Type: tea.KeyEsc})
	if m.screen != settings {
		t.Fatal("help lost the previous screen")
	}
	m.start("textattack")
	m.blitzGame.Tracker = typing.New("?")
	m.blitzGame.Rewarded = make([]bool, 1)
	m = key(m, question)
	if m.screen != play || m.blitzGame.Sentences != 1 {
		t.Fatal("help stole a gameplay character")
	}
}

func TestHelpUsesQuestionMarkNotSlash(t *testing.T) {
	m := model(t)
	m = key(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
	if m.screen != menu {
		t.Fatal("unshifted slash opened help")
	}
	m = key(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}})
	if m.screen != helpScreen {
		t.Fatal("question mark did not open help")
	}
	m = key(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}})
	if m.screen != menu {
		t.Fatal("question mark did not return from help")
	}
}

func TestResultIsSimpleAndComparisonMatchesSetup(t *testing.T) {
	m := model(t)
	m.screen = results
	now := time.Now()
	m.result = statistics.Result{ID: "current", At: now, Mode: "normal", Config: casual.DefaultConfig(), Metrics: typing.Metrics{WPM: 80, Accuracy: 98, RawWPM: 100, Consistency: 75, Samples: []typing.Sample{{Seconds: 1, WPM: 70}, {Seconds: 2, WPM: 90}}}}
	previous := m.result
	previous.ID, previous.At, previous.Metrics.WPM = "previous", now.Add(-time.Minute), 72
	different := previous
	different.ID, different.Config.Amount, different.Metrics.WPM = "different", 15, 150
	m.Store.Data.Results = []statistics.Result{previous, different, m.result}
	view := ansi.Strip(m.View())
	for _, want := range []string{"80.0 WPM", "Accuracy 98.0%", "+8.0 WPM Vs Last Run", "30s · Paragraphs", "Speed · WPM Over Time", "│", "└"} {
		if !strings.Contains(view, want) {
			t.Fatal("missing result", want)
		}
	}
	for _, clutter := range []string{"Raw ", "Consistency", "Incorrect", "Errors"} {
		if strings.Contains(view, clutter) {
			t.Fatal("result has too much detail", clutter)
		}
	}
	m.screen = detail
	view = ansi.Strip(m.View())
	if !strings.Contains(view, "Raw 100.0") || !strings.Contains(view, "Consistency 75.0%") {
		t.Fatal("history lost its detailed metrics")
	}
}

func TestSettingsStayCompactWithHorizontalSpacing(t *testing.T) {
	m := model(t)
	m.screen = settings
	rows := strings.Split(ansi.Strip(m.View()), "\n")
	positions := map[string]int{}
	for y, row := range rows {
		for _, label := range []string{"Theme", "Cursor", "Animations"} {
			if strings.Contains(row, label) {
				positions[label] = y
			}
		}
	}
	if positions["Cursor"]-positions["Theme"] != 1 || positions["Animations"]-positions["Cursor"] != 1 {
		t.Fatal("settings gained unwanted vertical gaps")
	}
	view := ansi.Strip(m.View())
	for _, label := range []string{"Appearance", "Feedback", "Word tests", "Data", "Reset statistics"} {
		if !strings.Contains(view, label) {
			t.Fatal("compact settings must show every group at 80×24", label)
		}
	}
}

func TestRunChartsFitBothModes(t *testing.T) {
	m := model(t)
	m.screen = results
	m.personalBest = true
	for _, mode := range []string{"normal", "textattack"} {
		m.result = statistics.Result{
			ID: "run", Mode: mode, At: time.Now(), Config: casual.DefaultConfig(), Sentences: 8, Level: 2,
			Metrics: typing.Metrics{WPM: 80, Accuracy: 97, Duration: 30, Samples: []typing.Sample{{Seconds: 1, WPM: 70}, {Seconds: 2, WPM: 90}}},
		}
		previous := m.result
		previous.ID, previous.At = "previous", m.result.At.Add(-time.Minute)
		m.Store.Data.Results = []statistics.Result{previous, m.result}
		for _, size := range [][2]int{{24, 10}, {40, 12}, {60, 18}, {80, 24}, {100, 30}} {
			m.Width, m.Height = size[0], size[1]
			view := ansi.Strip(m.View())
			rows := strings.Split(view, "\n")
			if len(rows) != m.Height || !strings.Contains(rows[m.Height-2], "enter replay") {
				t.Fatal("run summary displaced the navigation", mode, size)
			}
			for _, row := range rows {
				if ansi.StringWidth(row) > m.Width {
					t.Fatal("run summary overflowed", mode, size)
				}
			}
			if m.Height >= 18 && (!strings.Contains(view, "Speed · WPM Over Time") || !strings.Contains(view, "└")) {
				t.Fatal("run summary lost its chart", mode, size)
			}
			if !strings.Contains(view, "Accuracy 97.0%") {
				t.Fatal("run summary lost accuracy", mode, size)
			}
			if !strings.Contains(view, "80.0 WPM") {
				t.Fatal("run summary lost numeric WPM", mode, size)
			}
			if mode == "textattack" && !strings.Contains(view, "8 Sentences") {
				t.Fatal("Blitz summary lost its sentence score", size)
			}
		}
	}
}
