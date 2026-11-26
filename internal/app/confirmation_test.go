package app

import (
	"os"
	"strings"
	"testing"

	"github.com/gaurv-m/keywi/internal/theme"
	"github.com/gaurv-m/keywi/internal/ui"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	"github.com/muesli/termenv"
)

func TestConfirmationBracketsAreStableAcrossThemes(t *testing.T) {
	t.Setenv("TERM", "xterm-256color")
	t.Setenv("NO_COLOR", "")
	os.Unsetenv("NO_COLOR")
	renderer := lipgloss.DefaultRenderer()
	previous := renderer.ColorProfile()
	renderer.SetColorProfile(termenv.TrueColor)
	t.Cleanup(func() { renderer.SetColorProfile(previous) })
	pairs := [][2]string{{"Continue", "Exit"}, {"Keep", "Abandon"}, {"Cancel", "Reset"}, {"Cancel", "Restart"}}
	for _, name := range theme.Names() {
		p := theme.New(name)
		for _, pair := range pairs {
			for _, width := range []int{20, 36, 76} {
				no := confirmationChoices(pair[0], pair[1], false, width, p)
				yes := confirmationChoices(pair[0], pair[1], true, width, p)
				left, right := ansi.Strip(no), ansi.Strip(yes)
				if ui.Width(left) > width || ui.Width(right) > width || ui.Width(left) != ui.Width(right) {
					t.Fatal("confirmation overflowed or resized", name, pair, width)
				}
				for _, text := range []string{left, right} {
					if strings.Count(text, "[") != 1 || strings.Count(text, "]") != 1 {
						t.Fatal("selection needs exactly one bracket pair", text)
					}
				}
				for _, label := range pair {
					if strings.Index(left, label) != strings.Index(right, label) {
						t.Fatal("confirmation label moved", name, pair)
					}
					if !strings.Contains(no, p.Correct.Render(label)) || !strings.Contains(yes, p.Correct.Render(label)) {
						t.Fatal("selection changed label colors", name, pair)
					}
				}
				if !strings.Contains(left, "[ "+pair[0]+" ]") && !strings.Contains(left, "["+pair[0]+"]") {
					t.Fatal("brackets do not mark the actual default")
				}
				if !strings.Contains(right, "[ "+pair[1]+" ]") && !strings.Contains(right, "["+pair[1]+"]") {
					t.Fatal("brackets do not mark the selected action")
				}
			}
		}
	}
}

func TestGameConfirmationsFollowBracketedChoice(t *testing.T) {
	for _, mode := range []string{"normal", "textattack"} {
		m := model(t)
		m.start(mode)
		m = key(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{m.tracker().Target[0]}})
		count := len(m.tracker().Input)
		m = key(m, tea.KeyMsg{Type: tea.KeyEsc})
		if m.confirmYes || !strings.Contains(ansi.Strip(m.View()), "[ Keep ]") {
			t.Fatal("abandon did not default to Keep", mode)
		}
		m = key(m, tea.KeyMsg{Type: tea.KeyRight})
		if !strings.Contains(ansi.Strip(m.View()), "[ Abandon ]") {
			t.Fatal("brackets did not follow selection", mode)
		}
		m = key(m, tea.KeyMsg{Type: tea.KeyEsc})
		if m.confirm != "" || len(m.tracker().Input) != count {
			t.Fatal("cancel changed the run", mode)
		}
		m = key(m, tea.KeyMsg{Type: tea.KeyTab})
		if !m.confirmYes || !strings.Contains(ansi.Strip(m.View()), "[ Restart ]") {
			t.Fatal("restart default is unclear", mode)
		}
		m = key(m, tea.KeyMsg{Type: tea.KeyLeft})
		if !strings.Contains(ansi.Strip(m.View()), "[ Cancel ]") {
			t.Fatal("restart Cancel is unclear", mode)
		}
		m = key(m, tea.KeyMsg{Type: tea.KeyEnter})
		if len(m.tracker().Input) != count || m.screen != play {
			t.Fatal("Cancel still restarted the run", mode)
		}
		m = key(m, tea.KeyMsg{Type: tea.KeyTab})
		m = key(m, tea.KeyMsg{Type: tea.KeyEnter})
		if len(m.tracker().Input) != 0 || !m.tracker().Started.IsZero() {
			t.Fatal("bracketed Restart did not restart", mode)
		}
	}
}
