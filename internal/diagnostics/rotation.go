package diagnostics

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"
)

type rotatingWriter struct {
	mu       sync.Mutex
	path     string
	maxBytes int64
	backups  int
	maxAge   time.Duration
	now      func() time.Time
	warn     func(error)
	failed   bool
	closed   bool
	lock     *os.File
}

func (w *rotatingWriter) initialize() error {
	if w.now == nil {
		w.now = time.Now
	}
	path, err := filepath.Abs(w.path)
	if err != nil {
		return err
	}
	w.path = path
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}
	lockPath := path + ".lock"
	if err := checkRegular(lockPath); err != nil {
		return err
	}
	handle, err := os.OpenFile(lockPath, os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		return err
	}
	if err := handle.Chmod(0600); err != nil {
		return errors.Join(err, handle.Close())
	}
	w.lock = handle
	return w.locked(func() error { return w.prune() })
}

func checkRegular(path string) error {
	info, err := os.Lstat(path)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	if !info.Mode().IsRegular() {
		return fmt.Errorf("diagnostic destination %q is not a regular file", path)
	}
	return nil
}

func (w *rotatingWriter) locked(action func() error) (err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 250*time.Millisecond)
	defer cancel()
	timer := time.NewTicker(10 * time.Millisecond)
	defer timer.Stop()
	for {
		locked, err := tryFileLock(w.lock)
		if err != nil {
			return err
		}
		if locked {
			break
		}
		select {
		case <-ctx.Done():
			return fmt.Errorf("timed out locking diagnostic file: %w", ctx.Err())
		case <-timer.C:
		}
	}
	defer func() { err = errors.Join(err, unlockFile(w.lock)) }()
	return action()
}

func (w *rotatingWriter) backup(n int) string {
	return w.path + "." + strconv.Itoa(n)
}

func (w *rotatingWriter) prune() error {
	agePath := w.path + ".age"
	if err := checkRegular(agePath); err != nil {
		return err
	}
	age, err := os.Stat(agePath)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	for n := 0; n <= w.backups; n++ {
		path := w.path
		if n > 0 {
			path = w.backup(n)
		}
		if err := checkRegular(path); err != nil {
			return err
		}
		info, err := os.Stat(path)
		if os.IsNotExist(err) {
			if n == 0 {
				if err := os.Remove(agePath); err != nil && !os.IsNotExist(err) {
					return err
				}
			}
			continue
		}
		if err != nil {
			return err
		}
		oldest := info.ModTime()
		if n == 0 && age != nil && age.ModTime().Before(oldest) {
			oldest = age.ModTime()
		}
		if w.now().Sub(oldest) > w.maxAge || info.Size() > w.maxBytes {
			if err := os.Remove(path); err != nil {
				return err
			}
			if n == 0 {
				if err := os.Remove(agePath); err != nil && !os.IsNotExist(err) {
					return err
				}
			}
		} else if err := os.Chmod(path, 0600); err != nil {
			return err
		}
	}
	return nil
}

func (w *rotatingWriter) markAge() error {
	// Appending changes the log's mtime, so track its oldest content separately.
	path := w.path + ".age"
	if _, err := os.Stat(path); err == nil {
		return nil
	} else if !os.IsNotExist(err) {
		return err
	}
	info, err := os.Stat(w.path)
	if err != nil {
		return err
	}
	handle, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	if err := handle.Close(); err != nil {
		return err
	}
	return os.Chtimes(path, info.ModTime(), info.ModTime())
}

func (w *rotatingWriter) rotate() error {
	if err := os.Remove(w.backup(w.backups)); err != nil && !os.IsNotExist(err) {
		return err
	}
	for n := w.backups - 1; n >= 0; n-- {
		src := w.path
		if n > 0 {
			src = w.backup(n)
		}
		if err := os.Rename(src, w.backup(n+1)); err != nil && !os.IsNotExist(err) {
			return err
		}
	}
	age, err := os.Stat(w.path + ".age")
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	if age != nil {
		if err := os.Chtimes(w.backup(1), age.ModTime(), age.ModTime()); err != nil {
			return err
		}
	}
	if err := os.Remove(w.path + ".age"); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

func (w *rotatingWriter) fail(err error) {
	if !w.failed {
		w.failed = true
		if w.warn != nil {
			w.warn(err)
		}
	}
}

func (w *rotatingWriter) Write(data []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.failed || w.closed {
		return len(data), nil
	}
	if int64(len(data)) > w.maxBytes {
		err := errors.New("diagnostic record exceeds rotation size limit")
		w.fail(err)
		return 0, err
	}
	err := w.locked(func() (err error) {
		if err := w.prune(); err != nil {
			return err
		}
		info, err := os.Stat(w.path)
		if err != nil && !os.IsNotExist(err) {
			return err
		}
		if info != nil && int64(len(data)) > w.maxBytes-info.Size() {
			if err := w.markAge(); err != nil {
				return err
			}
			if err := w.rotate(); err != nil {
				return err
			}
		}
		handle, err := os.OpenFile(w.path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
		if err != nil {
			return err
		}
		defer func() { err = errors.Join(err, handle.Close()) }()
		if err := w.markAge(); err != nil {
			return err
		}
		n, err := handle.Write(data)
		if err == nil && n != len(data) {
			err = io.ErrShortWrite
		}
		return err
	})
	if err != nil {
		w.fail(err)
		return 0, err
	}
	return len(data), nil
}

func (w *rotatingWriter) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.closed {
		return nil
	}
	w.closed = true
	if w.lock != nil {
		if err := w.lock.Close(); err != nil {
			w.fail(err)
		}
	}
	return nil
}
