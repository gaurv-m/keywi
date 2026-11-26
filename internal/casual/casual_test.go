package casual

import (
	"strings"
	"testing"
	"time"

	"github.com/gaurv-m/keywi/assets"
	"github.com/gaurv-m/keywi/internal/typing"
)

func TestTimerAndDeadline(t *testing.T) {
	g := New(DefaultConfig())
	now := time.Now()
	g.Advance(now.Add(time.Hour))
	if g.Done || g.Remaining(now) != 30 {
		t.Fatal("started without input")
	}
	g.Type(g.Tracker.Target[0], now)
	g.Type('x', now.Add(30*time.Second))
	if !g.Done || len(g.Tracker.Input) != 1 {
		t.Fatal("late input accepted")
	}
	if g.Tracker.Metrics(now).Duration != 30 {
		t.Fatal("deadline not exact")
	}
}
func TestWordCountAndRestart(t *testing.T) {
	for _, n := range []int{10, 25, 50, 100} {
		g := New(Config{Kind: "words", Amount: n})
		now := time.Now()
		for _, r := range g.Tracker.Target {
			g.Type(r, now)
			now = now.Add(10 * time.Millisecond)
		}
		if !g.Done || !g.Tracker.Perfect() {
			t.Fatalf("words %d", n)
		}
		fresh := New(g.Config)
		if fresh.Done || len(fresh.Tracker.Input) != 0 || !fresh.Tracker.Started.IsZero() {
			t.Fatal("restart state")
		}
	}
}
func TestTimedExtensionAndBackspace(t *testing.T) {
	g := New(DefaultConfig())
	now := time.Now()
	original := len(g.Tracker.Target)
	for i := 0; i < original-100; i++ {
		g.Type(g.Tracker.Target[i], now)
	}
	if len(g.Tracker.Target) <= original {
		t.Fatal("did not generate more text")
	}
	n := len(g.Tracker.Input)
	g.Backspace(now)
	if len(g.Tracker.Input) != n-1 {
		t.Fatal("backspace")
	}
}
func TestConfiguration(t *testing.T) {
	for _, c := range []Config{{Kind: "time", Amount: 15}, {Kind: "time", Amount: 120, Punctuation: true, Numbers: true}, {Kind: "words", Amount: 100, Punctuation: true}} {
		if !c.Valid() {
			t.Fatal(c)
		}
	}
	if (Config{Kind: "time", Amount: 10}).Valid() {
		t.Fatal("bad config")
	}
	if New(Config{}).Config != DefaultConfig() {
		t.Fatal("fallback")
	}
}

func TestSentenceCapitalization(t *testing.T) {
	start := true
	input := []string{"quiet", "water,", "i", "rest.", "123", "\"beyond", "the", "hill!\"", "élan", "returns?", "yes"}
	want := []string{"Quiet", "water,", "I", "rest.", "123", "\"Beyond", "the", "hill!\"", "Élan", "returns?", "Yes"}
	for i, word := range input {
		got, next := capitalizeWord(word, start)
		if got != want[i] {
			t.Fatalf("%q -> %q, want %q", word, got, want[i])
		}
		start = next
	}
}
func TestCapitalizationAcrossGeneratedBatches(t *testing.T) {
	g := New(Config{Kind: "words", Amount: 10, Capitalization: true})
	g.words = []string{"palm"}
	g.sentenceStart = false
	if got := g.generate(2); got != "palm palm" {
		t.Fatal("batch boundary introduced a capital", got)
	}
	g.sentenceStart = true
	if got := g.generate(2); got != "Palm palm" {
		t.Fatal("pending sentence capital lost", got)
	}
	g.Config.Capitalization = false
	g.sentenceStart = true
	if got := g.generate(2); got != "palm palm" {
		t.Fatal("disabled capitals ignored", got)
	}
}
func TestCapitalizationConfigurationKeys(t *testing.T) {
	c := Config{Kind: "time", Amount: 30}
	if c.Key() != "time/30/punctuation=false/numbers=false" {
		t.Fatal("legacy lowercase PB key changed")
	}
	old := c.Key()
	c.Capitalization = true
	if c.Key() == old || !DefaultConfig().Capitalization {
		t.Fatal("capitalized scores not separated")
	}
}

func TestTimedTestsUseCompleteParagraphs(t *testing.T) {
	g := New(Config{Kind: "time", Amount: 30})
	text := string(g.Tracker.Target)
	pool := assets.Paragraphs()
	var first string
	for _, p := range pool {
		if strings.HasPrefix(text, p+" ") {
			first = p
			break
		}
	}
	if first == "" {
		t.Fatal("timed test did not start with a complete paragraph")
	}
	remainder := strings.TrimPrefix(text, first+" ")
	found := false
	for _, p := range pool {
		if remainder == p && p != first {
			found = true
		}
	}
	if !found {
		t.Fatal("timed text did not retain complete, different paragraphs")
	}
	if g.Config.Text != "prose" || !g.Config.Punctuation || !g.Config.Capitalization || g.Config.Numbers {
		t.Fatal("prose config does not describe actual content")
	}
}
func TestTimedAppendDoesNotRewriteVisibleText(t *testing.T) {
	g := New(DefaultConfig())
	before := string(g.Tracker.Target)
	now := time.Now()
	for i := 0; i < len([]rune(before))-150; i++ {
		g.Type(g.Tracker.Target[i], now)
	}
	after := string(g.Tracker.Target)
	if len(after) <= len(before) || !strings.HasPrefix(after, before+" ") {
		t.Fatal("append changed existing text")
	}
	remainder := strings.TrimPrefix(after, before+" ")
	found := false
	for _, p := range assets.Paragraphs() {
		if remainder == p {
			found = true
		}
	}
	if !found {
		t.Fatal("appended text was not a complete paragraph")
	}
}
func TestProseBestsDoNotMixWithLegacyTimedWords(t *testing.T) {
	old := Config{Kind: "time", Amount: 30, Punctuation: true, Capitalization: true}
	next := New(old).Config
	if next.Key() == old.Key() {
		t.Fatal("new prose scores mixed with legacy words")
	}
	if next.Label() != "30s · paragraphs" {
		t.Fatal("unreadable prose label")
	}
	words := New(Config{Kind: "words", Amount: 10, Text: "prose"})
	if words.Config.Text != "" || len(strings.Fields(string(words.Tracker.Target))) != 10 {
		t.Fatal("word test changed into paragraphs")
	}
}

func TestTwoMistakesAtWordTestEndMustBeCorrected(t *testing.T) {
	g := New(Config{Kind: "words", Amount: 10})
	g.Tracker = typing.New("ab")
	now := time.Now()
	g.Type('x', now)
	g.Type('y', now)
	g.Type('a', now)
	if g.Done || string(g.Tracker.Input) != "xy" || g.Tracker.Attempts != 2 {
		t.Fatal("mistaken final characters bypassed the limit")
	}
	g.Backspace(now)
	g.Backspace(now)
	g.Type('a', now)
	g.Type('b', now)
	if !g.Done || !g.Tracker.Perfect() {
		t.Fatal("corrected word test did not finish")
	}
}

func TestTimedDeadlineContinuesWhileMistakeLocked(t *testing.T) {
	g := New(Config{Kind: "time", Amount: 15})
	g.Tracker = typing.New("abcdef")
	now := time.Now()
	g.Type('x', now)
	g.Type('y', now)
	g.Type('c', now.Add(time.Second))
	if len(g.Tracker.Input) != 2 {
		t.Fatal("blocked timed test accepted input")
	}
	g.Advance(now.Add(20 * time.Second))
	if !g.Done || g.Tracker.Metrics(now).Duration != 15 {
		t.Fatal("mistake limit paused the deadline")
	}
}

func TestWordsRemainingUsesTargetBoundaries(t *testing.T) {
	now := time.Now()
	g := New(Config{Kind: "words", Amount: 10})
	g.Tracker = typing.New("ab cd ef gh ij kl mn op qr st")
	g.Type(' ', now)
	if g.Remaining(now) != 10 {
		t.Fatal("a mistyped space consumed a word")
	}
	g.Backspace(now)
	g.Type('a', now)
	g.Type('b', now)
	g.Type('x', now)
	if g.Remaining(now) != 9 {
		t.Fatal("a wrong boundary key froze the word counter")
	}
	g.Backspace(now)
	if g.Remaining(now) != 10 {
		t.Fatal("backspace did not restore word position")
	}
	g.Type(' ', now)
	for i := len(g.Tracker.Input); i < len(g.Tracker.Target); i++ {
		g.Type(g.Tracker.Target[i], now)
	}
	if !g.Done || g.Remaining(now) != 0 {
		t.Fatal("completed word test still has words remaining")
	}
}
