// Package app runs the terminal interface.
package app

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/gaurv-m/keywi/assets"
	"github.com/gaurv-m/keywi/internal/blitz"
	"github.com/gaurv-m/keywi/internal/casual"
	"github.com/gaurv-m/keywi/internal/sound"
	"github.com/gaurv-m/keywi/internal/statistics"
	"github.com/gaurv-m/keywi/internal/storage"
	"github.com/gaurv-m/keywi/internal/typing"

	tea "github.com/charmbracelet/bubbletea"
)

type screen int

const (
	menu screen = iota
	configuration
	play
	results
	stats
	history
	detail
	settings
	helpScreen
)

type tickMsg time.Time

func (m Model) tick() tea.Cmd {
	interval := 33 * time.Millisecond
	if m.screen != play {
		interval = 100 * time.Millisecond
	}
	return tea.Tick(interval, func(t time.Time) tea.Msg { return tickMsg(t) })
}

type Model struct {
	Store         *storage.Store
	Audio         *sound.Player
	Width, Height int

	screen       screen
	selection    int
	configRow    int
	settingRow   int
	statsTab     int
	historyIndex int
	bestOffset   int
	helpReturn   screen
	helpOffset   int

	config        casual.Config
	mode          string
	casualGame    *casual.Game
	blitzGame     *blitz.Game
	paragraphDeck *assets.Deck
	blitzDeck     *assets.Deck

	now          time.Time
	result       statistics.Result
	personalBest bool
	confirm      string
	confirmYes   bool
	message      string
	messageUntil time.Time
	saveError    error
	warning      string
}

func New(s *storage.Store, warnings []string) Model {
	config := casual.DefaultConfig()
	config.Kind, config.Amount = s.Preferences.CasualKind, s.Preferences.CasualAmount
	if !config.Valid() {
		config = casual.DefaultConfig()
	}
	return Model{
		Store: s, Width: 80, Height: 24, config: config,
		now: time.Now(), warning: strings.Join(warnings, "; "),
		paragraphDeck: assets.NewParagraphDeck(), blitzDeck: assets.NewDeck(),
	}
}
func (m Model) Init() tea.Cmd { return m.tick() }
func (m *Model) start(mode string) {
	m.mode = mode
	m.screen = play
	m.confirm = ""
	m.message = ""
	m.saveError = nil
	m.now = time.Now()
	if mode == "normal" {
		m.config.Punctuation = m.Store.Preferences.Punctuation
		m.config.Numbers = m.Store.Preferences.Numbers
		m.config.Capitalization = m.Store.Preferences.Capitalization
		m.casualGame = casual.NewWithDeck(m.config, m.paragraphDeck)
		m.config = m.casualGame.Config
	} else {
		m.blitzGame = blitz.NewWithDeck(m.blitzDeck)
	}
}
func (m *Model) tracker() *typing.Tracker {
	if m.mode == "normal" {
		return &m.casualGame.Tracker
	}
	return &m.blitzGame.Tracker
}
func (m *Model) finish() {
	if m.screen != play {
		return
	}
	m.result = statistics.Result{ID: fmt.Sprintf("%d-%d", time.Now().UnixNano(), os.Getpid()), At: time.Now().UTC(), Mode: m.mode, Metrics: m.tracker().Metrics(m.now)}
	if m.mode == "normal" {
		m.result.Config = m.config
	} else {
		m.result.Sentences = m.blitzGame.Sentences
		m.result.Level = m.blitzGame.Level
	}
	m.personalBest = statistics.PersonalBest(m.Store.Data.Results, m.result)
	m.saveError = m.Store.Add(m.result)
	m.screen = results
	m.confirm = ""
}
func (m *Model) advance(now time.Time) {
	if now.Before(m.now) {
		now = m.now
	}
	m.now = now
	if !m.messageUntil.IsZero() && !now.Before(m.messageUntil) {
		m.message = ""
		m.messageUntil = time.Time{}
	}
	if m.screen != play {
		return
	}
	if m.mode == "normal" {
		m.casualGame.Advance(now)
		if m.casualGame.Done {
			m.finish()
		}
	} else {
		m.blitzGame.Advance(now)
		if m.blitzGame.Done {
			m.finish()
		}
	}
}
func (m *Model) saveSettings() {
	m.messageUntil = time.Time{}
	if err := m.Store.SavePreferences(); err != nil {
		m.message = "Preferences not saved: " + err.Error()
	} else {
		m.message = "Preferences saved"
		m.messageUntil = m.now.Add(2 * time.Second)
	}
}
func (m *Model) filtered() []statistics.Result {
	mode := "normal"
	if m.statsTab == 1 {
		mode = "textattack"
	}
	var out []statistics.Result
	for i := len(m.Store.Data.Results) - 1; i >= 0; i-- {
		if m.Store.Data.Results[i].Mode == mode {
			out = append(out, m.Store.Data.Results[i])
		}
	}
	return out
}

// PersistenceError reports an unsaved result.
func (m Model) PersistenceError() error { return m.saveError }
