package app

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/ansi"
)

func TestMistakeLimitStaysQuietAcrossSentencesAndRestarts(t *testing.T) {
	for _, mode := range []string{"normal", "textattack"} {
		m := model(t)
		m.start(mode)
		for episode := 1; episode <= 3; episode++ {
			m = key(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("~~")})
			m = key(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("more blocked input")})
			view := ansi.Strip(m.View())
			if strings.Contains(view, "backspace to fix") {
				t.Fatal("unexpected correction hint", mode, episode)
			}
			if !m.tracker().Blocked() || len(m.tracker().Input) != 2 {
				t.Fatal("hint changes broke the mistake limit", mode, episode)
			}
			m = key(m, tea.KeyMsg{Type: tea.KeyBackspace})
			m = key(m, tea.KeyMsg{Type: tea.KeyBackspace})
		}
		if mode == "textattack" {
			target := append([]rune(nil), m.tracker().Target...)
			m = key(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: target})
			if m.blitzGame.Sentences != 1 {
				t.Fatal("sentence completion reset warnings")
			}
			m = key(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("~~")})
			if strings.Contains(ansi.Strip(m.View()), "backspace to fix") {
				t.Fatal("warning returned on a new sentence")
			}
		}
		m = key(m, tea.KeyMsg{Type: tea.KeyTab})
		m = key(m, tea.KeyMsg{Type: tea.KeyEnter})
		if len(m.tracker().Input) != 0 {
			t.Fatal("restart did not reset warnings", mode)
		}
		m = key(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("~~")})
		if strings.Contains(ansi.Strip(m.View()), "backspace to fix") {
			t.Fatal("new run showed correction text", mode)
		}
	}
}
