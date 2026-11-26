package sound

import (
	"bytes"
	"encoding/binary"
	"os"
	"path/filepath"
	"testing"
)

func TestShatterIsSmallEightBitPCM(t *testing.T) {
	if string(shatter[:4]) != "RIFF" || string(shatter[8:12]) != "WAVE" || binary.LittleEndian.Uint16(shatter[34:36]) != 8 {
		t.Fatal("invalid 8-bit WAV")
	}
	if len(shatter) > 8192 || len(shatter) < 2000 {
		t.Fatal("unexpected effect size", len(shatter))
	}
	if bytes.Equal(shatter[44:], bytes.Repeat([]byte{128}, len(shatter)-44)) {
		t.Fatal("effect is silent")
	}
}

func TestPlaybackUsesBundledEffectAndCleansUp(t *testing.T) {
	dir := t.TempDir()
	output := filepath.Join(dir, "played.wav")
	t.Setenv("KEYWI_SOUND_TEST_OUTPUT", output)
	executable := filepath.Join(dir, "player")
	if err := os.WriteFile(executable, []byte("#!/bin/sh\ncp \"$1\" \"$KEYWI_SOUND_TEST_OUTPUT\"\nprintf '%s' \"$1\" > \"$KEYWI_SOUND_TEST_OUTPUT.path\"\n"), 0700); err != nil {
		t.Fatal(err)
	}
	p := New()
	defer p.Close()
	p.command, p.quiet = executable, false
	p.Play()
	played, err := os.ReadFile(output)
	if err != nil || !bytes.Equal(played, shatter) {
		t.Fatal("player did not receive the bundled sound", err)
	}
	path, err := os.ReadFile(output + ".path")
	if err != nil {
		t.Fatal(err)
	}
	if _, err = os.Stat(string(path)); !os.IsNotExist(err) {
		t.Fatal("temporary sound file remained", err)
	}
	p.Close()
	os.Remove(output)
	p.Play()
	if _, err = os.Stat(output); !os.IsNotExist(err) {
		t.Fatal("closed player still played")
	}
}
