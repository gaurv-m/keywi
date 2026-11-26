package assets

import (
	"crypto/sha256"
	"fmt"
	"strings"
	"testing"
)

func TestLibraries(t *testing.T) {
	if len(Words()) < 300 {
		t.Fatal("word pool too small")
	}
	seen := map[string]bool{}
	total := 0
	for level, pool := range Sentences() {
		if len(pool) < 64 {
			t.Fatal(level)
		}
		total += len(pool)
		for _, s := range pool {
			if seen[s] {
				t.Fatal("duplicate sentence")
			}
			if s == "" || s[0] < 'A' || s[0] > 'Z' {
				if !strings.HasPrefix(s, "\"") {
					t.Fatal("missing opening capital", s)
				}
			}
			seen[s] = true
		}
	}
	if total < 323 {
		t.Fatal("sentence library too small", total)
	}
}
func TestDeckNoRepeat(t *testing.T) {
	d := NewDeck()
	for level := 1; level <= 5; level++ {
		for cycle := 0; cycle < 3; cycle++ {
			seen := map[string]bool{}
			for range d.Pools[level] {
				s := d.Next(level)
				if seen[s] {
					t.Fatal("repeated before exhaustion")
				}
				seen[s] = true
			}
		}
	}
}
func TestExactLogo(t *testing.T) {
	raw, _ := FS.ReadFile("logo.txt")
	if len(strings.Split(strings.TrimSuffix(string(raw), "\n"), "\n")) != 9 {
		t.Fatal("logo height")
	}
	if got := fmt.Sprintf("%x", sha256.Sum256(raw)); got != "5ea9e2bf4b03025a2f2b2f441cfb51fb85cf1b10daa72fe016bdb70c895ad307" {
		t.Fatalf("logo modified %s", got)
	}
}

func TestParagraphDeckPreservesCompletePassages(t *testing.T) {
	pool := Paragraphs()
	if len(pool) < 32 {
		t.Fatal("paragraph pool too small")
	}
	d := NewParagraphDeck()
	last := ""
	for cycle := 0; cycle < 3; cycle++ {
		seen := map[string]bool{}
		for i := range pool {
			paragraph := d.Next(1)
			if seen[paragraph] {
				t.Fatal("paragraph repeated before exhaustion")
			}
			if i == 0 && paragraph == last {
				t.Fatal("paragraph repeated at reshuffle")
			}
			if strings.Count(paragraph, ".") < 3 || strings.ContainsAny(paragraph, "\n\t") {
				t.Fatal("paragraph is not a complete passage")
			}
			seen[paragraph] = true
			last = paragraph
		}
	}
}
