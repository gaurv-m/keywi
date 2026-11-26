package app

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/gaurv-m/keywi/internal/storage"
	"github.com/gaurv-m/keywi/internal/typing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	"github.com/muesli/termenv"
)

func model(t *testing.T) Model {
	t.Helper()
	s, _, e := storage.OpenAt(filepath.Join(t.TempDir(), "config"), filepath.Join(t.TempDir(), "data"))
	if e != nil {
		t.Fatal(e)
	}
	return New(s, nil)
}
func key(m Model, k tea.KeyMsg) Model { next, _ := m.Update(k); return next.(Model) }
func TestAllScreensResize(t *testing.T) {
	m := model(t)
	m.start("textattack")
	for _, size := range [][2]int{{1, 1}, {10, 5}, {24, 10}, {40, 12}, {60, 20}, {80, 24}, {120, 40}} {
		next, _ := m.Update(tea.WindowSizeMsg{Width: size[0], Height: size[1]})
		m = next.(Model)
		for _, s := range []screen{menu, configuration, play, results, stats, history, detail, settings, helpScreen} {
			m.screen = s
			v := m.View()
			ls := strings.Split(v, "\n")
			if len(ls) > m.Height {
				t.Fatalf("screen %d too tall %d > %d", s, len(ls), m.Height)
			}
			for _, l := range ls {
				if ansi.StringWidth(l) > m.Width {
					t.Fatalf("screen %d too wide", s)
				}
			}
		}
	}
}
func TestPasteNavigationAndRapidTyping(t *testing.T) {
	m := model(t)
	m.start("textattack")
	m = key(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("pasted"), Paste: true})
	if len(m.blitzGame.Tracker.Input) != 0 {
		t.Fatal("paste accepted")
	}
	m = key(m, tea.KeyMsg{Type: tea.KeyLeft})
	if !m.blitzGame.Tracker.Started.IsZero() {
		t.Fatal("navigation started game")
	}
	target := append([]rune(nil), m.blitzGame.Tracker.Target...)
	m = key(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: target})
	if m.blitzGame.Sentences != 1 {
		t.Fatal("rapid input lost")
	}
	first := m.blitzGame.Tracker.Target[0]
	m = key(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{first}})
	if len(m.blitzGame.Tracker.Input) != 1 {
		t.Fatal("animation blocked input")
	}
}
func TestRestartAndAbandonCountdown(t *testing.T) {
	m := model(t)
	m.start("textattack")
	m = key(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	deadline := m.blitzGame.Deadline
	m = key(m, tea.KeyMsg{Type: tea.KeyTab})
	next, _ := m.Update(tickMsg(deadline.Add(time.Second)))
	m = next.(Model)
	if m.screen != results || len(m.Store.Data.Results) != 1 {
		t.Fatal("confirmation paused countdown")
	}
	m = key(m, tea.KeyMsg{Type: tea.KeyEnter})
	if m.screen != play || !m.blitzGame.Tracker.Started.IsZero() {
		t.Fatal("replay")
	}
	m = key(m, tea.KeyMsg{Type: tea.KeyTab})
	m = key(m, tea.KeyMsg{Type: tea.KeyEnter})
	if !m.blitzGame.Tracker.Started.IsZero() {
		t.Fatal("restart started timer")
	}
	m = key(m, tea.KeyMsg{Type: tea.KeyEsc})
	m = key(m, tea.KeyMsg{Type: tea.KeyEnter})
	if m.screen != play {
		t.Fatal("abandon default not safe")
	}
	m = key(m, tea.KeyMsg{Type: tea.KeyEsc})
	m = key(m, tea.KeyMsg{Type: tea.KeyRight})
	m = key(m, tea.KeyMsg{Type: tea.KeyEnter})
	if m.screen != menu {
		t.Fatal("abandon failed")
	}
}
func TestCompletionSaveAndReplay(t *testing.T) {
	m := model(t)
	m.config.Kind = "words"
	m.config.Amount = 10
	m.start("normal")
	m.casualGame.Tracker = typing.New("ab")
	m = key(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("ab")})
	if m.screen != results || len(m.Store.Data.Results) != 1 {
		t.Fatal("not persisted")
	}
	m = key(m, tea.KeyMsg{Type: tea.KeyEsc})
	if m.screen != menu {
		t.Fatal("results nav")
	}
	next, cmd := m.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
	if next.(Model).screen != menu || cmd == nil {
		t.Fatal("quit")
	}
	if _, ok := cmd().(tea.QuitMsg); !ok {
		t.Fatal("unclean quit")
	}
}
func TestMenusAndSettings(t *testing.T) {
	m := model(t)
	m = key(m, tea.KeyMsg{Type: tea.KeyRight})
	if m.selection != 1 {
		t.Fatal("menu navigation")
	}
	m = key(m, tea.KeyMsg{Type: tea.KeyDown})
	if m.selection != 2 {
		t.Fatal("down must select the next menu item")
	}
	m = key(m, tea.KeyMsg{Type: tea.KeyUp})
	if m.selection != 1 {
		t.Fatal("up must select the previous menu item")
	}
	m.screen = settings
	m.settingRow = 0
	m = key(m, tea.KeyMsg{Type: tea.KeyEnter})
	if m.Store.Preferences.Theme != "contrast" {
		t.Fatal("theme")
	}
	if _, e := os.Stat(filepath.Join(m.Store.ConfigDir, "preferences.json")); e != nil {
		t.Fatal(e)
	}
	m.settingRow = 8
	m = key(m, tea.KeyMsg{Type: tea.KeyEnter})
	if m.confirm != "reset" || m.confirmYes {
		t.Fatal("reset not explicit")
	}
}
func TestLogoCentered(t *testing.T) {
	m := model(t)
	m.Width = 100
	m.Height = 30
	view := ansi.Strip(m.View())
	for _, l := range strings.Split(view, "\n") {
		if strings.Contains(l, "█▄") && strings.Contains(l, "▄▀▀▄") {
			leading := len(l) - len(strings.TrimLeft(l, " "))
			if leading != 27 {
				t.Fatalf("logo horizontal position %d", leading)
			}
			return
		}
	}
	t.Fatal("logo absent")
}

func TestBlitzSentenceBlockCentered(t *testing.T) {
	m := model(t)
	m.start("textattack")
	for _, width := range []int{40, 60, 80, 100, 120} {
		m.Width = width
		m.Height = 24
		m.blitzGame.Tracker = typing.New("The waves carry a silver shell.")
		for _, line := range strings.Split(ansi.Strip(m.View()), "\n") {
			if strings.Contains(line, "The waves") {
				left := len(line) - len(strings.TrimLeft(line, " "))
				right := width - left - ansi.StringWidth(strings.TrimSpace(line))
				if left-right < -1 || left-right > 1 {
					t.Fatalf("sentence off centre at width %d: left %d right %d", width, left, right)
				}
			}
		}
	}
}
func TestMenuLegendSpacing(t *testing.T) {
	m := model(t)
	m.Width = 80
	m.Height = 24
	rows := strings.Split(ansi.Strip(m.View()), "\n")
	options, legend := -1, -1
	for i, line := range rows {
		if strings.Contains(line, "Settings") {
			options = i
		}
		if strings.Contains(line, "enter confirm") {
			legend = i
			if !strings.Contains(line, "select        enter") {
				t.Fatal("legend hints not separated")
			}
		}
	}
	if options < 0 || legend != m.Height-2 || legend-options < 3 {
		t.Fatalf("legend must sit at the bottom, apart from the menu: %d %d", options, legend)
	}
	for _, width := range []int{24, 40, 52, 53, 80, 81, 100, 101} {
		m.Width = width
		var original [4][2]int
		for selected := 0; selected < 4; selected++ {
			m.selection = selected
			var positions [4][2]int
			rows := strings.Split(ansi.Strip(m.View()), "\n")
			labels := []string{"Casual", "Blitz", "Statistics", "Settings"}
			for i, label := range labels {
				found := false
				for y, row := range rows {
					index := strings.Index(row, label)
					if index < 0 {
						continue
					}
					x := ansi.StringWidth(row[:index])
					positions[i] = [2]int{y, x}
					marker := ansi.Cut(row, x-2, x-1)
					if (marker == "›") != (i == selected) {
						t.Fatal("wrong menu selection marker")
					}
					found = true
				}
				if !found {
					t.Fatal("missing menu option", label)
				}
			}
			left, right := positions[0][1]-2, positions[2][1]+10
			if width >= 53 {
				for i := 1; i < 4; i++ {
					if positions[i][0] != positions[0][0] || positions[i][1]-positions[i-1][1]-len(labels[i-1]) != 6 {
						t.Fatal("horizontal menu lost its spacing")
					}
				}
				right = positions[3][1] + 8
				if !strings.Contains(rows[m.Height-2], "← →") {
					t.Fatal("horizontal controls missing")
				}
			} else {
				offsets := []int{0, 1, 3, 4}
				for i := range positions {
					if positions[i][1] != positions[0][1] || positions[i][0]-positions[0][0] != offsets[i] {
						t.Fatal("compact menu groups shifted")
					}
				}
			}
			if delta := left - (width - right); delta < -1 || delta > 1 {
				t.Fatal("whole menu is off center", width)
			}
			if selected == 0 {
				original = positions
			} else if positions != original {
				t.Fatal("selection moved menu labels")
			}
		}
	}
	m.Width = 80
	for _, height := range []int{21, 22, 23, 24, 25, 30} {
		m.Height = height
		rows := strings.Split(ansi.Strip(m.View()), "\n")
		logoEnd, menuRow := -1, -1
		for y, row := range rows {
			if strings.Contains(row, "█▀███▀") && strings.Contains(row, "▀▄") {
				logoEnd = y
			}
			if strings.Contains(row, "Statistics") {
				menuRow = y
			}
		}
		above := menuRow - logoEnd - 1
		below := height - 2 - menuRow - 1
		if logoEnd < 0 || menuRow < 0 || above < 3 || above-below < -1 || above-below > 1 {
			t.Fatal("logo/menu/footer spacing is uneven", height, above, below)
		}
		for _, row := range rows[logoEnd+1 : menuRow] {
			if strings.TrimSpace(row) != "" {
				t.Fatal("logo/menu gap is not clear")
			}
		}
	}

}
func TestThemeSelectionAndPersistence(t *testing.T) {
	m := model(t)
	m.screen = settings
	for _, want := range []string{"contrast", "rainbow", "fall", "tide", "fire", "orchid", "kiwi"} {
		m = key(m, tea.KeyMsg{Type: tea.KeyRight})
		if m.Store.Preferences.Theme != want {
			t.Fatalf("got %s want %s", m.Store.Preferences.Theme, want)
		}
		reloaded, _, err := storage.OpenAt(m.Store.ConfigDir, m.Store.DataDir)
		if err != nil || reloaded.Preferences.Theme != want {
			t.Fatalf("theme not saved: %v", err)
		}
	}
	m = key(m, tea.KeyMsg{Type: tea.KeyLeft})
	if m.Store.Preferences.Theme != "orchid" {
		t.Fatal("left did not cycle backwards")
	}
}

func TestCompletionFeedbackKeepsSentenceAndControlsStable(t *testing.T) {
	for _, height := range []int{10, 11, 12, 13, 14, 15, 24, 30} {
		for _, animate := range []bool{true, false} {
			m := model(t)
			m.Width = 80
			m.Height = height
			m.start("textattack")
			m.Store.Preferences.Animations = animate
			now := time.Now()
			target := append([]rune(nil), m.blitzGame.Tracker.Target...)
			for _, r := range target {
				m.blitzGame.Type(r, now)
			}
			if m.blitzGame.Sentences != 1 {
				t.Fatal("sentence did not complete")
			}
			next := string(m.blitzGame.Tracker.Target)
			sentenceRow, controlsRow := -1, -1
			for _, age := range []time.Duration{0, 70 * time.Millisecond, 180 * time.Millisecond, 300 * time.Millisecond, 800 * time.Millisecond} {
				m.now = now.Add(age)
				view := ansi.Strip(m.View())
				rows := strings.Split(view, "\n")
				if strings.Contains(view, "sentence complete") || strings.Contains(view, "+0.0s") {
					t.Fatal("unexpected completion message")
				}
				hasShards := strings.ContainsAny(view, "░▒⠂⠄▪")
				if animate && age == 70*time.Millisecond && !hasShards {
					t.Fatal("shatter effect missing")
				}
				// Controls contain a middle dot; shard glyphs must disappear.
				if (!animate || age >= 220*time.Millisecond) && strings.ContainsAny(view, "░▒⠂⠄▪") {
					t.Fatal("stale or disabled shards")
				}
				foundSentence, foundControls := -1, -1
				for i, line := range rows {
					if strings.Contains(line, next) {
						foundSentence = i
					}
					if strings.Contains(line, "tab then enter restart") {
						foundControls = i
					}
				}
				if foundSentence < 0 || foundControls < 0 {
					t.Fatalf("sentence or controls missing at height %d", height)
				}
				if sentenceRow >= 0 && (sentenceRow != foundSentence || controlsRow != foundControls) {
					t.Fatalf("layout jumped at height %d, age %s", height, age)
				}
				sentenceRow, controlsRow = foundSentence, foundControls
			}
			// The visual transition never gates or delays the engine's next input.
			m.blitzGame.Type(m.blitzGame.Tracker.Target[0], now.Add(10*time.Millisecond))
			if len(m.blitzGame.Tracker.Input) != 1 {
				t.Fatal("next sentence input blocked")
			}
		}
	}
}

func TestBlitzTimerGap(t *testing.T) {
	m := model(t)
	m.Width = 80
	m.Height = 24
	m.start("textattack")
	sentence := string(m.blitzGame.Tracker.Target)
	barRow, textRow := -1, -1
	for y, line := range strings.Split(ansi.Strip(m.View()), "\n") {
		if strings.Contains(line, "━━━") {
			barRow = y
		}
		if strings.Contains(line, sentence) {
			textRow = y
		}
	}
	if textRow-barRow != 4 {
		t.Fatalf("expected three blank lines after bar, got rows %d %d", barRow, textRow)
	}
}
func TestCapitalsSettingPersistsAndApplies(t *testing.T) {
	m := model(t)
	m.screen = settings
	m.config.Kind = "words"
	m.config.Amount = 10
	m.settingRow = 6
	if !m.Store.Preferences.Capitalization {
		t.Fatal("capitals not enabled by default")
	}
	m = key(m, tea.KeyMsg{Type: tea.KeyEnter})
	reloaded, _, err := storage.OpenAt(m.Store.ConfigDir, m.Store.DataDir)
	if err != nil || reloaded.Preferences.Capitalization {
		t.Fatal("capitals off was not saved")
	}
	m.start("normal")
	if m.casualGame.Config.Capitalization {
		t.Fatal("normal ignored setting")
	}
	m.screen = settings
	m = key(m, tea.KeyMsg{Type: tea.KeyEnter})
	m.start("normal")
	if !m.casualGame.Config.Capitalization {
		t.Fatal("normal did not enable capitals")
	}
}

func TestNavigationFooterAcrossAllScreens(t *testing.T) {
	m := model(t)
	m.start("textattack")
	for _, size := range [][2]int{{24, 10}, {40, 12}, {80, 24}, {100, 30}} {
		m.Width, m.Height = size[0], size[1]
		for _, s := range []screen{menu, configuration, play, results, stats, history, detail, settings, helpScreen} {
			m.screen = s
			rows := strings.Split(ansi.Strip(m.View()), "\n")
			if len(rows) != m.Height || strings.TrimSpace(rows[m.Height-2]) == "" || strings.TrimSpace(rows[m.Height-1]) != "" {
				t.Fatalf("screen %d footer is not anchored at %v", s, size)
			}
		}
		for _, confirm := range []string{"abandon", "restart", "reset", "quit"} {
			m.confirm = confirm
			rows := strings.Split(ansi.Strip(m.View()), "\n")
			if strings.TrimSpace(rows[m.Height-2]) == "" {
				t.Fatal("confirmation has no bottom controls")
			}
		}
		m.confirm = ""
	}
}
func TestModeNamesAndSetupExplanation(t *testing.T) {
	m := model(t)
	for _, mode := range []string{"normal", "textattack"} {
		m.start(mode)
		view := ansi.Strip(m.View())
		name := "Casual"
		if mode == "textattack" {
			name = "Blitz"
		}
		if !strings.Contains(view, name) || strings.Contains(view, "Normal") || strings.Contains(view, "Text Attack") {
			t.Fatal("outdated mode name")
		}
		rows := strings.Split(view, "\n")
		prompt := -1
		for i, row := range rows {
			if strings.Contains(row, "type to begin") {
				prompt = i
			}
		}
		if prompt < 0 || prompt >= m.Height-3 {
			t.Fatal("start prompt too low")
		}
	}
	m.screen = configuration
	m.config.Kind = "words"
	m.config.Amount = 25
	view := ansi.Strip(m.View())
	if !strings.Contains(view, "Finish 25 words.") || !strings.Contains(view, "[Words]") || strings.Contains(view, "WPM =") {
		t.Fatal("missing word-test explanation")
	}
	m.config.Kind = "time"
	m.config.Amount = 30
	view = ansi.Strip(m.View())
	if !strings.Contains(view, "[30s]") || strings.Contains(view, "Type for") {
		t.Fatal("time presets must use suffixes without a caption")
	}
}

func TestCasualShowsBothFormatsAndAllPresets(t *testing.T) {
	m := model(t)
	m.Width, m.Height = 80, 24
	m.screen = configuration
	view := ansi.Strip(m.View())
	for _, label := range []string{"[Time]", "Words", "15s", "[30s]", "60s", "120s"} {
		if !strings.Contains(view, label) {
			t.Fatal("missing time choice", label)
		}
	}
	if strings.Contains(view, "WPM =") {
		t.Fatal("formula clutter returned")
	}
	rowFor := func(view, label string) int {
		for y, row := range strings.Split(view, "\n") {
			if strings.Contains(row, label) {
				return y
			}
		}
		t.Fatal("missing layout row", label)
		return -1
	}
	if rowFor(view, "[Time]")-rowFor(view, "Casual") != 3 || rowFor(view, "15s")-rowFor(view, "[Time]") != 3 {
		t.Fatal("Casual sections need two blank rows between them")
	}
	if !strings.Contains(view, "[Time]        ") || !strings.Contains(view, "15s       [30s]") {
		t.Fatal("Casual options need generous horizontal spacing")
	}
	m = key(m, tea.KeyMsg{Type: tea.KeyRight})
	view = ansi.Strip(m.View())
	for _, label := range []string{"Time", "[Words]", "10", "[25]", "50", "100", "Finish 25 words."} {
		if !strings.Contains(view, label) {
			t.Fatal("missing word choice", label)
		}
	}
	m = key(m, tea.KeyMsg{Type: tea.KeyDown})
	presetRow := rowFor(view, "[25]")
	before := strings.Split(ansi.Strip(m.View()), "\n")[presetRow]
	m = key(m, tea.KeyMsg{Type: tea.KeyLeft})
	if m.config.Amount != 10 {
		t.Fatal("preset navigation")
	}
	after := strings.Split(ansi.Strip(m.View()), "\n")[presetRow]
	for _, value := range []string{"10", "25", "50", "100"} {
		if strings.Index(before, value) != strings.Index(after, value) {
			t.Fatal("Casual preset moved when selected", value)
		}
	}
}
func TestSettingsGroupingAndNavigationOrder(t *testing.T) {
	m := model(t)
	m.screen = settings
	view := ansi.Strip(m.View())
	for _, row := range settingsOrder {
		m.settingRow = row
		view += "\n" + ansi.Strip(m.View())
	}
	m.settingRow = 0
	for _, label := range []string{"Appearance", "Feedback", "Word tests", "Data"} {
		if !strings.Contains(view, label) {
			t.Fatal("missing settings group", label)
		}
	}
	for _, want := range []int{1, 2, 3, 7, 6, 4, 5, 8, 0} {
		m = key(m, tea.KeyMsg{Type: tea.KeyDown})
		if m.settingRow != want {
			t.Fatal("settings navigation selected wrong action", m.settingRow, want)
		}
	}
	m = key(m, tea.KeyMsg{Type: tea.KeyUp})
	if m.settingRow != 8 {
		t.Fatal("reverse wrap")
	}
	m.Store.Preferences.Capitalization = false
	m.Store.Preferences.Punctuation = false
	m.start("normal")
	if m.config.Text != "prose" || !m.config.Capitalization || !m.config.Punctuation {
		t.Fatal("word settings damaged authored prose")
	}
}

func TestSettingsLabelsAndValuesShareColumns(t *testing.T) {
	m := model(t)
	m.screen = settings
	m.Width = 80
	m.Height = 24
	labelColumn, valueColumn := -1, -1
	options := []string{"Theme", "Cursor", "Animations", "Live WPM", "Sound", "Capitals", "Punctuation", "Numbers"}
	for _, selected := range settingsOrder {
		m.settingRow = selected
		for _, line := range strings.Split(ansi.Strip(m.View()), "\n") {
			for _, label := range options {
				at := strings.Index(line, label)
				if at < 0 || strings.Contains(line, "Word tests only") {
					continue
				}
				after := line[at+len(label):]
				gap := len(after) - len(strings.TrimLeft(after, " "))
				val := ansi.StringWidth(line[:at+len(label)+gap])
				at = ansi.StringWidth(line[:at])
				if labelColumn < 0 {
					labelColumn, valueColumn = at, val
				}
				if at != labelColumn || val != valueColumn {
					t.Fatalf("columns shifted for %s: %d %d", label, at, val)
				}
				if gap < 8 || val-at != 19 {
					t.Fatal("settings need a wider horizontal gap")
				}
			}
		}
	}
}

func TestTitleEffectsRespectAnimationsSetting(t *testing.T) {
	t.Setenv("TERM", "xterm-256color")
	old, present := os.LookupEnv("NO_COLOR")
	os.Unsetenv("NO_COLOR")
	t.Cleanup(func() {
		if present {
			os.Setenv("NO_COLOR", old)
		}
	})
	renderer := lipgloss.DefaultRenderer()
	prior := renderer.ColorProfile()
	renderer.SetColorProfile(termenv.TrueColor)
	t.Cleanup(func() { renderer.SetColorProfile(prior) })
	m := model(t)
	for _, theme := range []string{"kiwi", "contrast", "rainbow", "fall", "tide", "fire", "orchid"} {
		m.Store.Preferences.Theme = theme
		for _, enabled := range []bool{false, true} {
			m.Store.Preferences.Animations = enabled
			m.now = time.Unix(1, 200000000)
			first := m.View()
			m.now = time.Unix(2, 800000000)
			second := m.View()
			if (first != second) != enabled {
				t.Fatal("title animation preference ignored", theme, enabled)
			}
			if ansi.Strip(first) != ansi.Strip(second) {
				t.Fatal("title effect moved menu text")
			}
		}
	}
	m.screen = configuration
	first := m.View()
	m.now = m.now.Add(time.Second)
	if m.View() != first {
		t.Fatal("title effect animated another screen")
	}
}

func TestMistakeLimitAppliesToBothModesAndBurstInput(t *testing.T) {
	for _, mode := range []string{"normal", "textattack"} {
		m := model(t)
		m.start(mode)
		m = key(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("§§§§§§")})
		if !m.tracker().Blocked() || len(m.tracker().Input) != 2 || m.tracker().Attempts != 2 {
			t.Fatal("burst input bypassed mistake limit", mode)
		}
		if strings.Contains(ansi.Strip(m.View()), "2 mistakes") {
			t.Fatal("unexpected correction instruction", mode)
		}
		m = key(m, tea.KeyMsg{Type: tea.KeyBackspace})
		if m.tracker().Blocked() {
			t.Fatal("backspace did not release mistake lock", mode)
		}
		correct := m.tracker().Target[len(m.tracker().Input)]
		m = key(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{correct}})
		if len(m.tracker().Input) != 2 || m.tracker().Uncorrected() != 1 {
			t.Fatal("corrected input could not resume", mode)
		}
	}
}

func TestEscapeExitConfirmation(t *testing.T) {
	m := model(t)
	m.selection = 2
	next, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = next.(Model)
	if cmd != nil || m.confirm != "quit" || m.confirmYes {
		t.Fatal("Esc quit without a default-safe confirmation")
	}
	view := ansi.Strip(m.View())
	if !strings.Contains(view, "Exit KeyWi?") || !strings.Contains(view, "[ Continue ]") {
		t.Fatal("exit confirmation did not show choices")
	}
	m.Width, m.Height = 24, 10
	view = ansi.Strip(m.View())
	if !strings.Contains(view, "[Continue]") || !strings.Contains(view, "Exit KeyWi?") {
		t.Fatal("narrow exit prompt lost its Continue choice")
	}
	m.Width, m.Height = 80, 24
	next, cmd = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = next.(Model)
	if cmd != nil || m.confirm != "" || m.screen != menu || m.selection != 2 {
		t.Fatal("default Enter did not keep the menu open")
	}
	m = key(m, tea.KeyMsg{Type: tea.KeyEsc})
	m = key(m, tea.KeyMsg{Type: tea.KeyRight})
	next, cmd = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = next.(Model)
	if cmd != nil || m.confirm != "" {
		t.Fatal("Esc did not cancel a selected Exit")
	}
	m = key(m, tea.KeyMsg{Type: tea.KeyEsc})
	if m.confirmYes {
		t.Fatal("reopening the prompt did not reset to Continue")
	}
	m = key(m, tea.KeyMsg{Type: tea.KeyRight})
	_, cmd = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("confirmed exit did not quit")
	}
	if _, ok := cmd().(tea.QuitMsg); !ok {
		t.Fatal("confirmed exit returned the wrong command")
	}
	m = model(t)
	m.screen = settings
	m = key(m, tea.KeyMsg{Type: tea.KeyEsc})
	if m.screen != menu || m.confirm != "" {
		t.Fatal("utility Esc stopped returning to the menu")
	}
}
