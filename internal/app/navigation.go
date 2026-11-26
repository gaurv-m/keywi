package app

import (
	"time"

	"github.com/gaurv-m/keywi/internal/ui"
)

// Fresh sections start at the top; back navigation keeps its context.
func (m *Model) openSection(target screen) {
	m.screen = target
	m.message = ""
	m.messageUntil = time.Time{}
	switch target {
	case configuration:
		m.configRow = 0
	case settings:
		m.settingRow = settingsOrder[0]
	case stats:
		m.statsTab = 0
		m.bestOffset = 0
		m.historyIndex = 0
	case history:
		m.historyIndex = 0
	}
}

func (m Model) bestMaxOffset() int {
	if m.statsTab != 0 {
		return 0
	}
	count := len(casualBestRows(m.Store.Data.Results))
	visible := casualBestVisible(count, ui.ContentWidth(m.Width), m.Height)
	if visible == 0 {
		return 0
	}
	return max(0, count-visible)
}
