package storage

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"time"
)

// Use a child to test exits without deferred cleanup.
func TestLockProcessHelper(t *testing.T) {
	dir := os.Getenv("KEYWI_LOCK_TEST_DIR")
	if dir == "" {
		return
	}
	if os.Getenv("KEYWI_LOCK_TEST_LEGACY") == "1" {
		if err := os.WriteFile(filepath.Join(dir, "session.lock"), []byte(fmt.Sprintln(os.Getpid())), 0600); err != nil {
			t.Fatal(err)
		}
	} else {
		unlock, err := (&Store{DataDir: dir}).Lock()
		if err != nil {
			t.Fatal(err)
		}
		defer unlock()
	}
	fmt.Println("locked")
	var input [1]byte
	_, _ = os.Stdin.Read(input[:])
}

func TestLockRecoversAfterProcessExit(t *testing.T) {
	for _, legacy := range []bool{false, true} {
		for _, ending := range []string{"close", "terminate", "kill"} {
			t.Run(fmt.Sprintf("legacy=%t/%s", legacy, ending), func(t *testing.T) {
				s := store(t)
				executable, err := os.Executable()
				if err != nil {
					t.Fatal(err)
				}
				cmd := exec.Command(executable, "-test.run=^TestLockProcessHelper$")
				legacyValue := "0"
				if legacy {
					legacyValue = "1"
				}
				cmd.Env = append(os.Environ(), "KEYWI_LOCK_TEST_DIR="+s.DataDir, "KEYWI_LOCK_TEST_LEGACY="+legacyValue)
				cmd.Stderr = os.Stderr
				stdout, err := cmd.StdoutPipe()
				if err != nil {
					t.Fatal(err)
				}
				stdin, err := cmd.StdinPipe()
				if err != nil {
					t.Fatal(err)
				}
				if err = cmd.Start(); err != nil {
					t.Fatal(err)
				}
				t.Cleanup(func() {
					stdin.Close()
					if cmd.ProcessState == nil {
						_ = cmd.Process.Kill()
						_ = cmd.Wait()
					}
				})
				ready := make(chan string, 1)
				go func() {
					line, _ := bufio.NewReader(stdout).ReadString('\n')
					ready <- line
				}()
				select {
				case line := <-ready:
					if line != "locked\n" {
						t.Fatal("child failed to acquire the lock", line)
					}
				case <-time.After(10 * time.Second):
					t.Fatal("child did not acquire the lock")
				}
				before, err := os.ReadFile(filepath.Join(s.DataDir, "session.lock"))
				if err != nil {
					t.Fatal(err)
				}
				if unlock, err := s.Lock(); err == nil {
					unlock()
					t.Fatal("live writer admitted a second process")
				}
				after, err := os.ReadFile(filepath.Join(s.DataDir, "session.lock"))
				if err != nil || string(before) != string(after) {
					t.Fatal("rejected writer changed the existing lock", err)
				}
				switch ending {
				case "close":
					err = stdin.Close()
				case "terminate":
					err = cmd.Process.Signal(syscall.SIGTERM)
				case "kill":
					err = cmd.Process.Kill()
				}
				if err != nil {
					t.Fatal(err)
				}
				waitErr := cmd.Wait()
				if ending == "close" && waitErr != nil {
					t.Fatal(waitErr)
				}
				unlock, err := s.Lock()
				if err != nil {
					t.Fatal("exited process left the next launch locked out", err)
				}
				defer unlock()
				contents, err := os.ReadFile(filepath.Join(s.DataDir, "session.lock"))
				if err != nil || !strings.HasPrefix(string(contents), lockFormat+"\n") {
					t.Fatal("did not migrate to an advisory lock", err)
				}
			})
		}
	}
}

func TestLockKeepsFileAndUnlockIsIdempotent(t *testing.T) {
	s := store(t)
	unlock, err := s.Lock()
	if err != nil {
		t.Fatal(err)
	}
	before, err := os.Stat(filepath.Join(s.DataDir, "session.lock"))
	if err != nil {
		t.Fatal(err)
	}
	unlock()
	unlock2, err := s.Lock()
	if err != nil {
		t.Fatal(err)
	}
	defer unlock2()
	unlock()
	if extra, err := s.Lock(); err == nil {
		extra()
		t.Fatal("old unlock released the next writer's lock")
	}
	after, err := os.Stat(filepath.Join(s.DataDir, "session.lock"))
	if err != nil || !os.SameFile(before, after) {
		t.Fatal("lock inode changed between sessions", err)
	}
}
