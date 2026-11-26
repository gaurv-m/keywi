package storage

import (
	"errors"
	"fmt"
	"io"
	"os"

	"golang.org/x/sys/unix"
)

const (
	maxPreferencesBytes = 64 << 10
	maxHistoryBytes     = 64 << 20
)

var (
	errUnsafeStorage = errors.New("storage must be a regular file without symbolic links")
	errStorageLimit  = errors.New("storage file exceeds size limit; original left untouched")
)

func openRegular(path string, flags int, mode os.FileMode) (*os.File, error) {
	f, err := os.OpenFile(path, flags|unix.O_NOFOLLOW|unix.O_NONBLOCK, mode)
	if errors.Is(err, unix.ELOOP) {
		return nil, fmt.Errorf("%w: %s", errUnsafeStorage, path)
	}
	if err != nil {
		return nil, err
	}
	info, err := f.Stat()
	if err != nil {
		_ = f.Close()
		return nil, err
	}
	if !info.Mode().IsRegular() {
		_ = f.Close()
		return nil, fmt.Errorf("%w: %s", errUnsafeStorage, path)
	}
	return f, nil
}

func storageLimit(v any) int64 {
	switch v.(type) {
	case preferenceFile, *preferenceFile:
		return maxPreferencesBytes
	default:
		return maxHistoryBytes
	}
}

func readFile(path string, limit int64) ([]byte, error) {
	f, err := openRegular(path, os.O_RDONLY, 0)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	info, err := f.Stat()
	if err != nil {
		return nil, err
	}
	if info.Size() > limit {
		return nil, fmt.Errorf("%w (%d bytes): %s", errStorageLimit, limit, path)
	}
	data, err := io.ReadAll(io.LimitReader(f, limit+1))
	if err != nil {
		return nil, err
	}
	if int64(len(data)) > limit {
		return nil, fmt.Errorf("%w (%d bytes): %s", errStorageLimit, limit, path)
	}
	return data, nil
}
