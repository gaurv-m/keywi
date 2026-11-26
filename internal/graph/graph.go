// Package graph draws terminal charts.
package graph

import (
	"fmt"
	"math"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

type Point struct{ X, Y float64 }

type Options struct {
	XMin, XMax float64
	XUnit      string
	Average    *float64
}

type Styles struct {
	Line, Labels, Grid, Average lipgloss.Style
}

func Plot(points []Point, width, height int, xlabel string) string {
	return Draw(points, width, height, Options{XUnit: xlabel}, Styles{})
}

func StyledPlot(points []Point, width, height int, xlabel string, line, axis lipgloss.Style) string {
	return Draw(points, width, height, Options{XUnit: xlabel}, Styles{Line: line, Labels: axis.Faint(false), Grid: axis, Average: axis.Faint(false)})
}

// Draw keeps zero visible and labels the reference without covering the data.
func Draw(points []Point, width, height int, options Options, style Styles) string {
	if width < 18 || height < 4 {
		return "Graph needs more space"
	}
	var data []Point
	xmin, xmax, ymax := options.XMin, options.XMax, 0.0
	if !finite(xmin) || xmin < 0 {
		xmin = 0
	}
	if !finite(xmax) {
		xmax = 0
	}
	for _, p := range points {
		if !finite(p.X) || !finite(p.Y) || p.X < xmin {
			continue
		}
		p.Y = max(0, p.Y)
		data = append(data, p)
		xmax = max(xmax, p.X)
		ymax = max(ymax, p.Y)
	}
	if len(data) == 0 {
		return "No recorded samples yet"
	}
	if xmax <= xmin {
		xmax = xmin + 1
	}
	average := ""
	reference := 0.0
	if options.Average != nil && finite(*options.Average) && *options.Average >= 0 {
		reference = *options.Average
		ymax = max(ymax, reference)
		average = "Avg " + number(reference, true)
	}
	rows := height - 2
	step := max(1, niceStep(max(1, ymax)/float64(min(3, rows-1))))
	ymax = math.Ceil(max(1, ymax)/step) * step
	labelWidth := max(3, len(number(ymax, false)))
	prefix := labelWidth + 2
	referenceWidth := 0
	if average != "" {
		referenceWidth = len(average) + 1
	}
	cols := width - prefix - referenceWidth
	if cols < 8 {
		average = ""
		referenceWidth = 0
		cols = width - prefix
	}
	grid := raster(data, cols, rows, xmin, xmax, ymax)
	ypixel := func(value float64) int { return rows*4 - 1 - int(math.Round(value/ymax*float64(rows*4-1))) }
	labels := make(map[int]string)
	guides := make(map[int]bool)
	for v := 0.0; v <= ymax+step/100; v += step {
		row := max(0, min(rows-1, ypixel(v)/4))
		labels[row] = number(v, false)
		if v > 0 && v < ymax {
			guides[row] = true
		}
	}
	referenceRow := -1
	if average != "" {
		referenceRow = max(0, min(rows-1, ypixel(reference)/4))
	}
	var lines []string
	for y, cells := range grid {
		var trace strings.Builder
		for _, cell := range cells {
			switch {
			case cell != '\u2800':
				trace.WriteString(style.Line.Render(string(cell)))
			case y == referenceRow:
				trace.WriteString(style.Grid.Render("╌"))
			case guides[y]:
				trace.WriteString(style.Grid.Render("┄"))
			default:
				trace.WriteByte(' ')
			}
		}
		suffix := strings.Repeat(" ", referenceWidth)
		if y == referenceRow {
			suffix = " " + style.Average.Render(average)
		}
		lines = append(lines, style.Labels.Render(fmt.Sprintf("%*s", labelWidth, labels[y]))+style.Grid.Render(" │")+trace.String()+suffix)
	}
	ticks := xTicks(xmin, xmax, cols, options.XUnit)
	axis := []rune(strings.Repeat("─", cols))
	for _, tick := range ticks {
		axis[tick.position] = '┬'
	}
	lines = append(lines, style.Grid.Render(strings.Repeat(" ", labelWidth+1)+"└"+string(axis)))
	line := []rune(strings.Repeat(" ", cols))
	for _, tick := range ticks {
		copy(line[tick.start:], []rune(tick.label))
	}
	lines = append(lines, strings.Repeat(" ", prefix)+style.Labels.Render(string(line)))
	return strings.Join(lines, "\n")
}

func finite(value float64) bool { return !math.IsNaN(value) && !math.IsInf(value, 0) }
