// Package casual runs timed and word-count tests.
package casual

import (
	"fmt"
	"math/rand/v2"
	"strings"
	"time"
	"unicode"

	"github.com/gaurv-m/keywi/assets"
	"github.com/gaurv-m/keywi/internal/typing"
)

type Config struct {
	Kind           string `json:"kind"`
	Amount         int    `json:"amount"`
	Punctuation    bool   `json:"punctuation"`
	Numbers        bool   `json:"numbers"`
	Capitalization bool   `json:"capitalization"`
	Text           string `json:"text,omitempty"`
}

func DefaultConfig() Config {
	return Config{Kind: "time", Amount: 30, Punctuation: true, Capitalization: true, Text: "prose"}
}
func (c Config) Key() string {
	key := fmt.Sprintf("%s/%d/punctuation=%t/numbers=%t", c.Kind, c.Amount, c.Punctuation, c.Numbers)
	if c.Capitalization {
		key += "/capitalization=true"
	}
	if c.Text == "prose" {
		key += "/text=prose"
	}
	return key
}
func (c Config) Valid() bool {
	allowed := []int{15, 30, 60, 120}
	if c.Kind == "words" {
		allowed = []int{10, 25, 50, 100}
	} else if c.Kind != "time" {
		return false
	}
	for _, n := range allowed {
		if c.Amount == n {
			return true
		}
	}
	return false
}

type Game struct {
	Config        Config
	Tracker       typing.Tracker
	Done          bool
	words         []string
	sentenceStart bool
	paragraphs    *assets.Deck
}

func New(c Config) *Game {
	return NewWithDeck(c, nil)
}

func NewWithDeck(c Config, paragraphs *assets.Deck) *Game {
	if !c.Valid() {
		c = DefaultConfig()
	}
	// Timed passages keep their original capitals and punctuation.
	if c.Kind == "time" {
		c.Text = "prose"
		c.Punctuation = true
		c.Capitalization = true
		c.Numbers = false
	} else {
		c.Text = ""
	}
	g := &Game{Config: c, words: assets.Words(), sentenceStart: true}
	if c.Kind == "time" {
		g.paragraphs = paragraphs
		if g.paragraphs == nil {
			g.paragraphs = assets.NewParagraphDeck()
		}
		g.Tracker = typing.New(g.prose(2))
	} else {
		g.Tracker = typing.New(g.generate(c.Amount))
	}
	return g
}
func (g *Game) prose(n int) string {
	parts := make([]string, n)
	for i := range parts {
		parts[i] = g.paragraphs.Next(1)
	}
	return strings.Join(parts, " ")
}
func (g *Game) generate(n int) string {
	words := make([]string, n)
	for i := range words {
		w := g.words[rand.IntN(len(g.words))]
		if g.Config.Numbers && rand.IntN(8) == 0 {
			w = fmt.Sprint(rand.IntN(1000))
		}
		if g.Config.Punctuation && rand.IntN(5) == 0 {
			w += []string{",", ".", ";", "?", "!"}[rand.IntN(5)]
		}
		if g.Config.Capitalization {
			w, g.sentenceStart = capitalizeWord(w, g.sentenceStart)
		}
		words[i] = w
	}
	return strings.Join(words, " ")
}
func (g *Game) Advance(now time.Time) {
	if g.Done || g.Tracker.Started.IsZero() {
		return
	}
	if g.Config.Kind == "time" {
		end := g.Tracker.Started.Add(time.Duration(g.Config.Amount) * time.Second)
		if !now.Before(end) {
			g.Done = true
			g.Tracker.Finish(end)
			return
		}
	}
	g.Tracker.Sample(now, false)
}
func (g *Game) Type(r rune, now time.Time) {
	g.Advance(now)
	if g.Done || !typing.Valid(r) {
		return
	}
	g.Tracker.Type(r, now)
	if g.Config.Kind == "time" && len(g.Tracker.Target)-len(g.Tracker.Input) < 200 {
		g.Tracker.Target = append(g.Tracker.Target, []rune(" "+g.prose(1))...)
	}
	if g.Config.Kind == "words" && g.Tracker.Complete() && !g.Tracker.Blocked() {
		g.Done = true
		g.Tracker.Finish(now)
	}
}
func (g *Game) Backspace(now time.Time) {
	g.Advance(now)
	if !g.Done {
		g.Tracker.Backspace()
	}
}
func (g *Game) Remaining(now time.Time) float64 {
	if g.Config.Kind == "words" {
		if g.Done {
			return 0
		}
		remaining := g.Config.Amount
		for _, r := range g.Tracker.Target[:min(len(g.Tracker.Input), len(g.Tracker.Target))] {
			if r == ' ' {
				remaining--
			}
		}
		return float64(max(0, remaining))
	}
	if g.Tracker.Started.IsZero() {
		return float64(g.Config.Amount)
	}
	return max(0, float64(g.Config.Amount)-now.Sub(g.Tracker.Started).Seconds())
}

// Keep capitalization state across text batches.
func capitalizeWord(word string, start bool) (string, bool) {
	runes := []rune(word)
	core := strings.Trim(word, "\"'()[]{}.,;:!?")
	pronoun := core == "i" || strings.HasPrefix(core, "i'") || strings.HasPrefix(core, "i’")
	for i, r := range runes {
		if !unicode.IsLetter(r) {
			continue
		}
		if start || pronoun {
			runes[i] = unicode.ToUpper(r)
		}
		start = false
		break
	}
	result := string(runes)
	end := strings.TrimRight(result, "\"')]}:")
	if strings.HasSuffix(end, ".") || strings.HasSuffix(end, "!") || strings.HasSuffix(end, "?") {
		start = true
	}
	return result, start
}

// ShortLabel names the selected preset.
func (c Config) ShortLabel() string {
	if c.Kind == "words" {
		return fmt.Sprintf("%d words", c.Amount)
	}
	return fmt.Sprintf("%ds", c.Amount)
}
func (c Config) Label() string {
	if c.Text == "prose" {
		return c.ShortLabel() + " · paragraphs"
	}
	options := []string{"lowercase"}
	if c.Capitalization {
		options[0] = "capitals"
	}
	if c.Punctuation {
		options = append(options, "punctuation")
	}
	if c.Numbers {
		options = append(options, "numbers")
	}
	return c.ShortLabel() + " · " + strings.Join(options, " · ")
}
