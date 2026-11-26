package graph

import (
	"math"
	"reflect"
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
)

func TestSmoothWeightsIntervalsAndKeepsTheSource(t *testing.T) {
	points := []Point{{1, 0}, {4, 100}, {5, 0}}
	before := append([]Point(nil), points...)
	result := Smooth(points, 4)
	if result[1].Y != 75 || result[2].Y != 75 {
		t.Fatal("irregular intervals lost their duration weights", result)
	}
	if !reflect.DeepEqual(points, before) {
		t.Fatal("smoothing changed recorded data")
	}
	cutoff := Smooth(points, 1)
	if cutoff[2].Y != 0 {
		t.Fatal("a pause was removed from the curve", cutoff)
	}
}

func TestSmoothLimitsSpikesAndRejectsInvalidSamples(t *testing.T) {
	points := []Point{{1, 0}, {2, 100}, {3, 0}, {4, 100}, {5, 0}}
	result := Smooth(points, 5)
	if result[1].Y != 50 || result[4].Y != 40 {
		t.Fatal("rolling window did not soften the spikes", result)
	}
	invalid := Smooth([]Point{{math.NaN(), 1}, {1, math.Inf(1)}, {2, 20}, {2, 999}, {1, 999}, {3, 30}}, 3)
	if len(invalid) != 2 || invalid[1].X != 3 || !finite(invalid[1].Y) {
		t.Fatal("invalid or unordered intervals entered the chart", invalid)
	}
}

func TestChartHasReadableTimeMarksAndAverage(t *testing.T) {
	average := 52.1
	plot := Draw([]Point{{0, 0}, {120, 110}}, 76, 10, Options{XUnit: "s", Average: &average}, Styles{})
	for _, label := range []string{"120 │", "80 │", "40 │", "0 │", "0s", "30s", "60s", "90s", "120s", "Avg 52.1"} {
		if !strings.Contains(plot, label) {
			t.Fatal("missing chart label", label, plot)
		}
	}
	if !strings.Contains(plot, "╌") {
		t.Fatal("average reference line missing")
	}
}

func TestChartsFitNarrowWindowsAndFlatData(t *testing.T) {
	for _, average := range []float64{0, 52.1, 1000} {
		for _, width := range []int{18, 20, 24, 32, 40, 60, 76} {
			for height := 4; height <= 12; height++ {
				result := Draw([]Point{{0, average}, {120, average}}, width, height, Options{XUnit: "s", Average: &average}, Styles{})
				rows := strings.Split(result, "\n")
				if len(rows) != height {
					t.Fatal("chart height changed", width, height)
				}
				for _, row := range rows {
					if ansi.StringWidth(row) > width {
						t.Fatal("labels overflowed the chart", width, height, row)
					}
				}
				if strings.Contains(result, "NaN") || strings.Contains(result, "Inf") {
					t.Fatal("invalid flat-data scale")
				}
				if width >= 32 && !strings.Contains(result, "Avg ") {
					t.Fatal("average label disappeared", width, result)
				}
			}
		}
	}
}

func TestRecentRunsStartAtOneAndPreserveEveryPoint(t *testing.T) {
	points := []Point{{1, 40}, {2, 80}, {3, 20}, {4, 60}, {5, 100}}
	before := append([]Point(nil), points...)
	average := 60.0
	result := Draw(points, 60, 8, Options{XMin: 1, XUnit: "run", Average: &average}, Styles{})
	axis := strings.Split(result, "\n")[7]
	if !strings.HasPrefix(strings.TrimSpace(axis), "1") || !strings.HasSuffix(axis, "5 Runs") {
		t.Fatal("run axis lost its bounds", axis)
	}
	if !reflect.DeepEqual(before, points) {
		t.Fatal("plotting changed run scores")
	}
}
