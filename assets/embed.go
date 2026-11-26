// Package assets holds the text and logo.
package assets

import (
	"embed"
	"math/rand/v2"
	"strings"
)

//go:embed logo.txt words.txt sentences.txt paragraphs.txt
var FS embed.FS

func Logo() string    { b, _ := FS.ReadFile("logo.txt"); return strings.TrimSuffix(string(b), "\n") }
func Words() []string { b, _ := FS.ReadFile("words.txt"); return strings.Fields(string(b)) }

// Sentences returns the original sentence library grouped by difficulty (1–5).
func Sentences() map[int][]string {
	b, _ := FS.ReadFile("sentences.txt")
	pools := make(map[int][]string)
	for _, line := range strings.Split(string(b), "\n") {
		if len(line) > 2 && line[0] >= '1' && line[0] <= '5' && line[1] == '|' {
			pools[int(line[0]-'0')] = append(pools[int(line[0]-'0')], line[2:])
		}
	}
	return pools
}

// Deck exhausts each shuffled pool before repeating.
type Deck struct {
	Pools map[int][]string
	order map[int][]int
	last  map[int]int
}

func NewDeck() *Deck { return &Deck{Pools: Sentences(), order: map[int][]int{}, last: map[int]int{}} }
func (d *Deck) Next(level int) string {
	// Higher levels keep using the hardest sentence pool.
	level = min(5, max(1, level))
	pool := d.Pools[level]
	if len(pool) == 0 {
		return "The quiet sea reflects the sky."
	}
	if len(d.order[level]) == 0 {
		p := rand.Perm(len(pool))
		if len(p) > 1 && p[0] == d.last[level]-1 {
			p[0], p[1] = p[1], p[0]
		}
		d.order[level] = p
	}
	n := d.order[level][0]
	d.order[level] = d.order[level][1:]
	d.last[level] = n + 1
	return pool[n]
}

// Paragraphs are original, self-contained passages kept intact when shuffled.
func Paragraphs() []string {
	b, _ := FS.ReadFile("paragraphs.txt")
	var out []string
	for _, block := range strings.Split(strings.TrimSpace(string(b)), "\n\n") {
		paragraph := strings.Join(strings.Fields(block), " ")
		if paragraph != "" {
			out = append(out, paragraph)
		}
	}
	return out
}
func NewParagraphDeck() *Deck {
	return &Deck{Pools: map[int][]string{1: Paragraphs()}, order: map[int][]int{}, last: map[int]int{}}
}
