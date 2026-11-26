package ui

import (
	"math"
	"strings"

	"github.com/gaurv-m/keywi/internal/theme"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
)

// Logo preserves the original characters and spacing.
func Logo(art string, p theme.Palette) string {
	return logoFrame(art, p, 0, false)
}

// LogoFrame changes shades while keeping the final i fixed.
func LogoFrame(art string, p theme.Palette, frame int) string {
	return logoFrame(art, p, max(0, frame), true)
}

func logoFrame(art string, p theme.Palette, frame int, animate bool) string {
	width := Width(art)
	if width == 0 {
		return art
	}
	lines := strings.Split(art, "\n")
	styles := []lipgloss.Style{p.Primary, p.LogoI}
	if (animate || p.Flame) && len(p.LogoShades) > 0 {
		styles = append(styles, p.LogoShades...)
	} else if len(p.Spectrum) > 0 {
		styles = append(styles, p.Spectrum...)
	}
	shade := func(x, y int) int {
		switch {
		case x >= 42:
			return 1
		case p.Flame && len(p.LogoShades) > 0:
			tick := frame
			if !animate {
				tick = 0
			}
			return 2 + fireShade(x, y, len(lines), tick, len(p.LogoShades))
		case animate && len(p.LogoShades) > 0:
			return 2 + titleShade(x, y, frame, len(p.LogoShades))
		case len(p.Spectrum) > 0:
			return 2 + (x/5)%len(p.Spectrum)
		default:
			return 0
		}
	}
	// Cache ANSI wrappers for each shade.
	type paint struct {
		prefix, suffix string
		ready          bool
	}
	paints := make([]paint, len(styles))

	for y, line := range lines {
		var out, run strings.Builder
		active, column := shade(0, y), 0
		flush := func() {
			if run.Len() > 0 {
				if !paints[active].ready {
					prefix, suffix, _ := strings.Cut(styles[active].Render("x"), "x")
					paints[active] = paint{prefix, suffix, true}
				}
				out.WriteString(paints[active].prefix)
				out.WriteString(run.String())
				out.WriteString(paints[active].suffix)
				run.Reset()
			}
		}
		for _, r := range line {
			// Empty cells need no shade transition. Keep the i's region fixed.
			if r == ' ' && column < 42 {
				run.WriteRune(r)
				column++
				continue
			}
			next := shade(column, y)
			if next != active {
				flush()
				active = next
			}
			run.WriteRune(r)
			column += ansi.StringWidth(string(r))
		}
		flush()
		lines[y] = out.String()
	}
	return strings.Join(lines, "\n")
}

// Noise changes scattered shades without moving glyphs.
func titleShade(x, y, frame, count int) int {
	x0, y0 := x/6, y/3
	u, v := soften(float64(x%6)/6), soften(float64(y%3)/3)
	top := mix(titleNoise(x0, y0, frame), titleNoise(x0+1, y0, frame), u)
	bottom := mix(titleNoise(x0, y0+1, frame), titleNoise(x0+1, y0+1, frame), u)
	return int(math.Round(mix(top, bottom, v) * float64(max(0, count-1))))
}

const titleChangeFrames = 6 // A new patch colour every 0.6 s at the same 10 Hz.

func titleNoise(x, y, frame int) float64 {
	local := frame + int(titleHash(x, y, 0)%titleChangeFrames)
	epoch := local/titleChangeFrames + 1
	weight := soften(float64(local%titleChangeFrames) / titleChangeFrames)
	a := float64(titleHash(x, y, epoch)&65535) / 65535
	b := float64(titleHash(x, y, epoch+1)&65535) / 65535
	return mix(a, b, weight)
}

func titleHash(x, y, epoch int) uint32 {
	n := uint32(x)*0x9e3779b9 ^ uint32(y)*0x85ebca6b ^ uint32(epoch)*0xc2b2ae35
	n ^= n >> 16
	n *= 0x7feb352d
	n ^= n >> 15
	n *= 0x846ca68b
	return n ^ (n >> 16)
}

func soften(t float64) float64         { return t * t * (3 - 2*t) }
func mix(a, b, weight float64) float64 { return a + (b-a)*weight }
