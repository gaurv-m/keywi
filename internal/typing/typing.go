// Package typing tracks input and scores.
package typing

import (
	"math"
	"time"
	"unicode"
)

type Sample struct {
	Seconds float64 `json:"seconds"`
	WPM     float64 `json:"wpm"`
}
type Metrics struct {
	WPM         float64  `json:"wpm"`
	RawWPM      float64  `json:"raw_wpm"`
	Accuracy    float64  `json:"accuracy"`
	Consistency float64  `json:"consistency"`
	Correct     int      `json:"correct"`
	Incorrect   int      `json:"incorrect"`
	Keystrokes  int      `json:"keystrokes"`
	Errors      int      `json:"errors"`
	Duration    float64  `json:"duration"`
	Samples     []Sample `json:"samples"`
}

type Tracker struct {
	Target, Input                                             []rune
	Attempts, CorrectAttempts, BankedCorrect, BankedIncorrect int
	Started, Ended                                            time.Time
	BlockedAt                                                 time.Time
	Samples                                                   []Sample
	lastSample                                                time.Time
	lastCorrect                                               int
}

const MistakeLimit = 2

func New(text string) Tracker { return Tracker{Target: []rune(text)} }
func Valid(r rune) bool       { return r == ' ' || (unicode.IsPrint(r) && !unicode.IsControl(r)) }
func (t *Tracker) Start(now time.Time) {
	if t.Started.IsZero() {
		t.Started = now
		t.lastSample = now
	}
}
func (t *Tracker) Type(r rune, now time.Time) bool {
	if !Valid(r) {
		return false
	}
	if t.Blocked() {
		t.BlockedAt = now
		return false
	}
	if len(t.Input) >= len(t.Target) {
		if t.CorrectionRequired() {
			t.BlockedAt = now
		}
		return false
	}
	t.Start(now)
	t.Attempts++
	correct := r == t.Target[len(t.Input)]
	if correct {
		t.CorrectAttempts++
	}
	t.Input = append(t.Input, r)
	if t.CorrectionRequired() {
		t.BlockedAt = now
	}
	return correct
}
func (t *Tracker) Backspace() {
	if len(t.Input) > 0 {
		t.Input = t.Input[:len(t.Input)-1]
		if !t.CorrectionRequired() {
			t.BlockedAt = time.Time{}
		}
	}
}

// Uncorrected counts errors still in the editable text.
func (t *Tracker) Uncorrected() int {
	errors := 0
	for i, r := range t.Input {
		if i >= len(t.Target) || r != t.Target[i] {
			errors++
		}
	}
	return errors
}

func (t *Tracker) Blocked() bool { return t.Uncorrected() >= MistakeLimit }

// A complete but imperfect Blitz sentence also needs a visible correction cue.
func (t *Tracker) CorrectionRequired() bool {
	errors := t.Uncorrected()
	return errors >= MistakeLimit || (t.Complete() && errors > 0)
}
func (t *Tracker) Counts() (int, int) {
	c, e := t.BankedCorrect, t.BankedIncorrect
	for i, r := range t.Input {
		if i < len(t.Target) && r == t.Target[i] {
			c++
		} else {
			e++
		}
	}
	return c, e
}
func (t *Tracker) Complete() bool { return len(t.Input) == len(t.Target) }
func (t *Tracker) Perfect() bool {
	if !t.Complete() {
		return false
	}
	for i, r := range t.Input {
		if r != t.Target[i] {
			return false
		}
	}
	return true
}

// Next keeps run totals while replacing the editable sentence.
func (t *Tracker) Next(text string) {
	c, e := t.Counts()
	t.BankedCorrect, t.BankedIncorrect = c, e
	t.Target = []rune(text)
	t.Input = nil
	t.BlockedAt = time.Time{}
}
func (t *Tracker) Sample(now time.Time, force bool) {
	if t.Started.IsZero() {
		return
	}
	if !t.Ended.IsZero() {
		now = t.Ended
	}
	dt := now.Sub(t.lastSample).Seconds()
	if dt <= 0 || (!force && dt < 1) {
		return
	}
	c, _ := t.Counts()
	wpm := float64(max(0, c-t.lastCorrect)) * 12 / dt
	t.Samples = append(t.Samples, Sample{now.Sub(t.Started).Seconds(), wpm})
	t.lastSample = now
	t.lastCorrect = c
}
func (t *Tracker) Finish(now time.Time) {
	if t.Ended.IsZero() {
		t.Ended = now
		t.Sample(now, true)
	}
}

// Metrics uses retained correct characters for speed, attempts for accuracy.
func (t *Tracker) Metrics(now time.Time) Metrics {
	c, e := t.Counts()
	m := Metrics{Correct: c, Incorrect: e, Keystrokes: t.Attempts, Errors: t.Attempts - t.CorrectAttempts, Samples: append([]Sample(nil), t.Samples...)}
	if !t.Ended.IsZero() {
		now = t.Ended
	}
	if !t.Started.IsZero() {
		m.Duration = math.Max(0, now.Sub(t.Started).Seconds())
	}
	if m.Duration > 0 {
		m.WPM = float64(c) * 12 / m.Duration
		m.RawWPM = float64(t.Attempts) * 12 / m.Duration
	}
	if t.Attempts > 0 {
		m.Accuracy = 100 * float64(t.CorrectAttempts) / float64(t.Attempts)
	}
	rates := make([]float64, len(m.Samples))
	for i, s := range m.Samples {
		rates[i] = s.WPM
	}
	m.Consistency = Consistency(rates)
	return m
}

// Consistency measures variation in sampled speed.
func Consistency(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	mean := 0.0
	for _, v := range values {
		mean += v
	}
	mean /= float64(len(values))
	if mean <= 0 {
		return 0
	}
	variance := 0.0
	for _, v := range values {
		variance += (v - mean) * (v - mean)
	}
	variance /= float64(len(values))
	return math.Max(0, 100*(1-math.Sqrt(variance)/mean))
}
