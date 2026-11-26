package ui

import (
	"strings"
	"testing"
	"time"

	"github.com/gaurv-m/keywi/assets"
	"github.com/gaurv-m/keywi/internal/theme"
	"github.com/gaurv-m/keywi/internal/typing"
)

func TestDisplayWidthAndLogoBlock(t *testing.T) {
	if Width("a界e\u0301") != 4 {
		t.Fatal("width")
	}
	logo := assets.Logo()
	rendered := CenterBlock(logo, 80)
	raw := strings.Split(logo, "\n")
	lines := strings.Split(rendered, "\n")
	bw := Width(logo)
	pad := (80 - bw) / 2
	for i, l := range lines {
		if !strings.HasPrefix(l, strings.Repeat(" ", pad)+raw[i]) {
			t.Fatalf("line %d independently centered", i)
		}
		if Width(l) != pad+bw {
			t.Fatal("rectangular block")
		}
	}
}
func TestSmallBoundsAndResize(t *testing.T) {
	for w := 1; w <= 100; w++ {
		for h := 1; h <= 30; h++ {
			view := Place(CenterBlock("one\n界 two\nthree", w), w, h)
			lines := strings.Split(view, "\n")
			if len(lines) > h {
				t.Fatal("height", w, h)
			}
			for _, l := range lines {
				if Width(l) > w {
					t.Fatal("width", w, h)
				}
			}
		}
	}
}
func TestWrapAndTyping(t *testing.T) {
	text := []rune("one two three four 界界 five")
	for w := 2; w < 30; w++ {
		lines := Wrap(text, w)
		last := 0
		for _, l := range lines {
			if l.Start != last || l.End <= l.Start || Width(string(text[l.Start:l.End])) > w {
				t.Fatalf("bad wrap %+v width %d", l, w)
			}
			last = l.End
		}
		if last != len(text) {
			t.Fatal("lost text")
		}
	}
	g := typing.New(string(text))
	for _, style := range []string{"block", "line", "underline"} {
		s := Typing(&g, 20, style, theme.New("kiwi"))
		if len(strings.Split(s, "\n")) != 3 {
			t.Fatal("typing line count")
		}
	}
}

func TestCenteredWrappedSentenceKeepsLeftAlignment(t *testing.T) {
	tracker := typing.New("A quiet river winds between the hills before it reaches the coast.")
	text := Typing(&tracker, 32, "block", theme.New("kiwi"))
	centered := CenterBlock(text, 40)
	rows := strings.Split(Plain(centered), "\n")
	left := -1
	for _, line := range rows {
		if strings.TrimSpace(line) == "" {
			continue
		}
		offset := len(line) - len(strings.TrimLeft(line, " "))
		if left >= 0 && offset != left {
			t.Fatal("wrapped lines were independently centered")
		}
		left = offset
	}
	if Width(centered) > 40 {
		t.Fatal("centered sentence exceeds region")
	}
}
func TestLogoColoursPreserveArtwork(t *testing.T) {
	for _, name := range theme.Names() {
		got := Logo(assets.Logo(), theme.New(name))
		if Plain(got) != assets.Logo() {
			t.Fatalf("theme %s changed artwork", name)
		}
		if Width(got) != Width(assets.Logo()) {
			t.Fatalf("theme %s changed width", name)
		}
	}
}

func TestNavigationStaysOnBottomRow(t *testing.T) {
	for _, size := range [][2]int{{24, 10}, {40, 12}, {80, 24}, {120, 40}} {
		page := Page(CenterBlock("content", size[0]), CenterLine("esc back", size[0]), size[0], size[1])
		rows := strings.Split(page, "\n")
		if len(rows) != size[1] || strings.TrimSpace(rows[size[1]-2]) != "esc back" || strings.TrimSpace(rows[size[1]-1]) != "" {
			t.Fatal("navigation moved away from bottom", size)
		}
	}
}

func TestBlockedCursorPulsesWithoutMoving(t *testing.T) {
	logoColors(t)
	now := time.Now()
	tracker := typing.New("abcde")
	tracker.Type('x', now)
	tracker.Type('y', now)
	p := theme.New("kiwi")
	for _, cursor := range []string{"block", "underline", "line"} {
		bright := TypingFrame(&tracker, 10, cursor, p, now.Add(40*time.Millisecond), true)
		dim := TypingFrame(&tracker, 10, cursor, p, now.Add(80*time.Millisecond), true)
		if bright == dim || Plain(bright) != Plain(dim) || Width(bright) > 10 || Width(dim) > 10 {
			t.Fatal("cursor feedback moved the text or failed to pulse", cursor)
		}
		if cursor == "block" && (!strings.Contains(bright, p.Error.Reverse(true).Render("c")) || !strings.Contains(dim, p.Error.Foreground(p.Muted.GetForeground()).Faint(true).Reverse(true).Render("c"))) {
			t.Fatal("block pulse moved away from the insertion cell")
		}
		static := TypingFrame(&tracker, 10, cursor, p, now.Add(40*time.Millisecond), false)
		if static != TypingFrame(&tracker, 10, cursor, p, now.Add(80*time.Millisecond), false) {
			t.Fatal("Animations off still pulsed the cursor")
		}
	}
	for _, name := range theme.Names() {
		colors := theme.New(name)
		spaces := typing.New("Th waves")
		spaces.Type('x', now)
		spaces.Type('y', now)
		spaceFrame := TypingFrame(&spaces, 20, "block", colors, now.Add(80*time.Millisecond), true)
		prefix, suffix, _ := strings.Cut(colors.Error.Foreground(colors.Muted.GetForeground()).Faint(true).Reverse(true).Render("x"), "x")
		if !strings.Contains(spaceFrame, prefix+" "+suffix) {
			t.Fatal("block cursor disappeared when pulsing on a space", name)
		}
	}
	atEnd := typing.New("ab")
	atEnd.Type('x', now)
	atEnd.Type('y', now)
	view := TypingFrame(&atEnd, 10, "block", p, now.Add(time.Second), true)
	if !strings.Contains(view, p.Error.Reverse(true).Render("b")) {
		t.Fatal("mistake at target end hid the correction cursor")
	}
}
