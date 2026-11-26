package ui

import (
	"strings"
	"time"

	"github.com/gaurv-m/keywi/internal/animation"
	"github.com/gaurv-m/keywi/internal/theme"
	"github.com/gaurv-m/keywi/internal/typing"
)

// BlitzStage reserves five rows for typing and fragments.
func BlitzStage(current *typing.Tracker, previous string, age time.Duration, width int, cursor string, p theme.Palette, animate bool, now ...time.Time) string {
	width = max(1, width)
	clock := time.Time{}
	if len(now) > 0 {
		clock = now[0]
	}
	raw := strings.Split(TypingFrame(current, width, cursor, p, clock, animate), "\n")
	blockWidth := 0
	for _, line := range raw {
		blockWidth = max(blockWidth, Width(line))
	}
	origin := max(0, (width-blockWidth)/2)
	blank := strings.Repeat(" ", width)
	rows := []string{blank, blank, blank, blank, blank}
	lengths := make([]int, 3)
	for y, line := range raw {
		lengths[y] = Width(line)
		rows[y+1] = strings.Repeat(" ", origin) + line + strings.Repeat(" ", max(0, width-origin-lengths[y]))
	}
	if !animate || previous == "" || age < 0 || age >= animation.Duration {
		return strings.Join(rows, "\n")
	}
	old := typing.New(previous)
	old.Input = append([]rune(nil), old.Target...)
	oldText := CenterBlock(Typing(&old, width, cursor, p), width)
	source := append([]string{blank}, strings.Split(Plain(oldText), "\n")...)
	source = append(source, blank)
	frame := animation.ShatterRows(source, age, width, len(rows))
	// Compose fragments once to avoid duplicating ANSI sequences.
	cells := make([][]string, len(rows))
	for y := range cells {
		cells[y] = make([]string, width)
		for x := range cells[y] {
			cells[y][x] = " "
		}
	}
	for _, particle := range frame.Particles {
		x, y := particle.X, particle.Y
		if y >= 1 && y <= 3 && lengths[y-1] > 0 && x < origin+lengths[y-1]+1 && x+1 > origin-1 {
			continue
		}
		cells[y][x] = particle.Glyph
	}
	render := func(fragment []string) string {
		plain := strings.Join(fragment, "")
		if strings.TrimSpace(plain) == "" {
			return plain
		}
		return p.Correct.Faint(frame.Faint).Render(plain)
	}
	for y := range rows {
		if y >= 1 && y <= 3 && lengths[y-1] > 0 {
			start := min(origin, len(cells[y]))
			end := min(origin+lengths[y-1], len(cells[y]))
			rows[y] = render(cells[y][:start]) + raw[y-1] + render(cells[y][end:])
		} else {
			rows[y] = render(cells[y])
		}
	}
	return strings.Join(rows, "\n")
}
