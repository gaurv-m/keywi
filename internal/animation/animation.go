// Package animation renders sentence fragments.
package animation

import (
	"math"
	"strings"
	"time"
	"unicode"

	"github.com/charmbracelet/x/ansi"
)

const Duration = 220 * time.Millisecond
const BonusDuration = 650 * time.Millisecond

type Particle struct {
	X, Y  int
	Glyph string
}

type Frame struct {
	Particles []Particle
	Rows      []string
	Faint     bool
}

// Shatter scatters fragments along fixed paths.
func Shatter(text string, age time.Duration, width, height int) Frame {
	if width < 1 || height < 1 {
		return Frame{}
	}
	lines := strings.Split(ansi.Wordwrap(text, width, ""), "\n")
	if len(lines) > height {
		lines = lines[len(lines)-height:]
	}
	blockWidth := 0
	for _, line := range lines {
		blockWidth = max(blockWidth, ansi.StringWidth(line))
	}
	origin := max(0, (width-blockWidth)/2)
	for i := range lines {
		lines[i] = strings.Repeat(" ", origin) + lines[i]
	}
	return ShatterRows(lines, age, width, height)
}

// ShatterRows preserves each row's original position.
func ShatterRows(lines []string, age time.Duration, width, height int) Frame {
	if width < 1 || height < 1 {
		return Frame{}
	}
	rows := make([][]string, height)
	occupied := make([][]bool, height)
	for y := range rows {
		rows[y] = make([]string, width)
		occupied[y] = make([]bool, width)
		for x := range rows[y] {
			rows[y][x] = " "
		}
	}
	f := Frame{Rows: make([]string, height), Faint: age >= Duration/2}
	// A fixed blank canvas also clears every old fragment after the effect expires.
	render := func() Frame {
		for y := range rows {
			f.Rows[y] = strings.Join(rows[y], "")
		}
		return f
	}
	if age < 0 || age >= Duration {
		return render()
	}
	progress := float64(age) / float64(Duration)
	ease := 1 - (1-progress)*(1-progress)
	shards := []rune{'▪', '▒', '░', '⠂', '⠄', '·'}
	index := 0
	for row, line := range lines {
		if row >= height {
			break
		}
		column := 0
		for _, r := range line {
			rw := ansi.StringWidth(string(r))
			sourceX := column
			column += rw
			index++
			if unicode.IsSpace(r) || rw == 0 {
				continue
			}
			seed := uint32(index) * 2654435761
			lifetime := 0.55 + 0.45*float64(seed%101)/100
			if progress >= lifetime {
				continue
			}
			direction := -1
			if sourceX >= width/2 {
				direction = 1
			}
			distance := float64(1 + (seed>>8)%3)
			x := sourceX + direction*int(math.Round(ease*distance))
			// Fragments scatter above and below their original character baseline.
			vertical := int((seed>>16)%3) - 1
			y := row + int(math.Round(ease*float64(vertical)*2))
			// Break at the original character cell.
			glyph := string(shards[(seed>>4)%uint32(len(shards))])
			if progress > 0.65 {
				glyph = "·"
			}
			cells := ansi.StringWidth(glyph)
			if y < 0 || y >= height || x < 0 || x+cells > width {
				continue
			}
			blocked := false
			for c := x; c < x+cells; c++ {
				if occupied[y][c] {
					blocked = true
				}
			}
			if blocked {
				continue
			}
			f.Particles = append(f.Particles, Particle{X: x, Y: y, Glyph: glyph})
			rows[y][x] = glyph
			for c := x; c < x+cells; c++ {
				occupied[y][c] = true
				if c > x {
					rows[y][c] = ""
				}
			}
		}
	}
	return render()
}
