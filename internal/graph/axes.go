package graph

import (
	"fmt"
	"math"
	"sort"
	"strconv"
)

func number(value float64, decimal bool) string {
	if value >= 10000 {
		return fmt.Sprintf("%.3g", value)
	}
	if decimal {
		return fmt.Sprintf("%.1f", value)
	}
	return strconv.FormatFloat(math.Round(value*100)/100, 'f', -1, 64)
}

func niceStep(value float64) float64 {
	scale := math.Pow(10, math.Floor(math.Log10(value)))
	for _, factor := range []float64{1, 2, 2.5, 4, 5, 10} {
		if value <= factor*scale {
			return factor * scale
		}
	}
	return 10 * scale
}

type tick struct {
	position, start int
	label           string
}

// Tick labels reserve their cells before intermediate marks are added.
func xTicks(xmin, xmax float64, cols int, unit string) []tick {
	count := min(5, max(2, cols/10+1))
	step := niceStep((xmax - xmin) / float64(count-1))
	if unit == "s" {
		for _, candidate := range []float64{1, 2, 5, 10, 15, 30, 60, 120, 300, 600, 900, 1800, 3600, 7200, 14400, 28800, 86400} {
			if candidate >= (xmax-xmin)/float64(count-1) {
				step = candidate
				break
			}
		}
	} else {
		step = max(1, math.Ceil(step))
	}
	label := func(value float64, last bool) string {
		text := number(value, false)
		if unit == "s" {
			return text + "s"
		}
		if unit == "run" && last {
			if value == 1 {
				return text + " Run"
			}
			return text + " Runs"
		}
		return text
	}
	makeTick := func(value float64, last bool) tick {
		position := max(0, min(cols-1, int(math.Round((value-xmin)/(xmax-xmin)*float64(cols-1)))))
		text := label(value, last)
		start := max(0, min(cols-len(text), position-len(text)/2))
		return tick{position, start, text}
	}
	var ticks []tick
	occupied := make([]bool, cols)
	add := func(t tick) {
		if t.start+len(t.label) > cols {
			return
		}
		for x := max(0, t.start-1); x < min(cols, t.start+len(t.label)+1); x++ {
			if occupied[x] {
				return
			}
		}
		for x := t.start; x < t.start+len(t.label); x++ {
			occupied[x] = true
		}
		ticks = append(ticks, t)
	}
	add(makeTick(xmin, false))
	add(makeTick(xmax, true))
	first := math.Ceil(xmin/step) * step
	for i := 0; i <= count+2; i++ {
		value := first + float64(i)*step
		if !finite(value) || value >= xmax {
			break
		}
		if value > xmin {
			add(makeTick(value, false))
		}
	}
	sort.Slice(ticks, func(i, j int) bool { return ticks[i].position < ticks[j].position })
	return ticks
}
