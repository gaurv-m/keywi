package animation

import (
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/charmbracelet/x/ansi"
)

func TestShatterBoundsAndLifetime(t *testing.T) {
	for width := 1; width < 80; width++ {
		for ms := -10; ms < 300; ms += 10 {
			f := Shatter("The ocean glimmers beneath the afternoon sun.", time.Duration(ms)*time.Millisecond, width, 2)
			if len(f.Rows) != 2 {
				t.Fatal("canvas height changed")
			}
			for _, line := range f.Rows {
				if ansi.StringWidth(line) != width {
					t.Fatalf("canvas width changed: %d %q", width, line)
				}
				if (ms < 0 || ms >= 220) && strings.TrimSpace(line) != "" {
					t.Fatal("shards outlived animation")
				}
			}
		}
	}
	for _, size := range [][2]int{{0, 2}, {2, 0}, {-1, 2}} {
		if len(Shatter("abc", 0, size[0], size[1]).Rows) != 0 {
			t.Fatal("invalid canvas accepted")
		}
	}
}
func TestShatterIsDeterministicAndDisperses(t *testing.T) {
	text := "The waves carry a silver shell."
	start := Shatter(text, 0, 60, 2)
	if len(start.Particles) == 0 || start.Particles[0].X != (60-ansi.StringWidth(text))/2 || start.Particles[0].Y != 0 {
		t.Fatal("effect did not originate from finished text")
	}
	moving := Shatter(text, 80*time.Millisecond, 60, 2)
	if !reflect.DeepEqual(moving, Shatter(text, 80*time.Millisecond, 60, 2)) {
		t.Fatal("shards changed unpredictably")
	}
	if !strings.ContainsAny(strings.Join(moving.Rows, ""), "▒⠂⠄▪·") {
		t.Fatal("shatter fragments missing")
	}
	late := Shatter(text, 200*time.Millisecond, 60, 2)
	count := func(f Frame) int {
		n := 0
		for _, line := range f.Rows {
			for _, r := range line {
				if r != ' ' {
					n++
				}
			}
		}
		return n
	}
	if count(late) >= count(moving) {
		t.Fatal("shards did not disperse")
	}
	if !late.Faint {
		t.Fatal("shards did not fade")
	}
}
func TestShatterHandlesWrappedUnicode(t *testing.T) {
	for ms := 0; ms <= 220; ms += 20 {
		f := Shatter("A quiet river meets 界界 and e\u0301 by the shore.", time.Duration(ms)*time.Millisecond, 24, 2)
		for _, line := range f.Rows {
			if ansi.StringWidth(line) != 24 {
				t.Fatal("Unicode shattered outside canvas")
			}
		}
	}
}
