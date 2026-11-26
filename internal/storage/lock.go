package storage

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"golang.org/x/sys/unix"
)

const lockFormat = "keywi-advisory-lock"

// Lock holds a kernel lock. Keep its inode in place.
func (s *Store) Lock() (func(), error) {
	path := filepath.Join(s.DataDir, "session.lock")
	f, err := openRegular(path, os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		return nil, fmt.Errorf("cannot open session lock: %w", err)
	}
	var info unix.Stat_t
	if err = unix.Fstat(int(f.Fd()), &info); err != nil || info.Nlink != 1 {
		_ = f.Close()
		return nil, fmt.Errorf("cannot use linked session lock: %w", errUnsafeStorage)
	}
	if err = unix.Flock(int(f.Fd()), unix.LOCK_EX|unix.LOCK_NB); err != nil {
		_ = f.Close()
		if errors.Is(err, unix.EWOULDBLOCK) || errors.Is(err, unix.EAGAIN) {
			return nil, fmt.Errorf("another Keywi instance is running; close it before starting a new one")
		}
		return nil, fmt.Errorf("cannot acquire session lock: %w", err)
	}
	ready := false
	defer func() {
		if !ready {
			_ = f.Close()
		}
	}()

	// Recover old PID locks only after the process exits.
	previous, err := io.ReadAll(io.LimitReader(f, 256))
	if err != nil {
		return nil, fmt.Errorf("cannot read session lock: %w", err)
	}
	if pid, parseErr := strconv.Atoi(strings.TrimSpace(string(previous))); parseErr == nil && pid > 0 {
		err = unix.Kill(pid, 0)
		if err == nil || errors.Is(err, unix.EPERM) {
			return nil, fmt.Errorf("another Keywi instance may be running (legacy PID %d); close it before starting a new one", pid)
		}
		if !errors.Is(err, unix.ESRCH) {
			return nil, fmt.Errorf("cannot check previous Keywi instance: %w", err)
		}
	}
	if err = f.Truncate(0); err == nil {
		_, err = f.Seek(0, io.SeekStart)
	}
	if err == nil {
		_, err = fmt.Fprintf(f, "%s\n%d\n", lockFormat, os.Getpid())
	}
	if err == nil {
		err = f.Sync()
	}
	if err != nil {
		return nil, fmt.Errorf("cannot write session lock: %w", err)
	}
	ready = true
	var once sync.Once
	return func() { once.Do(func() { _ = f.Close() }) }, nil
}
