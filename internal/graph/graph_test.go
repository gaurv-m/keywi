package graph

import (
	"math"
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
)

func TestScalingAndBraille(t *testing.T) {
	s := Plot([]Point{{0, 0}, {1, 100}}, 30, 6, "s")
	lines := strings.Split(s, "\n")
	if len(lines) != 6 {
		t.Fatalf("height %d", len(lines))
	}
	if !strings.Contains(lines[0], "120 │") || !strings.Contains(lines[3], "0 │") || !strings.HasSuffix(lines[5], "1s") {
		t.Fatal(s)
	}
	var dots int
	for _, r := range s {
		if r > '\u2800' && r <= '\u28ff' {
			dots++
		}
	}
	if dots < 10 {
		t.Fatal("no real rasterization")
	}
	for _, l := range lines {
		if ansi.StringWidth(l) > 30 {
			t.Fatal("clipped", l)
		}
	}
}
func TestEndpointsAndInvalid(t *testing.T) {
	s := Plot([]Point{{0, 0}, {1, 100}}, 18, 4, "s")
	lines := strings.Split(s, "\n")
	first := []rune(strings.SplitN(lines[0], "│", 2)[1])
	last := []rune(strings.SplitN(lines[1], "│", 2)[1])
	if first[len(first)-1]&0xff == 0 || last[0]&0xff == 0 {
		t.Fatal("endpoint missing")
	}
	if !strings.Contains(Plot([]Point{{math.NaN(), 1}}, 30, 6, "s"), "No recorded") {
		t.Fatal("invalid included")
	}
	for _, size := range []int{0, 1, 5, 17} {
		_ = Plot([]Point{{0, 0}}, size, 2, "s")
	}
}
