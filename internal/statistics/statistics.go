// Package statistics summarizes completed runs.
package statistics

import (
	"time"

	"github.com/gaurv-m/keywi/internal/casual"
	"github.com/gaurv-m/keywi/internal/typing"
)

type Result struct {
	ID        string         `json:"id"`
	At        time.Time      `json:"at"`
	Mode      string         `json:"mode"`
	Config    casual.Config  `json:"config"`
	Metrics   typing.Metrics `json:"metrics"`
	Sentences int            `json:"sentences,omitempty"`
	Level     int            `json:"level,omitempty"`
}
type Summary struct {
	Count                                                                        int
	AverageWPM, AverageAccuracy, AverageConsistency, TotalTime, AverageSentences float64
	BestWPM, LongestSurvival                                                     float64
	BestSentences, HighestLevel                                                  int
}

func Aggregate(results []Result, mode string) Summary {
	var s Summary
	for _, r := range results {
		if r.Mode != mode {
			continue
		}
		s.Count++
		s.AverageWPM += r.Metrics.WPM
		s.AverageAccuracy += r.Metrics.Accuracy
		s.AverageConsistency += r.Metrics.Consistency
		s.TotalTime += r.Metrics.Duration
		s.AverageSentences += float64(r.Sentences)
		s.BestWPM = max(s.BestWPM, r.Metrics.WPM)
		s.BestSentences = max(s.BestSentences, r.Sentences)
		s.HighestLevel = max(s.HighestLevel, r.Level)
		s.LongestSurvival = max(s.LongestSurvival, r.Metrics.Duration)
	}
	if s.Count > 0 {
		n := float64(s.Count)
		s.AverageWPM /= n
		s.AverageAccuracy /= n
		s.AverageConsistency /= n
		s.AverageSentences /= n
	}
	return s
}
func PersonalBest(results []Result, r Result) bool {
	found := false
	for _, old := range results {
		if old.Mode != r.Mode {
			continue
		}
		if r.Mode == "normal" {
			if old.Config.Key() != r.Config.Key() {
				continue
			}
			found = true
			if old.Metrics.WPM >= r.Metrics.WPM {
				return false
			}
		} else {
			found = true
			if old.Sentences >= r.Sentences {
				return false
			}
		}
	}
	if r.Mode == "normal" {
		return r.Metrics.WPM > 0 || !found
	}
	return r.Sentences > 0 || !found
}
func BestByConfig(results []Result) map[string]float64 {
	b := map[string]float64{}
	for _, r := range results {
		if r.Mode == "normal" {
			k := r.Config.Key()
			b[k] = max(b[k], r.Metrics.WPM)
		}
	}
	return b
}
