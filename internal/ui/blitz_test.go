package ui

import (
	"os"
	"strings"
	"testing"
	"time"

	"github.com/gaurv-m/keywi/internal/animation"
	"github.com/gaurv-m/keywi/internal/theme"
	"github.com/gaurv-m/keywi/internal/typing"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	"github.com/muesli/termenv"
)

func TestShatterOriginsMatchFinishedTypingCells(t *testing.T) {
	for _, width := range []int{24, 40, 76} {
		old := typing.New("The ocean glimmers beneath the afternoon sun.")
		old.Input = append([]rune(nil), old.Target...)
		expected := strings.Split(Plain(CenterBlock(Typing(&old, width, "block", theme.New("kiwi")), width)), "\n")
		source := append([]string{strings.Repeat(" ", width)}, expected...)
		source = append(source, strings.Repeat(" ", width))
		frame := animation.ShatterRows(source, 0, width, 5)
		for _, particle := range frame.Particles {
			if particle.Y < 1 || particle.Y > 3 {
				t.Fatal("shatter started outside the original typing rows")
			}
			if ansi.Cut(expected[particle.Y-1], particle.X, particle.X+1) == " " {
				t.Fatal("shard started away from a completed character")
			}
		}
	}
}
func TestShatterCannotCoverNextSentence(t *testing.T) {
	for _, width := range []int{20, 36, 76} {
		for _, cursor := range []string{"block", "underline", "line"} {
			current := typing.New("The waves carry a silver shell.")
			p := theme.New("kiwi")
			baseline := strings.Split(BlitzStage(&current, "", 0, width, cursor, p, true), "\n")
			raw := strings.Split(Typing(&current, width, cursor, p), "\n")
			bw := 0
			for _, row := range raw {
				bw = max(bw, Width(row))
			}
			left := (width - bw) / 2
			for age := time.Duration(0); age <= 300*time.Millisecond; age += 20 * time.Millisecond {
				stage := strings.Split(BlitzStage(&current, "A small boat rests by the shore.", age, width, cursor, p, true), "\n")
				if len(stage) != 5 {
					t.Fatal("stage height changed")
				}
				for y, row := range stage {
					if Width(row) != width {
						t.Fatalf("row %d exceeded width", y)
					}
				}
				for y, text := range raw {
					end := left + Width(text)
					if Plain(ansi.Cut(stage[y+1], left, end)) != Plain(ansi.Cut(baseline[y+1], left, end)) || !strings.Contains(stage[y+1], text) {
						t.Fatalf("shards covered target/cursor at age %s", age)
					}
				}
			}
		}
	}
}
func TestDisabledShatterMatchesIdleStage(t *testing.T) {
	current := typing.New("The warm wind moves the leaves.")
	p := theme.New("kiwi")
	got := BlitzStage(&current, "An old sentence.", 80*time.Millisecond, 76, "block", p, false)
	want := BlitzStage(&current, "", 0, 76, "block", p, true)
	if got != want {
		t.Fatal("disabled animation painted shards")
	}
}

func TestColoredShatterOutputStaysBounded(t *testing.T) {
	t.Setenv("TERM", "xterm-256color")
	priorNoColor, present := os.LookupEnv("NO_COLOR")
	os.Unsetenv("NO_COLOR")
	t.Cleanup(func() {
		if present {
			os.Setenv("NO_COLOR", priorNoColor)
		}
	})
	renderer := lipgloss.DefaultRenderer()
	previous := renderer.ColorProfile()
	t.Cleanup(func() { renderer.SetColorProfile(previous) })
	for _, profile := range []termenv.Profile{termenv.TrueColor, termenv.ANSI256, termenv.ANSI} {
		renderer.SetColorProfile(profile)
		current := typing.New("The afternoon light falls across the pages of an open book beside the narrow kitchen window.")
		p := theme.New("contrast")
		for age := 33 * time.Millisecond; age < 220*time.Millisecond; age += 33 * time.Millisecond {
			frame := BlitzStage(&current, "A quiet river winds between the hills before it reaches the coast.", age, 76, "block", p, true)
			if len(frame) > 16384 {
				t.Fatalf("formatting codes expanded to %d bytes", len(frame))
			}
			for _, line := range strings.Split(frame, "\n") {
				if Width(line) != 76 {
					t.Fatal("colored frame exceeded its surface")
				}
			}
		}
	}
}
