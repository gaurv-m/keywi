package storage

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/gaurv-m/keywi/internal/statistics"
	"github.com/gaurv-m/keywi/internal/typing"
)

func store(t *testing.T) *Store {
	t.Helper()
	root := t.TempDir()
	s, w, e := OpenAt(filepath.Join(root, "config"), filepath.Join(root, "data"))
	if e != nil || len(w) > 0 {
		t.Fatalf("%v %v", e, w)
	}
	return s
}

func TestFreshInstallUsesKiwiAndKeepsSavedTheme(t *testing.T) {
	s := store(t)
	if s.Preferences.Theme != "kiwi" {
		t.Fatal("fresh installation did not use Kiwi")
	}
	s.Preferences.Theme = "fire"
	if err := s.SavePreferences(); err != nil {
		t.Fatal(err)
	}
	reopened, _, err := OpenAt(s.ConfigDir, s.DataDir)
	if err != nil || reopened.Preferences.Theme != "fire" {
		t.Fatal("launch replaced a saved theme", err)
	}
}
func TestRoundTripAndReset(t *testing.T) {
	s := store(t)
	s.Preferences.Theme = "contrast"
	if e := s.SavePreferences(); e != nil {
		t.Fatal(e)
	}
	r := statistics.Result{ID: "one", Mode: "normal", At: time.Now().UTC(), Metrics: typing.Metrics{WPM: 70, Samples: []typing.Sample{{Seconds: 1, WPM: 70}}}}
	if e := s.Add(r); e != nil {
		t.Fatal(e)
	}
	if e := s.Add(r); e != nil {
		t.Fatal(e)
	}
	s2, w, e := OpenAt(s.ConfigDir, s.DataDir)
	if e != nil || len(w) != 0 || s2.Preferences.Theme != "contrast" || len(s2.Data.Results) != 1 || len(s2.Data.Results[0].Metrics.Samples) != 1 {
		t.Fatalf("%v %v %+v", e, w, s2)
	}
	backup, e := s2.Reset()
	if e != nil {
		t.Fatal(e)
	}
	b, e := os.ReadFile(backup)
	if e != nil || !strings.Contains(string(b), "one") {
		t.Fatal("reset lost backup")
	}
	if len(s2.Data.Results) != 0 {
		t.Fatal("reset")
	}
	st, _ := os.Stat(s.dataPath())
	if st.Mode().Perm() != 0600 {
		t.Fatal("permissions")
	}
}
func TestCorruptPreserved(t *testing.T) {
	for _, content := range []string{"{bad", "{\"version\":1,\"results\":[", "{}", "{\"version\":1} trailing"} {
		t.Run(content, func(t *testing.T) {
			s := store(t)
			os.WriteFile(s.dataPath(), []byte(content), 0600)
			s2, w, e := OpenAt(s.ConfigDir, s.DataDir)
			if e != nil || len(w) != 1 || len(s2.Data.Results) != 0 {
				t.Fatalf("%v %v", e, w)
			}
			paths, _ := filepath.Glob(s.dataPath() + ".corrupt-*")
			if len(paths) != 1 {
				t.Fatal("missing backup")
			}
			b, _ := os.ReadFile(paths[0])
			if string(b) != content {
				t.Fatal("modified backup")
			}
		})
	}
}
func TestFutureVersionUntouched(t *testing.T) {
	s := store(t)
	b := []byte(`{"version":999,"results":[]}`)
	os.WriteFile(s.dataPath(), b, 0600)
	_, _, e := OpenAt(s.ConfigDir, s.DataDir)
	if e == nil {
		t.Fatal("future version accepted")
	}
	after, _ := os.ReadFile(s.dataPath())
	if string(after) != string(b) {
		t.Fatal("future file modified")
	}
}
func TestFailedWriteDoesNotCommit(t *testing.T) {
	s := store(t)
	s.DataDir = filepath.Join(t.TempDir(), "missing")
	if s.Add(statistics.Result{ID: "x"}) == nil {
		t.Fatal("write succeeded")
	}
	if len(s.Data.Results) != 0 {
		t.Fatal("memory committed on failure")
	}
}
func TestLock(t *testing.T) {
	s := store(t)
	unlock, e := s.Lock()
	if e != nil {
		t.Fatal(e)
	}
	if _, e = s.Lock(); e == nil {
		t.Fatal("concurrent writer allowed")
	}
	unlock()
	unlock2, e := s.Lock()
	if e != nil {
		t.Fatal(e)
	}
	unlock2()
}
func TestPreferencesRecovery(t *testing.T) {
	s := store(t)
	os.WriteFile(s.prefPath(), []byte("nope"), 0600)
	s2, w, e := OpenAt(s.ConfigDir, s.DataDir)
	if e != nil || len(w) != 1 || s2.Preferences != Defaults() {
		t.Fatal(e, w)
	}
	s2.Preferences.Cursor = "invalid"
	s2.Preferences.Normalize()
	if s2.Preferences.Cursor != "block" {
		t.Fatal("normalization")
	}
	b, _ := json.Marshal(s2.Data)
	if !strings.Contains(string(b), `"version":1`) {
		t.Fatal("version")
	}
}

func TestLockedOpenSerializesReads(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	t.Setenv("XDG_DATA_HOME", t.TempDir())
	s, _, unlock, e := OpenLocked()
	if e != nil {
		t.Fatal(e)
	}
	if _, _, _, e = OpenLocked(); e == nil {
		unlock()
		t.Fatal("second locked reader admitted")
	}
	if e = s.Add(statistics.Result{ID: "last-writer", Mode: "normal"}); e != nil {
		unlock()
		t.Fatal(e)
	}
	unlock()
	s2, _, unlock2, e := OpenLocked()
	if e != nil {
		t.Fatal(e)
	}
	defer unlock2()
	if len(s2.Data.Results) != 1 || s2.Data.Results[0].ID != "last-writer" {
		t.Fatal("stale data loaded")
	}
}

func TestThemePreferenceMigration(t *testing.T) {
	if Defaults().Theme != "kiwi" {
		t.Fatal("Kiwi must be default")
	}
	for _, old := range []string{"tropical", "monochrome", "", "unknown"} {
		p := Defaults()
		p.Theme = old
		p.Normalize()
		if p.Theme != "kiwi" {
			t.Fatal("default migration")
		}
	}
	for _, name := range []string{"kiwi", "contrast", "rainbow", "fall", "tide", "fire", "orchid"} {
		p := Defaults()
		p.Theme = name
		p.Normalize()
		if p.Theme != name {
			t.Fatal("valid theme overwritten")
		}
	}
}

func TestLegacyPreferencesEnableCapitalization(t *testing.T) {
	s := store(t)
	if err := os.WriteFile(s.prefPath(), []byte(`{"version":1,"preferences":{"theme":"kiwi","cursor":"block"}}`), 0600); err != nil {
		t.Fatal(err)
	}
	migrated, _, err := OpenAt(s.ConfigDir, s.DataDir)
	if err != nil || !migrated.Preferences.Capitalization {
		t.Fatal("new setting missing from existing preferences", err)
	}
	migrated.Preferences.Capitalization = false
	if err = migrated.SavePreferences(); err != nil {
		t.Fatal(err)
	}
	again, _, err := OpenAt(s.ConfigDir, s.DataDir)
	if err != nil || again.Preferences.Capitalization {
		t.Fatal("explicit off was not preserved")
	}
}

func TestEmberMigratesToFall(t *testing.T) {
	s := store(t)
	s.Preferences.Theme = "ember"
	s.Preferences.Cursor = "line"
	if err := s.SavePreferences(); err != nil {
		t.Fatal(err)
	}
	migrated, _, err := OpenAt(s.ConfigDir, s.DataDir)
	if err != nil || migrated.Preferences.Theme != "fall" || migrated.Preferences.Cursor != "line" {
		t.Fatal("Ember did not migrate without changing other choices", err)
	}
	if err := migrated.SavePreferences(); err != nil {
		t.Fatal(err)
	}
	reopened, _, err := OpenAt(s.ConfigDir, s.DataDir)
	if err != nil || reopened.Preferences.Theme != "fall" {
		t.Fatal("Fall migration was not saved", err)
	}
}
