package app

import (
	"fmt"
	"strings"
	"time"

	"github.com/gaurv-m/keywi/internal/animation"
	"github.com/gaurv-m/keywi/internal/theme"
	"github.com/gaurv-m/keywi/internal/ui"
)

func (m Model) playView(w int, p theme.Palette) string {
	line := func(s string) string { return ui.CenterLine(s, w) }
	title := func(s string) string { return line(p.Accent(s)) }
	hint := func(s string) string { return line(p.Muted.Render(s)) }
	var body string
	t := m.tracker()
	if m.mode == "normal" {
		// Center the paragraph as one block, keeping its lines left-aligned.
		centerWidth := w + (m.Width-w)%2
		text := ui.CenterBlock(ui.TypingFrame(t, w, m.Store.Preferences.Cursor, p, m.now, m.Store.Preferences.Animations), centerWidth)
		status := fmt.Sprintf("%.0f", m.casualGame.Remaining(m.now))
		if m.config.Kind == "time" {
			status += "s"
		} else {
			status += " " + plural(int(m.casualGame.Remaining(m.now)), "word") + " left"
		}
		if m.Store.Preferences.LiveWPM {
			gap := "     "
			if w < 40 {
				gap = " · "
			}
			if w < 32 && m.config.Kind == "words" {
				status = strings.TrimSuffix(status, " left")
			}
			status += fmt.Sprintf("%s%.0f wpm", gap, t.Metrics(m.now).WPM)
		}
		body = ui.Join(title("Casual"), "", line(p.Secondary.Render(status)), "", "", text)
		if m.Height < 15 {
			body = ui.Join(title("Casual"), line(p.Secondary.Render(status)), text)
		}
	} else {
		remaining := m.blitzGame.Remaining(m.now).Seconds()
		accent := p.Primary
		if remaining < 15 {
			accent = p.Bonus
		}
		if remaining < 7 {
			accent = p.Error
		}
		status := fmt.Sprintf("%.1fs     %02d %s     level %d", remaining, m.blitzGame.Sentences, plural(m.blitzGame.Sentences, "sentence"), m.blitzGame.Level)
		if w < 44 {
			status = fmt.Sprintf("%.1fs  %02d sent  L%d", remaining, m.blitzGame.Sentences, m.blitzGame.Level)
		}
		barWidth := min(44, w)
		filled := min(barWidth, max(0, int(remaining/30*float64(barWidth))))
		bar := accent.Render(strings.Repeat("━", filled)) + p.Muted.Render(strings.Repeat("─", barWidth-filled))
		completion := m.blitzGame.LastCompletion
		age := m.now.Sub(completion.At)
		stage := ui.BlitzStage(t, completion.Text, age, w, m.Store.Preferences.Cursor, p, m.Store.Preferences.Animations, m.now)
		if !completion.At.IsZero() && age >= 0 && age < animation.BonusDuration && completion.Bonus >= 50*time.Millisecond && (!m.Store.Preferences.Animations || age >= animation.Duration) {
			rows := strings.Split(stage, "\n")
			rows[len(rows)-1] = line(p.Bonus.Render(fmt.Sprintf("+%.1fs", completion.Bonus.Seconds())))
			stage = strings.Join(rows, "\n")
		}
		// Leave room above the typing line for fragments.
		body = ui.Join(title("Blitz"), "", line(accent.Render(status)), "", line(bar), "", "", stage)
		if m.Height == 15 {
			body = ui.Join(title("Blitz"), "", line(accent.Render(status)), "", line(bar), "", stage)
		}
		if m.Height < 15 {
			body = ui.Join(title("Blitz"), line(accent.Render(status)), line(bar), stage)
			if m.Height < 12 {
				body = ui.Join(line(accent.Render(status)), stage)
			}
		}
	}
	if t.Started.IsZero() {
		body += "\n" + hint("type to begin")
	} else {
		body += "\n"
	}
	return body
}
