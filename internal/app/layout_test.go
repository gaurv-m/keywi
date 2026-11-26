package app

import (
	"strings"
	"testing"
	"time"

	"github.com/gaurv-m/keywi/internal/casual"
	"github.com/gaurv-m/keywi/internal/statistics"
	"github.com/gaurv-m/keywi/internal/theme"
	"github.com/gaurv-m/keywi/internal/typing"
	"github.com/gaurv-m/keywi/internal/ui"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/ansi"
)

func TestDeadlineKeyDoesNotActOnResults(t *testing.T) {
	for _, mode := range []string{"normal", "textattack"} {
		for _, kind := range []tea.KeyType{tea.KeyEnter, tea.KeyEsc, tea.KeySpace, tea.KeyBackspace} {
			for _, confirm := range []string{"", "abandon", "restart"} {
				m := model(t)
				m.start(mode)
				past := time.Now().Add(-time.Minute)
				if mode == "normal" {
					m.casualGame.Type(m.tracker().Target[0], past)
				} else {
					m.blitzGame.Type(m.tracker().Target[0], past)
				}
				m.confirm, m.confirmYes = confirm, true
				m = key(m, tea.KeyMsg{Type: kind})
				if m.screen != results || m.confirm != "" || len(m.Store.Data.Results) != 1 {
					t.Fatal("deadline key escaped or replayed the results", mode, kind, confirm)
				}
				if m.tracker().Ended.IsZero() {
					t.Fatal("expired tracker was replaced", mode, kind, confirm)
				}
			}
		}
	}
}

func TestAuditLabelsAndHelpFit(t *testing.T) {
	m := model(t)
	m.start("normal")
	m.casualGame = casual.New(casual.Config{Kind: "words", Amount: 10})
	m.config = m.casualGame.Config
	m.casualGame.Tracker = typing.New("a b c d e f g h i j")
	for _, r := range "a b c d e f g h i " {
		m.casualGame.Type(r, time.Now())
	}
	if !strings.Contains(ansi.Strip(m.View()), "1 word left") {
		t.Fatal("last word label is unclear")
	}
	m.screen = results
	m.result = statistics.Result{Mode: "textattack", Sentences: 1, Level: 1, Metrics: typing.Metrics{WPM: 70, Accuracy: 99, Duration: 30}}
	view := ansi.Strip(m.View())
	if !strings.Contains(view, "1 Sentence") || strings.Contains(view, "1 Sentences") {
		t.Fatal("singular sentence score is wrong")
	}
	m.screen = helpScreen
	rows := helpRows(64, theme.New("kiwi"))
	if len(rows) > 14 || !strings.Contains(ui.Plain(strings.Join(rows, "\n")), "fixed mistakes") {
		t.Fatal("scoring help is missing or too long")
	}
}

func TestConfirmationHelpMatchesTheRunClock(t *testing.T) {
	for _, mode := range []string{"normal", "textattack"} {
		m := model(t)
		if mode == "normal" {
			m.config.Kind, m.config.Amount = "words", 10
		}
		m.start(mode)
		m = key(m, tea.KeyMsg{Type: tea.KeyEsc})
		if !strings.Contains(ansi.Strip(m.View()), "run hasn't started") {
			t.Fatal("idle confirmation claims the clock is running", mode)
		}
		m = key(m, tea.KeyMsg{Type: tea.KeyEsc})
		m = key(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{m.tracker().Target[0]}})
		m = key(m, tea.KeyMsg{Type: tea.KeyEsc})
		view := ansi.Strip(m.View())
		want := "The timer keeps running."
		if mode == "normal" {
			want = "Your elapsed time keeps running."
		}
		if !strings.Contains(view, want) {
			t.Fatal("confirmation describes the wrong clock", mode)
		}
	}
}

func TestSoundToggleOnlyQueuesBlitzPlaybackWhenEnabled(t *testing.T) {
	for _, enabled := range []bool{false, true} {
		m := model(t)
		m.Store.Preferences.Sound = enabled
		m.start("textattack")
		target := append([]rune(nil), m.tracker().Target...)
		next, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: target})
		m = next.(Model)
		if m.blitzGame.Sentences != 1 || (cmd != nil) != enabled {
			t.Fatal("Sound preference did not control playback", enabled)
		}
	}
}
