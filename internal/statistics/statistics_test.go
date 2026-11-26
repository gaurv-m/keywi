package statistics

import (
	"testing"

	"github.com/gaurv-m/keywi/internal/casual"
	"github.com/gaurv-m/keywi/internal/typing"
)

func TestBestsSeparatedByModeAndConfig(t *testing.T) {
	base := Result{Mode: "normal", Config: casual.DefaultConfig(), Metrics: typing.Metrics{WPM: 80, Accuracy: 95, Duration: 30, Consistency: 90}}
	runs := []Result{base, {Mode: "textattack", Sentences: 10, Level: 3, Metrics: typing.Metrics{WPM: 100, Duration: 50, Accuracy: 99}}}
	if PersonalBest(runs, base) {
		t.Fatal("tie is best")
	}
	base.Config.Amount = 60
	if !PersonalBest(runs, base) {
		t.Fatal("config not separate")
	}
	if PersonalBest(runs, Result{Mode: "textattack", Sentences: 9}) {
		t.Fatal("lower attack score PB")
	}
	if !PersonalBest(runs, Result{Mode: "textattack", Sentences: 11}) {
		t.Fatal("higher attack score")
	}
	s := Aggregate(runs, "normal")
	if s.Count != 1 || s.AverageWPM != 80 || s.TotalTime != 30 || s.AverageConsistency != 90 {
		t.Fatal(s)
	}
	s = Aggregate(runs, "textattack")
	if s.BestSentences != 10 || s.HighestLevel != 3 || s.LongestSurvival != 50 || s.AverageSentences != 10 {
		t.Fatal(s)
	}
	if len(BestByConfig(runs)) != 1 {
		t.Fatal("config bests")
	}
}
