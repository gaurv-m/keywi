package app

import (
	"strings"
	"testing"
	"time"

	"github.com/gaurv-m/keywi/internal/typing"

	"github.com/charmbracelet/x/ansi"
)

func TestCasualCentersTheParagraphBlockWithSharedLeftEdge(t *testing.T) {
	for _, width := range []int{24, 40, 60, 80, 100, 101, 120} {
		for _, target := range []string{
			"A short passage.",
			"A short passage opens the page. Another sentence continues through the garden and ends beside the station.",
		} {
			m := model(t)
			m.Width, m.Height = width, 30
			m.start("normal")
			m.casualGame.Tracker = typing.New(target)
			origin := -1
			for _, blocked := range []bool{false, true} {
				if blocked {
					m.casualGame.Tracker.Type('~', m.now)
					m.casualGame.Tracker.Type('~', m.now)
					m.now = m.now.Add(80 * time.Millisecond)
				}
				rows := strings.Split(ansi.Strip(m.View()), "\n")
				start := -1
				for y, row := range rows {
					if strings.Contains(row, "passage") {
						start = y
						break
					}
				}
				if start < 0 {
					t.Fatalf("paragraph disappeared: width=%d blocked=%v target=%q view=%q", width, blocked, target, strings.Join(rows, "\n"))
				}
				left, end := -1, 0
				for _, row := range rows[start : start+3] {
					if strings.TrimSpace(row) == "" {
						continue
					}
					x := len(row) - len(strings.TrimLeft(row, " "))
					if left >= 0 && x != left {
						t.Fatal("lines were centered separately", width)
					}
					left = x
					end = max(end, ansi.StringWidth(strings.TrimRight(row, " ")))
				}
				// A wrap separator and odd terminal width may each use one cell.
				difference := left - (width - end)
				if difference < -2 || difference > 2 {
					t.Fatal("paragraph block is off center", width, left, width-end)
				}
				if blocked && left != origin {
					t.Fatal("cursor feedback shifted the paragraph", width)
				}
				origin = left
				if !strings.Contains(rows[m.Height-2], "esc") {
					t.Fatal("paragraph layout displaced navigation")
				}
			}
		}
	}
}
