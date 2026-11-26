package app

import (
	"time"

	"github.com/gaurv-m/keywi/internal/theme"

	tea "github.com/charmbracelet/bubbletea"
)

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch v := msg.(type) {
	case tea.WindowSizeMsg:
		m.Width = max(1, v.Width)
		m.Height = max(1, v.Height)
		m.bestOffset = min(m.bestOffset, m.bestMaxOffset())
		if m.screen == helpScreen {
			m.helpOffset = min(m.helpOffset, m.helpMaxOffset())
		}
		return m, nil
	case tickMsg:
		m.advance(time.Time(v))
		return m, m.tick()
	case tea.KeyMsg:
		if v.String() == "ctrl+c" {
			return m, tea.Quit
		}
		// Ignore pasted input, including embedded controls.
		if v.Paste {
			return m, nil
		}
		now := time.Now()
		wasPlaying := m.screen == play
		m.advance(now)
		if wasPlaying && m.screen != play {
			return m, nil
		}
		key := v.String()
		if m.confirm != "" {
			switch key {
			case "esc":
				m.confirm = ""
			case "left", "right":
				m.confirmYes = !m.confirmYes
			case "enter":
				action := m.confirm
				yes := m.confirmYes
				m.confirm = ""
				if action == "restart" && yes {
					m.start(m.mode)
				} else if yes {
					if action == "quit" {
						return m, tea.Quit
					} else if action == "abandon" {
						m.screen = menu
					} else if action == "reset" {
						m.messageUntil = time.Time{}
						backup, err := m.Store.Reset()
						if err != nil {
							m.message = "Reset failed: " + err.Error()
						} else {
							m.message = "Statistics reset"
							if backup != "" {
								m.message += "; backup: " + backup
							}
						}
					}
				}
			}
			return m, nil
		}
		if key == "?" && m.screen != play {
			if m.screen == helpScreen {
				m.screen = m.helpReturn
			} else {
				m.helpReturn, m.screen, m.helpOffset = m.screen, helpScreen, 0
			}
			return m, nil
		}
		switch m.screen {
		case helpScreen:
			switch key {
			case "esc", "enter":
				m.screen = m.helpReturn
			case "up":
				m.helpOffset = max(0, m.helpOffset-1)
			case "down":
				m.helpOffset = min(m.helpOffset+1, m.helpMaxOffset())
			case "pgup":
				m.helpOffset = max(0, m.helpOffset-max(1, m.Height-9))
			case "pgdown":
				m.helpOffset = min(m.helpOffset+max(1, m.Height-9), m.helpMaxOffset())
			}
		case menu:
			switch key {
			case "left", "up":
				m.selection = (m.selection + 3) % 4
			case "right", "down":
				m.selection = (m.selection + 1) % 4
			case "enter":
				m.message = ""
				switch m.selection {
				case 0:
					m.openSection(configuration)
				case 1:
					m.start("textattack")
				case 2:
					m.openSection(stats)
				case 3:
					m.openSection(settings)
				}
			case "esc":
				m.confirm = "quit"
				m.confirmYes = false
			}
		case configuration:
			switch key {
			case "esc":
				m.screen = menu
			case "up", "down":
				m.configRow = 1 - m.configRow
			case "left", "right":
				if m.configRow == 0 {
					if m.config.Kind == "time" {
						m.config.Kind = "words"
						m.config.Amount = 25
					} else {
						m.config.Kind = "time"
						m.config.Amount = 30
					}
				} else {
					values := []int{15, 30, 60, 120}
					if m.config.Kind == "words" {
						values = []int{10, 25, 50, 100}
					}
					idx := 0
					for i, n := range values {
						if n == m.config.Amount {
							idx = i
						}
					}
					delta := 1
					if key == "left" {
						delta = 3
					}
					m.config.Amount = values[(idx+delta)%4]
				}
				m.Store.Preferences.CasualKind, m.Store.Preferences.CasualAmount = m.config.Kind, m.config.Amount
				m.saveSettings()
			case "enter":
				m.start("normal")
			}
		case play:
			switch key {
			case "esc":
				m.confirm = "abandon"
				m.confirmYes = false
			case "tab":
				m.confirm = "restart"
				m.confirmYes = true
			case "backspace", "ctrl+h":
				if m.mode == "normal" {
					m.casualGame.Backspace(now)
				} else {
					m.blitzGame.Backspace(now)
				}
			default:
				if v.Type == tea.KeyRunes || v.Type == tea.KeySpace {
					runes := v.Runes
					if v.Type == tea.KeySpace {
						runes = []rune{' '}
					}
					completed := 0
					if m.mode == "textattack" {
						completed = m.blitzGame.Sentences
					}
					for _, r := range runes {
						if m.screen != play {
							break
						}
						if m.mode == "normal" {
							m.casualGame.Type(r, now)
							if m.casualGame.Done {
								m.finish()
							}
						} else {
							m.blitzGame.Type(r, now)
							if m.blitzGame.Done {
								m.finish()
							}
						}
					}
					if m.Store.Preferences.Sound && m.mode == "textattack" && m.blitzGame.Sentences > completed {
						return m, func() tea.Msg {
							if m.Audio != nil {
								m.Audio.Play()
							}
							return nil
						}
					}
				}
			}
		case results:
			if m.saveError != nil {
				if key == "enter" {
					m.saveError = m.Store.Add(m.result)
				}
				return m, nil
			}
			switch key {
			case "enter":
				m.start(m.mode)
			case "esc":
				m.screen = menu
			}
		case stats:
			switch key {
			case "left", "right":
				m.statsTab = 1 - m.statsTab
				m.bestOffset = 0
			case "up":
				m.bestOffset = max(0, m.bestOffset-1)
			case "down":
				m.bestOffset = min(m.bestOffset+1, m.bestMaxOffset())
			case "enter":
				m.openSection(history)
			case "esc":
				m.screen = menu
			}
		case history:
			list := m.filtered()
			switch key {
			case "up":
				m.historyIndex = max(0, m.historyIndex-1)
			case "down":
				m.historyIndex = min(max(0, len(list)-1), m.historyIndex+1)
			case "enter":
				if len(list) > 0 {
					m.result = list[m.historyIndex]
					m.screen = detail
				}
			case "esc":
				m.screen = stats
			}
		case detail:
			if key == "esc" || key == "enter" {
				m.screen = history
			}
		case settings:
			switch key {
			case "esc":
				m.screen = menu
			case "up":
				m.settingRow = nextSetting(m.settingRow, -1)
			case "down":
				m.settingRow = nextSetting(m.settingRow, 1)
			case "enter", "left", "right":
				p := &m.Store.Preferences
				switch m.settingRow {
				case 0:
					delta := 1
					if key == "left" {
						delta = -1
					}
					p.Theme = theme.Next(p.Theme, delta)
				case 1:
					styles := []string{"block", "underline", "line"}
					idx := 0
					for i, s := range styles {
						if s == p.Cursor {
							idx = i
						}
					}
					delta := 1
					if key == "left" {
						delta = 2
					}
					p.Cursor = styles[(idx+delta)%3]
				case 2:
					p.Animations = !p.Animations
				case 3:
					p.LiveWPM = !p.LiveWPM
				case 4:
					p.Punctuation = !p.Punctuation
				case 5:
					p.Numbers = !p.Numbers
				case 6:
					p.Capitalization = !p.Capitalization
				case 7:
					p.Sound = !p.Sound
				case 8:
					if key == "enter" {
						m.confirm = "reset"
						m.confirmYes = false
					}
					return m, nil
				}
				m.saveSettings()
			}
		}
	}
	return m, nil
}
