// Package storage saves local settings and history.
package storage

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/gaurv-m/keywi/internal/casual"
	"github.com/gaurv-m/keywi/internal/statistics"
	"github.com/gaurv-m/keywi/internal/theme"
)

const Version = 1

type Preferences struct {
	Theme          string `json:"theme"`
	Cursor         string `json:"cursor"`
	Animations     bool   `json:"animations"`
	LiveWPM        bool   `json:"live_wpm"`
	Punctuation    bool   `json:"punctuation"`
	Numbers        bool   `json:"numbers"`
	Capitalization bool   `json:"capitalization"`
	Sound          bool   `json:"sound"`
	CasualKind     string `json:"casual_kind,omitempty"`
	CasualAmount   int    `json:"casual_amount,omitempty"`
}

// Defaults applies only when no saved preference overrides it.
func Defaults() Preferences {
	return Preferences{Theme: "kiwi", Cursor: "block", Animations: true, LiveWPM: true, Capitalization: true, CasualKind: "time", CasualAmount: 30}
}
func (p *Preferences) Normalize() {
	p.Theme = theme.Normalize(p.Theme)
	if p.Cursor != "block" && p.Cursor != "underline" && p.Cursor != "line" {
		p.Cursor = "block"
	}
	if !(casual.Config{Kind: p.CasualKind, Amount: p.CasualAmount}).Valid() {
		p.CasualKind, p.CasualAmount = "time", 30
	}
}

type Data struct {
	Version int                 `json:"version"`
	Results []statistics.Result `json:"results"`
}
type preferenceFile struct {
	Version     int         `json:"version"`
	Preferences Preferences `json:"preferences"`
}
type Store struct {
	ConfigDir, DataDir string
	Preferences        Preferences
	Data               Data
}

func Open() (*Store, []string, error) {
	config, data, err := directories()
	if err != nil {
		return nil, nil, err
	}
	return OpenAt(config, data)
}

// OpenLocked locks before reading history.
func OpenLocked() (*Store, []string, func(), error) {
	config, data, err := directories()
	if err != nil {
		return nil, nil, nil, err
	}
	if err = os.MkdirAll(data, 0700); err != nil {
		return nil, nil, nil, err
	}
	unlock, err := (&Store{DataDir: data}).Lock()
	if err != nil {
		return nil, nil, nil, err
	}
	s, warnings, err := OpenAt(config, data)
	if err != nil {
		unlock()
		return nil, warnings, nil, err
	}
	return s, warnings, unlock, nil
}

func directories() (string, string, error) {
	config, err := os.UserConfigDir()
	if err != nil {
		return "", "", err
	}
	var data string
	if runtime.GOOS == "darwin" {
		data = filepath.Join(config, "keywi")
	} else {
		data = os.Getenv("XDG_DATA_HOME")
		if data == "" || !filepath.IsAbs(data) {
			home, e := os.UserHomeDir()
			if e != nil {
				return "", "", e
			}
			data = filepath.Join(home, ".local", "share")
		}
		data = filepath.Join(data, "keywi")
	}
	return filepath.Join(config, "keywi"), data, nil
}
func OpenAt(config, data string) (*Store, []string, error) {
	s := &Store{ConfigDir: config, DataDir: data, Preferences: Defaults(), Data: Data{Version: Version}}
	for _, p := range []string{config, data} {
		if err := os.MkdirAll(p, 0700); err != nil {
			return nil, nil, err
		}
	}
	var warnings []string
	p := preferenceFile{Version: Version, Preferences: Defaults()}
	for _, entry := range []struct {
		path  string
		value any
	}{{s.prefPath(), &p}, {s.dataPath(), &s.Data}} {
		err := read(entry.path, entry.value)
		if err == nil || errors.Is(err, os.ErrNotExist) {
			continue
		}
		if errors.Is(err, errUnsafeStorage) || errors.Is(err, errStorageLimit) {
			return nil, warnings, err
		}
		// Preserve every byte before recovery. Failure to preserve is a hard error.
		backup := entry.path + ".corrupt-" + time.Now().UTC().Format("20060102T150405.000000000")
		if e := os.Rename(entry.path, backup); e != nil {
			return nil, warnings, fmt.Errorf("cannot preserve corrupt data: %w", e)
		}
		warnings = append(warnings, fmt.Sprintf("Recovered unreadable JSON; original preserved at %s (%v)", backup, err))
		if entry.path == s.dataPath() {
			s.Data = Data{Version: Version}
		} else {
			p = preferenceFile{Version: Version, Preferences: Defaults()}
		}
	}
	if s.Data.Version != Version || p.Version != Version {
		return nil, warnings, errors.New("unsupported storage version; files left untouched (upgrade Keywi)")
	}
	s.Preferences = p.Preferences
	s.Preferences.Normalize()
	return s, warnings, nil
}
func read(path string, v any) error {
	raw, err := readFile(path, storageLimit(v))
	if err != nil {
		return err
	}
	var envelope map[string]json.RawMessage
	if err = json.Unmarshal(raw, &envelope); err != nil {
		return err
	}
	if _, ok := envelope["version"]; !ok {
		return errors.New("missing storage version")
	}
	d := json.NewDecoder(strings.NewReader(string(raw)))
	if err = d.Decode(v); err != nil {
		return err
	}
	var extra any
	if err = d.Decode(&extra); err != io.EOF {
		return errors.New("extra data after JSON document")
	}
	return nil
}
func (s *Store) prefPath() string { return filepath.Join(s.ConfigDir, "preferences.json") }
func (s *Store) dataPath() string { return filepath.Join(s.DataDir, "history.json") }
func atomic(path string, v any) error {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	b = append(b, '\n')
	if int64(len(b)) > storageLimit(v) {
		return fmt.Errorf("%w: %s", errStorageLimit, path)
	}
	f, err := os.CreateTemp(filepath.Dir(path), ".keywi-*")
	if err != nil {
		return err
	}
	name := f.Name()
	defer os.Remove(name)
	if err = f.Chmod(0600); err == nil {
		_, err = f.Write(b)
	}
	if err == nil {
		err = f.Sync()
	}
	closeErr := f.Close()
	if err == nil {
		err = closeErr
	}
	if err != nil {
		return err
	}
	if err = os.Rename(name, path); err != nil {
		return err
	}
	dir, err := os.Open(filepath.Dir(path))
	if err != nil {
		return err
	}
	defer dir.Close()
	return dir.Sync()
}
func (s *Store) SavePreferences() error {
	return atomic(s.prefPath(), preferenceFile{Version, s.Preferences})
}
func (s *Store) Add(r statistics.Result) error {
	for _, old := range s.Data.Results {
		if old.ID == r.ID {
			return nil
		}
	}
	// Update in memory only after successful commit, allowing retries on failure.
	next := Data{Version, append(append([]statistics.Result(nil), s.Data.Results...), r)}
	if err := atomic(s.dataPath(), next); err != nil {
		return err
	}
	s.Data = next
	return nil
}
func (s *Store) Reset() (string, error) {
	backup := ""
	if len(s.Data.Results) > 0 {
		backup = s.dataPath() + ".reset-" + time.Now().UTC().Format("20060102T150405.000000000")
		src, e := readFile(s.dataPath(), maxHistoryBytes)
		if e != nil {
			return "", e
		}
		f, e := os.OpenFile(backup, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
		if e != nil {
			return "", e
		}
		_, e = f.Write(src)
		if e == nil {
			e = f.Sync()
		}
		closeErr := f.Close()
		if e == nil {
			e = closeErr
		}
		if e != nil {
			return "", e
		}
	}
	if err := atomic(s.dataPath(), Data{Version: Version}); err != nil {
		return backup, err
	}
	s.Data = Data{Version: Version}
	return backup, nil
}
