package typing

import (
	"math"
	"testing"
	"time"
)

func near(t *testing.T, got, want float64) {
	t.Helper()
	if math.Abs(got-want) > 0.00001 {
		t.Fatalf("got %v want %v", got, want)
	}
}
func TestCorrectionsAndFormulas(t *testing.T) {
	now := time.Now()
	g := New("cat")
	for _, r := range "cx" {
		g.Type(r, now)
	}
	g.Backspace()
	g.Type('a', now)
	g.Type('t', now)
	if !g.Perfect() {
		t.Fatal("correction did not complete target")
	}
	g.Finish(now.Add(30 * time.Second))
	m := g.Metrics(now.Add(time.Hour))
	near(t, m.WPM, 1.2)
	near(t, m.RawWPM, 1.6)
	near(t, m.Accuracy, 75)
	near(t, m.Duration, 30)
	if m.Correct != 3 || m.Incorrect != 0 || m.Errors != 1 || m.Keystrokes != 4 {
		t.Fatalf("%+v", m)
	}
}
func TestUnicodeAndInput(t *testing.T) {
	g := New("界é")
	now := time.Now()
	g.Backspace()
	g.Type('\n', now)
	if !g.Started.IsZero() {
		t.Fatal("control started timer")
	}
	g.Type('界', now)
	g.Type('x', now)
	if g.Perfect() {
		t.Fatal("accepted error")
	}
	g.Backspace()
	g.Type('é', now)
	if !g.Perfect() {
		t.Fatal("unicode correction")
	}
	if g.Type('a', now) {
		t.Fatal("accepted overflow")
	}
}
func TestConsistency(t *testing.T) {
	near(t, Consistency(nil), 0)
	near(t, Consistency([]float64{0, 0}), 0)
	near(t, Consistency([]float64{60, 60}), 100)
	near(t, Consistency([]float64{30, 90}), 50)
}
func TestBankingAndSamples(t *testing.T) {
	now := time.Now()
	g := New("ab")
	g.Type('a', now)
	g.Sample(now.Add(time.Second), false)
	g.Type('b', now.Add(time.Second))
	g.Next("cd")
	g.Type('c', now.Add(2*time.Second))
	g.Sample(now.Add(2*time.Second), false)
	c, e := g.Counts()
	if c != 3 || e != 0 || len(g.Samples) != 2 {
		t.Fatalf("%+v", g)
	}
	near(t, g.Samples[0].WPM, 12)
	near(t, g.Samples[1].WPM, 24)
}

func TestTwoMistakeLimitAndCorrection(t *testing.T) {
	now := time.Now()
	g := New("abcde")
	for _, r := range "xby" {
		g.Type(r, now)
	}
	if !g.Blocked() || g.Uncorrected() != 2 || len(g.Input) != 3 {
		t.Fatal("limit did not count separated mistakes")
	}
	for _, r := range "defgh" {
		g.Type(r, now.Add(time.Second))
	}
	if len(g.Input) != 3 || g.Attempts != 3 || g.CorrectAttempts != 1 {
		t.Fatal("locked keys changed input or scoring")
	}
	if !g.BlockedAt.Equal(now.Add(time.Second)) {
		t.Fatal("rejected key did not trigger cursor feedback")
	}
	g.Backspace()
	if g.Blocked() || !g.BlockedAt.IsZero() {
		t.Fatal("backspace did not unlock input")
	}
	g.Type('c', now.Add(time.Second))
	if string(g.Input) != "xbc" || g.Attempts != 4 || g.Uncorrected() != 1 {
		t.Fatal("corrected input did not resume")
	}
	if g.Metrics(now.Add(time.Second)).Errors != 2 {
		t.Fatal("correcting a mistake erased historical errors")
	}
}

func TestBankedMistakesDoNotBlockAnotherSentence(t *testing.T) {
	g := New("ab")
	now := time.Now()
	g.Type('x', now)
	g.Type('y', now)
	g.Next("cd")
	if g.Blocked() || g.CorrectionRequired() || !g.BlockedAt.IsZero() {
		t.Fatal("past errors locked the new sentence")
	}
	if !g.Type('c', now) {
		t.Fatal("new sentence did not accept correct input")
	}
	_, errors := g.Counts()
	if errors != 2 {
		t.Fatal("banking lost past mistakes")
	}
}
