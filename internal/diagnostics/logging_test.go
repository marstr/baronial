package diagnostics

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

func newTestWriter(t *testing.T, path string, limit int64) *rotatingWriter {
	t.Helper()
	writer := &rotatingWriter{
		path: path, maxBytes: limit, backups: 3, maxAge: 7 * 24 * time.Hour,
		warn: func(err error) { t.Errorf("unexpected logging failure: %v", err) },
	}
	if err := writer.initialize(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = writer.Close() })
	return writer
}

func TestJSONFilteringAndSensitiveErrors(t *testing.T) {
	for _, level := range []slog.Level{slog.LevelInfo, slog.LevelDebug} {
		t.Run(level.String(), func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "logs", "baronial.log")
			logger, closer := Open(path, level, int64(DefaultMaxBytes), func(err error) { t.Fatal(err) })
			defer closer.Close()
			logger.Debug("detail", "operation", "resolve")
			logger.Error("failure", "error", fmt.Errorf("private bank-id and USD 123.45: %w", os.ErrPermission))
			contents, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			if strings.Contains(string(contents), "private") || strings.Contains(string(contents), "123.45") {
				t.Fatalf("sensitive error leaked: %s", contents)
			}
			if strings.Contains(string(contents), "detail") != (level == slog.LevelDebug) {
				t.Fatalf("wrong filtering: %s", contents)
			}
			for _, line := range strings.Split(strings.TrimSpace(string(contents)), "\n") {
				var record map[string]interface{}
				if err := json.Unmarshal([]byte(line), &record); err != nil {
					t.Fatalf("invalid JSON record: %v", err)
				}
			}
			if runtime.GOOS != "windows" {
				for _, file := range []string{path, path + ".lock", path + ".age"} {
					info, err := os.Stat(file)
					if err != nil || info.Mode().Perm() != 0600 {
						t.Fatalf("unsafe permissions on %s: %v, %v", file, info, err)
					}
				}
				info, err := os.Stat(filepath.Dir(path))
				if err != nil || info.Mode().Perm() != 0700 {
					t.Fatalf("unsafe directory permissions: %v, %v", info, err)
				}
			}
		})
	}
}

func TestRotationBoundsAndRetention(t *testing.T) {
	path := filepath.Join(t.TempDir(), "diagnostics.log")
	writer := newTestWriter(t, path, 32)
	for n := 0; n < 12; n++ {
		if _, err := writer.Write([]byte(strings.Repeat("x", 20) + "\n")); err != nil {
			t.Fatal(err)
		}
	}
	var total int64
	for n := 0; n <= 3; n++ {
		file := path
		if n > 0 {
			file = writer.backup(n)
		}
		info, err := os.Stat(file)
		if err != nil {
			t.Fatal(err)
		}
		if info.Size() > 32 {
			t.Fatalf("oversized file: %v", info.Size())
		}
		total += info.Size()
	}
	if total > 4*32 {
		t.Fatalf("unbounded logs: %d", total)
	}
	unrelated := filepath.Join(filepath.Dir(path), "keep.txt")
	if err := os.WriteFile(unrelated, []byte("keep"), 0600); err != nil {
		t.Fatal(err)
	}
	old := time.Now().Add(-8 * 24 * time.Hour)
	for _, file := range []string{path + ".age", writer.backup(1)} {
		if err := os.Chtimes(file, old, old); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := writer.Write([]byte("fresh\n")); err != nil {
		t.Fatal(err)
	}
	contents, err := os.ReadFile(path)
	if err != nil || string(contents) != "fresh\n" {
		t.Fatalf("stale active content retained: %q, %v", contents, err)
	}
	if _, err := os.Stat(writer.backup(1)); !os.IsNotExist(err) {
		t.Fatalf("stale backup retained: %v", err)
	}
	if contents, err := os.ReadFile(unrelated); err != nil || string(contents) != "keep" {
		t.Fatalf("unrelated file modified: %q, %v", contents, err)
	}
}

func TestOpenUsesConfiguredSize(t *testing.T) {
	for _, limit := range []int64{256, 512} {
		t.Run(fmt.Sprint(limit), func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "diagnostics.log")
			logger, closer := Open(path, slog.LevelInfo, limit, func(err error) { t.Fatal(err) })
			defer closer.Close()
			for n := 0; n < 12; n++ {
				logger.Info(strings.Repeat("x", 80))
			}
			for n := 0; n <= 3; n++ {
				file := path
				if n > 0 {
					file += fmt.Sprintf(".%d", n)
				}
				info, err := os.Stat(file)
				if err != nil {
					t.Fatal(err)
				}
				if info.Size() > limit {
					t.Fatalf("configured limit %d ignored: %s is %d bytes", limit, file, info.Size())
				}
			}
		})
	}
}

func TestStaleActiveWithoutAgeMarker(t *testing.T) {
	path := filepath.Join(t.TempDir(), "diagnostics.log")
	if err := os.WriteFile(path, []byte("old\n"), 0600); err != nil {
		t.Fatal(err)
	}
	old := time.Now().Add(-8 * 24 * time.Hour)
	if err := os.Chtimes(path, old, old); err != nil {
		t.Fatal(err)
	}
	writer := newTestWriter(t, path, 100)
	if _, err := writer.Write([]byte("new\n")); err != nil {
		t.Fatal(err)
	}
	contents, err := os.ReadFile(path)
	if err != nil || string(contents) != "new\n" {
		t.Fatalf("stale active file retained: %q, %v", contents, err)
	}
}

func TestLoggingFailuresWarnOnce(t *testing.T) {
	for _, failure := range []string{"open", "write", "oversized", "lock"} {
		t.Run(failure, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "diagnostics.log")
			warnings := 0
			if failure == "open" {
				if err := os.Mkdir(path, 0700); err != nil {
					t.Fatal(err)
				}
			}
			writer := &rotatingWriter{
				path: path, maxBytes: 32, backups: 3, maxAge: time.Hour,
				warn: func(error) { warnings++ },
			}
			if err := writer.initialize(); err != nil {
				writer.fail(err)
			}
			defer writer.Close()
			if failure == "write" {
				if err := os.Mkdir(path, 0700); err != nil {
					t.Fatal(err)
				}
			}
			if failure == "lock" {
				blocker, err := os.OpenFile(path+".lock", os.O_RDWR, 0600)
				if err != nil {
					t.Fatal(err)
				}
				defer blocker.Close()
				locked, err := tryFileLock(blocker)
				if err != nil || !locked {
					t.Fatalf("acquire competing lock: %v, %v", locked, err)
				}
				defer func() {
					if err := unlockFile(blocker); err != nil {
						t.Errorf("release competing lock: %v", err)
					}
				}()
			}
			message := []byte("message\n")
			if failure == "oversized" {
				message = []byte(strings.Repeat("x", 33))
			}
			_, _ = writer.Write(message)
			_, _ = writer.Write(message)
			if warnings != 1 {
				t.Fatalf("expected one warning, got %d", warnings)
			}
		})
	}
}

func TestConcurrentProcesses(t *testing.T) {
	if path := os.Getenv("BARONIAL_ROTATION_TEST_PATH"); path != "" {
		writer := newTestWriter(t, path, 256)
		for n := 0; n < 10; n++ {
			record := fmt.Sprintf("%s:%02d\n", os.Getenv("BARONIAL_ROTATION_TEST_WORKER"), n)
			if _, err := writer.Write([]byte(record)); err != nil {
				t.Fatal(err)
			}
		}
		return
	}
	path := filepath.Join(t.TempDir(), "diagnostics.log")
	var children []*exec.Cmd
	for worker := 0; worker < 4; worker++ {
		cmd := exec.Command(os.Args[0], "-test.run=^TestConcurrentProcesses$")
		cmd.Env = append(os.Environ(), "BARONIAL_ROTATION_TEST_PATH="+path,
			fmt.Sprintf("BARONIAL_ROTATION_TEST_WORKER=worker-%d", worker))
		if err := cmd.Start(); err != nil {
			t.Fatal(err)
		}
		children = append(children, cmd)
	}
	for _, cmd := range children {
		if err := cmd.Wait(); err != nil {
			t.Fatal(err)
		}
	}
	seen := make(map[string]bool)
	for n := 0; n <= 3; n++ {
		file := path
		if n > 0 {
			file += fmt.Sprintf(".%d", n)
		}
		contents, err := os.ReadFile(file)
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			t.Fatal(err)
		}
		if len(contents) > 256 {
			t.Fatalf("oversized concurrent file: %d", len(contents))
		}
		for _, record := range strings.Split(strings.TrimSpace(string(contents)), "\n") {
			if seen[record] {
				t.Fatalf("duplicate record: %s", record)
			}
			seen[record] = true
		}
	}
	if len(seen) != 40 {
		t.Fatalf("lost concurrent records: got %d, want 40", len(seen))
	}
}

func TestDefaultPathAndDisabledLevel(t *testing.T) {
	level, enabled, err := ParseLevel("off")
	if err != nil || enabled || level != slog.LevelError {
		t.Fatalf("off did not disable diagnostics: %v, %v, %v", level, enabled, err)
	}
	if _, _, err := ParseLevel("typo"); err == nil {
		t.Fatal("invalid level accepted")
	}
	if runtime.GOOS == "linux" {
		base := t.TempDir()
		t.Setenv("XDG_STATE_HOME", base)
		got, err := DefaultPath()
		if err != nil || got != filepath.Join(base, "baronial", "baronial.log") {
			t.Fatalf("wrong state path: %s, %v", got, err)
		}
	}
	if got := ErrorKind(fmt.Errorf("private: %w", context.DeadlineExceeded)); got != "deadline exceeded" {
		t.Fatalf("unexpected error classification: %s", got)
	}
	if got := ErrorKind(errors.New("secret payload")); strings.Contains(got, "secret") {
		t.Fatalf("unsafe error classification: %s", got)
	}
}
