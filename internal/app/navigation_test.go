package app

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/ansi"
)

func TestFreshSectionsResetNavigation(t *testing.T) {
	m := statsFixture(t)
	m.screen, m.selection = menu, 3
	m.settingRow = 7
	m.Store.Preferences.Theme, m.Store.Preferences.Cursor = "contrast", "line"
	m.config.Kind, m.config.Amount = "words", 50
	m = key(m, tea.KeyMsg{Type: tea.KeyEnter})
	if m.screen != settings || m.settingRow != 0 {
		t.Fatal("Settings did not open at Theme")
	}
	for range 8 {
		m = key(m, tea.KeyMsg{Type: tea.KeyDown})
	}
	m = key(m, tea.KeyMsg{Type: tea.KeyEsc})
	m = key(m, tea.KeyMsg{Type: tea.KeyEnter})
	if m.settingRow != 0 {
		t.Fatal("Settings kept the old highlighted row")
	}
	if m.Store.Preferences.Theme != "contrast" || m.Store.Preferences.Cursor != "line" {
		t.Fatal("opening a section reset saved choices")
	}
	m = key(m, tea.KeyMsg{Type: tea.KeyEsc})
	m.selection, m.configRow = 0, 1
	m = key(m, tea.KeyMsg{Type: tea.KeyEnter})
	if m.screen != configuration || m.configRow != 0 || m.config.Kind != "words" || m.config.Amount != 50 {
		t.Fatal("Casual did not reset focus while keeping its setup")
	}
	m = key(m, tea.KeyMsg{Type: tea.KeyEsc})
	m.selection, m.statsTab, m.bestOffset, m.historyIndex = 2, 1, 99, 4
	m = key(m, tea.KeyMsg{Type: tea.KeyEnter})
	if m.screen != stats || m.statsTab != 0 || m.bestOffset != 0 || m.historyIndex != 0 {
		t.Fatal("Statistics did not open at the beginning")
	}
	m = key(m, tea.KeyMsg{Type: tea.KeyEnter})
	m = key(m, tea.KeyMsg{Type: tea.KeyDown})
	m = key(m, tea.KeyMsg{Type: tea.KeyEsc})
	m = key(m, tea.KeyMsg{Type: tea.KeyEnter})
	if m.screen != history || m.historyIndex != 0 {
		t.Fatal("fresh history entry kept an old selection")
	}
}

func TestBackKeepsParentContext(t *testing.T) {
	m := statsFixture(t)
	m.screen, m.historyIndex = history, 2
	m = key(m, tea.KeyMsg{Type: tea.KeyEnter})
	m = key(m, tea.KeyMsg{Type: tea.KeyEsc})
	if m.screen != history || m.historyIndex != 2 {
		t.Fatal("detail Back lost its selected history row")
	}
	m.screen, m.settingRow = settings, 7
	question := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}}
	m = key(m, question)
	m = key(m, tea.KeyMsg{Type: tea.KeyEsc})
	if m.screen != settings || m.settingRow != 7 {
		t.Fatal("Help changed the current selection")
	}
	m = key(m, tea.KeyMsg{Type: tea.KeyEsc})
	m.selection = 3
	m = key(m, tea.KeyMsg{Type: tea.KeyEnter})
	if m.settingRow != 0 {
		t.Fatal("fresh entry did not reset after leaving Help")
	}
}

func TestStatisticsScrollStopsAtVisibleEnd(t *testing.T) {
	m := statsFixture(t)
	for range 100 {
		m = key(m, tea.KeyMsg{Type: tea.KeyDown})
	}
	if m.bestOffset != 2 {
		t.Fatal("Statistics scrolled past its last page", m.bestOffset)
	}
	if !strings.Contains(ansi.Strip(m.View()), "3–4 of 4") {
		t.Fatal("wrong last page")
	}
	m = key(m, tea.KeyMsg{Type: tea.KeyUp})
	if m.bestOffset != 1 || !strings.Contains(ansi.Strip(m.View()), "2–3 of 4") {
		t.Fatal("Up did not immediately move the page")
	}
	m.bestOffset = 2
	resized, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 30})
	m = resized.(Model)
	if m.bestOffset != 0 {
		t.Fatal("resize left the page offset past its bounds")
	}
}
