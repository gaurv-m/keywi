// Package blitz runs the survival mode.
package blitz

import (
	"strings"
	"time"

	"github.com/gaurv-m/keywi/assets"
	"github.com/gaurv-m/keywi/internal/typing"
)

// Balance sets timers, rewards, and level scaling.
type Balance struct {
	Capacity, Character, Sentence            time.Duration
	SentencesPerLevel, ScaleFrom, ScaleSteps int
	MinimumRewardFactor                      float64
}

var DefaultBalance = Balance{30 * time.Second, 100 * time.Millisecond, 2 * time.Second, 5, 5, 6, 0.4}

func (b Balance) Factor(level int) float64 {
	steps := max(0, min(b.ScaleSteps, level-b.ScaleFrom+1))
	return max(b.MinimumRewardFactor, 1-float64(steps)*0.1)
}
func (b Balance) Rewards(level int) (time.Duration, time.Duration) {
	f := b.Factor(level)
	return time.Duration(float64(b.Character) * f), time.Duration(float64(b.Sentence) * max(0.5, f))
}

type Completion struct {
	Text  string
	Bonus time.Duration
	At    time.Time
}
type Game struct {
	Tracker          typing.Tracker
	Balance          Balance
	Deck             *assets.Deck
	Deadline         time.Time
	Rewarded         []bool
	Sentences, Level int
	Done             bool
	LastCompletion   Completion
}

func New() *Game {
	return NewWithDeck(assets.NewDeck())
}

func NewWithDeck(d *assets.Deck) *Game {
	if d == nil {
		d = assets.NewDeck()
	}
	text := sentenceTarget(d.Next(1))
	return &Game{Tracker: typing.New(text), Balance: DefaultBalance, Deck: d, Rewarded: make([]bool, len([]rune(text))), Level: 1}
}
func (g *Game) Remaining(now time.Time) time.Duration {
	if g.Tracker.Started.IsZero() {
		return g.Balance.Capacity
	}
	if g.Done {
		return 0
	}
	return max(time.Duration(0), g.Deadline.Sub(now))
}
func (g *Game) Advance(now time.Time) {
	if g.Done || g.Tracker.Started.IsZero() {
		return
	}
	if !now.Before(g.Deadline) {
		g.Done = true
		g.Tracker.Finish(g.Deadline)
		return
	}
	g.Tracker.Sample(now, false)
}
func (g *Game) reward(amount time.Duration, now time.Time) time.Duration {
	old := g.Deadline
	g.Deadline = minTime(g.Deadline.Add(amount), now.Add(g.Balance.Capacity))
	return g.Deadline.Sub(old)
}
func minTime(a, b time.Time) time.Time {
	if a.Before(b) {
		return a
	}
	return b
}
func (g *Game) Type(r rune, now time.Time) {
	g.Advance(now)
	if g.Done || !typing.Valid(r) {
		return
	}
	if g.Tracker.Blocked() {
		g.Tracker.Type(r, now)
		return
	}
	if len(g.Tracker.Input) >= len(g.Tracker.Target) {
		g.Tracker.Type(r, now)
		return
	}
	if g.Tracker.Started.IsZero() {
		g.Tracker.Start(now)
		g.Deadline = now.Add(g.Balance.Capacity)
	}
	pos := len(g.Tracker.Input)
	correct := g.Tracker.Type(r, now)
	charBonus, sentenceBonus := g.Balance.Rewards(g.Level)
	if correct && !g.Rewarded[pos] {
		g.Rewarded[pos] = true
		g.reward(charBonus, now)
	}
	if g.Tracker.Perfect() {
		text := string(g.Tracker.Target)
		bonus := g.reward(sentenceBonus, now)
		g.Sentences++
		g.Level = 1 + g.Sentences/g.Balance.SentencesPerLevel
		g.LastCompletion = Completion{text, bonus, now}
		next := sentenceTarget(g.Deck.Next(g.Level))
		g.Tracker.Next(next)
		g.Rewarded = make([]bool, len(g.Tracker.Target))
	}
}

// Keep internal punctuation; omit trailing periods.
func sentenceTarget(text string) string {
	text = strings.TrimSpace(text)
	body := strings.TrimRight(text, "\"'”’)]}")
	return strings.TrimRight(body, ".") + text[len(body):]
}
func (g *Game) Backspace(now time.Time) {
	g.Advance(now)
	if !g.Done {
		g.Tracker.Backspace()
	}
}
