package blitz

import (
	"strings"
	"testing"
	"time"

	"github.com/gaurv-m/keywi/internal/typing"
)

func short() *Game { g := New(); g.Tracker = typing.New("ab"); g.Rewarded = make([]bool, 2); return g }
func TestRewardsCapAndFarming(t *testing.T) {
	g := short()
	now := time.Now()
	g.Type('a', now)
	if g.Remaining(now) != 30*time.Second {
		t.Fatal("cap")
	}
	deadline := g.Deadline
	g.Backspace(now.Add(time.Second))
	g.Type('a', now.Add(time.Second))
	if !g.Deadline.Equal(deadline) {
		t.Fatal("time farming")
	}
	g.Type('x', now.Add(2*time.Second))
	if !g.Deadline.Equal(deadline) || g.Sentences != 0 {
		t.Fatal("wrong rewarded")
	}
	g.Backspace(now.Add(2 * time.Second))
	g.Type('b', now.Add(2*time.Second))
	if g.Sentences != 1 {
		t.Fatal("completion")
	}
	if g.Remaining(now.Add(2*time.Second)) != 30*time.Second {
		t.Fatal("sentence cap")
	}
	if len(g.Tracker.Input) != 0 {
		t.Fatal("next not immediately ready")
	}
}
func TestExactRewards(t *testing.T) {
	g := short()
	now := time.Now()
	g.Type('a', now)
	g.Type('b', now.Add(10*time.Second))
	want := now.Add(32100 * time.Millisecond)
	if !g.Deadline.Equal(want) {
		t.Fatalf("got %v want %v", g.Deadline, want)
	}
	if g.LastCompletion.Bonus != 2*time.Second {
		t.Fatal("bonus")
	}
}
func TestExpiryPrecedesInput(t *testing.T) {
	g := short()
	now := time.Now()
	g.Type('a', now)
	g.Type('b', now.Add(30*time.Second))
	if !g.Done || g.Sentences != 0 || g.Tracker.Metrics(now).Duration != 30 {
		t.Fatal("late input extended deadline")
	}
	g.Backspace(now)
	if len(g.Tracker.Input) != 1 {
		t.Fatal("backspace after expiry")
	}
}
func TestNoEarlyClockAndDeadlineLoad(t *testing.T) {
	g := short()
	now := time.Now()
	g.Advance(now.Add(time.Hour))
	if g.Done {
		t.Fatal("unstarted expiry")
	}
	g.Type('a', now)
	g.Advance(now.Add(5 * time.Hour))
	if g.Tracker.Metrics(now).Duration != 30 {
		t.Fatal("render load extended clock")
	}
}
func TestDifficultyAndRapidTransitions(t *testing.T) {
	g := New()
	now := time.Now()
	for n := 1; n <= 60; n++ {
		target := append([]rune(nil), g.Tracker.Target...)
		for _, r := range target {
			g.Type(r, now)
		}
		if g.Sentences != n || g.Level != 1+n/5 {
			t.Fatalf("progress %d level %d", g.Sentences, g.Level)
		}
		if g.Remaining(now) > 30*time.Second {
			t.Fatal("cap")
		}
	}
	a, b := g.Balance.Rewards(100)
	if a != 40*time.Millisecond || b != time.Second {
		t.Fatalf("unbounded %v %v", a, b)
	}
}
func TestUncorrectedSentence(t *testing.T) {
	g := short()
	now := time.Now()
	g.Type('x', now)
	g.Type('b', now)
	if g.Sentences != 0 {
		t.Fatal("error sentence completed")
	}
	g.Backspace(now)
	g.Backspace(now)
	g.Type('a', now)
	g.Type('b', now)
	if g.Sentences != 1 {
		t.Fatal("corrected sentence not completed")
	}
}

func TestBlitzSentenceTargetsOmitClosingPeriods(t *testing.T) {
	cases := map[string]string{
		"The tide is low.":          "The tide is low",
		"At 18.7 degrees, we left.": "At 18.7 degrees, we left",
		"\"Please wait.\"":          "\"Please wait\"",
		"Ready?":                    "Ready?",
		"Wait!":                     "Wait!",
	}
	for input, want := range cases {
		if got := sentenceTarget(input); got != want {
			t.Fatalf("%q became %q, want %q", input, got, want)
		}
	}
	g := New()
	now := time.Now()
	for n := 1; n <= 80; n++ {
		target := string(g.Tracker.Target)
		if strings.HasSuffix(strings.TrimRight(target, "\"'”’)]}"), ".") {
			t.Fatal("sentence still required a closing period", target)
		}
		for _, r := range target {
			g.Type(r, now)
		}
		if g.Sentences != n || g.LastCompletion.Text != target {
			t.Fatal("last visible character did not complete the sentence")
		}
	}
}

func TestBlockedBlitzKeysEarnNothingAndClockKeepsRunning(t *testing.T) {
	g := New()
	g.Tracker = typing.New("abcd")
	g.Rewarded = make([]bool, 4)
	now := time.Now()
	g.Type('x', now)
	g.Type('y', now)
	deadline := g.Deadline
	g.Type('c', now.Add(time.Second))
	if len(g.Tracker.Input) != 2 || g.Tracker.Attempts != 2 || !g.Deadline.Equal(deadline) || g.Rewarded[2] {
		t.Fatal("blocked input earned score or time")
	}
	g.Advance(now.Add(31 * time.Second))
	if !g.Done || g.Tracker.Metrics(now).Duration != 30 {
		t.Fatal("blocked input paused the Blitz clock")
	}
}
