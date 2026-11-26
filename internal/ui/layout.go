// Package ui lays out terminal text.
package ui

import (
	"strings"
	"time"

	"github.com/gaurv-m/keywi/internal/theme"
	"github.com/gaurv-m/keywi/internal/typing"

	"github.com/charmbracelet/x/ansi"
)

func Width(s string) int {
	w := 0
	for _, line := range strings.Split(s, "\n") {
		w = max(w, ansi.StringWidth(line))
	}
	return w
}
func ContentWidth(w int) int { return max(1, min(76, w-4)) }

// CenterBlock centers a rectangle without changing its spacing.
func CenterBlock(block string, width int) string {
	lines := strings.Split(block, "\n")
	bw := 0
	for _, l := range lines {
		bw = max(bw, Width(l))
	}
	pad := strings.Repeat(" ", max(0, (width-bw)/2))
	for i, l := range lines {
		lines[i] = pad + l + strings.Repeat(" ", max(0, bw-Width(l)))
	}
	return strings.Join(lines, "\n")
}
func CenterLine(s string, w int) string { return strings.Repeat(" ", max(0, (w-Width(s))/2)) + s }
func Place(block string, width, height int) string {
	if width < 1 || height < 1 {
		return ""
	}
	lines := strings.Split(block, "\n")
	if len(lines) > height {
		lines = lines[:height]
	}
	for i, l := range lines {
		lines[i] = ansi.Truncate(l, width, "")
	}
	top := max(0, (height-len(lines))/2)
	return strings.Repeat("\n", top) + strings.Join(lines, "\n") + strings.Repeat("\n", max(0, height-top-len(lines)))
}

type Line struct{ Start, End int }

// Wrap returns rune offsets, keeping words together when they fit.
func Wrap(text []rune, width int) []Line {
	width = max(1, width)
	var lines []Line
	for start := 0; start < len(text); {
		end, cells, lastSpace := start, 0, -1
		for end < len(text) {
			rw := Width(string(text[end]))
			if cells+rw > width && end > start {
				break
			}
			if cells+rw > width {
				end++
				break
			}
			cells += rw
			if text[end] == ' ' {
				lastSpace = end
			}
			end++
		}
		if end < len(text) && lastSpace > start {
			end = lastSpace + 1
		}
		if end == start {
			end++
		}
		lines = append(lines, Line{start, end})
		start = end
	}
	if len(lines) == 0 {
		lines = []Line{{0, 0}}
	}
	return lines
}

// Typing renders three rows around the cursor.
func Typing(t *typing.Tracker, width int, cursor string, p theme.Palette) string {
	return TypingFrame(t, width, cursor, p, time.Time{}, false)
}

const CursorPulseDuration = 180 * time.Millisecond

// TypingFrame pulses a blocked cursor in place.
func TypingFrame(t *typing.Tracker, width int, cursor string, p theme.Palette, now time.Time, animate bool) string {
	lines := Wrap(t.Target, width)
	pos := len(t.Input)
	active := 0
	for i, l := range lines {
		if pos >= l.Start {
			active = i
		}
	}
	locked := t.CorrectionRequired()
	cursorPos := pos
	dimPulse := false
	if locked && len(t.Target) > 0 {
		cursorPos = min(cursorPos, len(t.Target)-1)
		age := now.Sub(t.BlockedAt)
		if animate && !t.BlockedAt.IsZero() && age >= 0 && age < CursorPulseDuration {
			dimPulse = age >= 45*time.Millisecond && age < 105*time.Millisecond
		}
	}
	first := max(0, active-1)
	first = min(first, max(0, len(lines)-3))
	end := min(len(lines), first+3)
	out := make([]string, 0, 3)
	for _, l := range lines[first:end] {
		var s strings.Builder
		for i := l.Start; i < l.End; i++ {
			r := t.Target[i]
			display := string(r)
			st := p.Muted
			if i < pos {
				if t.Input[i] == r {
					st = p.Correct
				} else {
					st = p.Error
					if r == ' ' {
						display = "·"
					}
				}
			}
			if i == cursorPos {
				st = p.Primary
				if locked {
					st = p.Error
					if dimPulse {
						st = st.Foreground(p.Muted.GetForeground()).Faint(true)
					}
				}
				switch cursor {
				case "underline":
					st = st.Underline(true)
				case "line":
					display += "\u20d2"
					st = st.Bold(true)
				default:
					st = st.Reverse(true)
				}
			}
			if i == cursorPos {
				// Preserve cursor styling on spaces.
				prefix, suffix, _ := strings.Cut(st.Render("x"), "x")
				s.WriteString(prefix)
				s.WriteString(display)
				s.WriteString(suffix)
			} else {
				s.WriteString(st.Render(display))
			}
		}
		out = append(out, s.String())
	}
	for len(out) < 3 {
		out = append(out, "")
	}
	return strings.Join(out, "\n")
}
func Plain(s string) string       { return ansi.Strip(s) }
func Join(lines ...string) string { return strings.Join(lines, "\n") }

// MenuBody balances the gaps around the menu.
func MenuBody(title, menu string, height int) string {
	titleRows, menuRows := strings.Split(title, "\n"), strings.Split(menu, "\n")
	top := max(0, (height-len(titleRows)-len(menuRows)-1)/2)
	gap := max(1, (height+1-top-len(titleRows)-len(menuRows))/2)
	menuTop := top + len(titleRows) + gap
	rows := make([]string, max(height, menuTop+len(menuRows)))
	copy(rows[top:], titleRows)
	copy(rows[menuTop:], menuRows)
	return strings.Join(rows, "\n")
}

// Page anchors navigation above the bottom margin.
func Page(body, footer string, width, height int) string {
	if height < 4 {
		return Place(body, width, height)
	}
	footerLines := strings.Split(footer, "\n")
	for i, line := range footerLines {
		footerLines[i] = ansi.Truncate(line, width, "")
	}
	return Place(body, width, max(1, height-len(footerLines)-2)) + "\n\n" + strings.Join(footerLines, "\n") + "\n"
}
