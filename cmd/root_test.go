package cmd

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/marstr/baronial/internal/diagnostics"
	"github.com/spf13/cobra"
)

func TestCLIOutputAndDiagnostics(t *testing.T) {
	binary := filepath.Join(t.TempDir(), "baronial")
	if runtime.GOOS == "windows" {
		binary += ".exe"
	}
	build := exec.Command("go", "build", "-o", binary, "..")
	if output, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build CLI: %v\n%s", err, output)
	}
	home := t.TempDir()
	run := func(dir string, args ...string) (string, string, error) {
		t.Helper()
		command := exec.Command(binary, args...)
		command.Dir = dir
		command.Env = append(os.Environ(), "HOME="+home, "USERPROFILE="+home,
			"XDG_STATE_HOME="+filepath.Join(home, "state"), "LOCALAPPDATA="+filepath.Join(home, "local"))
		var out, stderr bytes.Buffer
		command.Stdout, command.Stderr = &out, &stderr
		err := command.Run()
		return out.String(), stderr.String(), err
	}
	for _, name := range []string{"diff", "log", "balance", "credit"} {
		t.Run("failure-"+name, func(t *testing.T) {
			dir := t.TempDir()
			path := filepath.Join(dir, "diagnostics.log")
			stdout, stderr, err := run(dir, name, "--log-file", path)
			if err == nil || stdout != "" || stderr == "" {
				t.Fatalf("wrong failure output: stdout=%q stderr=%q err=%v", stdout, stderr, err)
			}
			if strings.Contains(stderr, "level=") || strings.Contains(stderr, "\"level\"") ||
				strings.Contains(stderr, "Usage:") || strings.Count(stderr, "\n") != 1 {
				t.Fatalf("non-plain or duplicated error: %q", stderr)
			}
			logs, err := os.ReadFile(path)
			if err != nil || strings.Count(string(logs), "command failed") != 1 {
				t.Fatalf("missing final diagnostic: %s, %v", logs, err)
			}
		})
	}
	for _, flag := range []string{
		"--log-level=typo", "--timeout=-1s", "--log-max-bytes=0", "--log-max-bytes=-1",
		"--log-max-bytes=not-a-size", "--log-max-bytes=9223372036854775808",
	} {
		dir := t.TempDir()
		stdout, stderr, err := run(dir, "init", flag)
		if err == nil || stdout != "" || stderr == "" {
			t.Fatalf("invalid flag accepted: %q, %q, %v", stdout, stderr, err)
		}
		if _, err := os.Stat(filepath.Join(dir, ".baronial")); !os.IsNotExist(err) {
			t.Fatalf("invalid flag caused side effects: %v", err)
		}
	}
	for _, level := range []string{"info", "debug", "off"} {
		t.Run("level-"+level, func(t *testing.T) {
			dir := t.TempDir()
			path := filepath.Join(dir, "logs", "diagnostics.log")
			args := []string{"init", "--log-file", path}
			if level != "info" {
				args = append(args, "--log-level", level)
			}
			stdout, stderr, err := run(dir, args...)
			if err != nil || stdout != "" || stderr != "" {
				t.Fatalf("init failed: %q, %q, %v", stdout, stderr, err)
			}
			logs, err := os.ReadFile(path)
			if level == "off" {
				if !os.IsNotExist(err) {
					t.Fatalf("off created logs: %s, %v", logs, err)
				}
				if _, err := os.Stat(filepath.Dir(path)); !os.IsNotExist(err) {
					t.Fatalf("off created log directory: %v", err)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			hasLibrary := strings.Contains(string(logs), "\"component\":\"envelopes\"")
			if hasLibrary != (level == "debug") {
				t.Fatalf("wrong envelopes filtering for %s: %s", level, logs)
			}
			if level == "info" && strings.Contains(string(logs), "\"level\":\"DEBUG\"") {
				t.Fatalf("default level is not info: %s", logs)
			}
		})
	}
	t.Run("configured-rotation-size", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "diagnostics.log")
		for n := 0; n < 2; n++ {
			stdout, stderr, err := run(dir, "init", "--log-file", path, "--log-max-bytes=512")
			if err != nil || stdout != "" || stderr != "" {
				t.Fatalf("configured size failed: %q, %q, %v", stdout, stderr, err)
			}
		}
		for _, file := range []string{path, path + ".1"} {
			info, err := os.Stat(file)
			if err != nil {
				t.Fatal(err)
			}
			if info.Size() > 512 {
				t.Fatalf("CLI rotation size ignored: %s is %d bytes", file, info.Size())
			}
		}
	})
	t.Run("unwritable-logs", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "not-a-log")
		if err := os.Mkdir(path, 0700); err != nil {
			t.Fatal(err)
		}
		stdout, stderr, err := run(dir, "init", "--log-file", path, "--log-level=debug")
		if err != nil || stdout != "" || strings.Count(stderr, "diagnostic logging unavailable") != 1 {
			t.Fatalf("logging failure affected command: %q, %q, %v", stdout, stderr, err)
		}
	})
	t.Run("balance-and-private-commit", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "diagnostics.log")
		mustRun := func(args ...string) string {
			t.Helper()
			args = append(args, "--log-file", path, "--log-level=debug")
			out, stderr, err := run(dir, args...)
			if err != nil || stderr != "" {
				t.Fatalf("command %v failed: %q, %v", args, stderr, err)
			}
			return out
		}
		mustRun("init")
		for _, name := range []string{"budget/grocery", "accounts/checking"} {
			if err := os.MkdirAll(filepath.Join(dir, name), 0700); err != nil {
				t.Fatal(err)
			}
		}
		mustRun("credit", "123.45", "budget/grocery", "accounts/checking")
		if out := mustRun("balance", "budget/grocery"); !strings.Contains(out, "123.45") {
			t.Fatalf("balance output changed: %q", out)
		}
		mustRun("commit", "--"+merchantFlag, "private-merchant", "--"+commentFlag, "private-comment",
			"--"+bankRecordIDFlag, "private-bank-id")
		id := strings.TrimSpace(mustRun("rev-parse", "HEAD"))
		out, stderr, err := run(dir, "checkout", id, "--log-level=off", "--log-file", path)
		if err != nil || out != "" || !strings.Contains(stderr, "can cause data loss") {
			t.Fatalf("off hid safety warning: %q, %q, %v", out, stderr, err)
		}
		_, stderr, err = run(dir, "credit", "private-invalid-amount", "budget/grocery", "--log-file", path, "--log-level=debug")
		if err == nil || !strings.Contains(stderr, "private-invalid-amount") {
			t.Fatalf("invalid amount not presented to user: %q, %v", stderr, err)
		}
		logs, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		for _, secret := range []string{"private-merchant", "private-comment", "private-bank-id", "private-invalid-amount", "123.45"} {
			if strings.Contains(string(logs), secret) {
				t.Fatalf("financial information %q leaked into diagnostics", secret)
			}
		}
	})
	t.Run("best-effort-transfer", func(t *testing.T) {
		dir := t.TempDir()
		if _, stderr, err := run(dir, "init", "--log-level=off"); err != nil {
			t.Fatalf("init: %q, %v", stderr, err)
		}
		src := filepath.Join(dir, "budget", "src")
		dest := filepath.Join(dir, "budget", "dest")
		if err := os.MkdirAll(filepath.Join(src, "cash.txt"), 0700); err != nil {
			t.Fatal(err)
		}
		if err := os.MkdirAll(dest, 0700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dest, "cash.txt"), []byte("USD 10\n"), 0600); err != nil {
			t.Fatal(err)
		}
		stdout, stderr, err := run(dir, "transfer", "1", src, dest, "--log-level=off")
		if err == nil || stdout != "" || stderr == "" {
			t.Fatalf("write failure not returned: %q, %q, %v", stdout, stderr, err)
		}
		contents, err := os.ReadFile(filepath.Join(dest, "cash.txt"))
		if err != nil || !strings.Contains(string(contents), "11.000") {
			t.Fatalf("independent destination write skipped: %q, %v", contents, err)
		}
	})
}

func TestExecutionCleanupAndFreshContext(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	root := &cobra.Command{Use: "test", SilenceErrors: true, SilenceUsage: true}
	root.PersistentFlags().String("log-level", "off", "")
	root.PersistentFlags().String("log-file", "", "")
	root.PersistentFlags().Int64(logMaxBytesFlag, int64(diagnostics.DefaultMaxBytes), "")
	root.PersistentFlags().Duration(timeoutFlag, 0, "")
	var stderr bytes.Buffer
	root.SetErr(&stderr)
	ctxs := make([]context.Context, 0, 2)
	cleaned := 0
	cmd := &cobra.Command{
		Use: "fail",
		PreRunE: func(cmd *cobra.Command, _ []string) error {
			run := cmd.Context().Value(executionKey{}).(*execution)
			run.cleanups = append(run.cleanups, func() error {
				cleaned++
				return errors.New("cleanup failed")
			})
			return nil
		},
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx, cancel, err := RootContext(cmd)
			if err != nil {
				return err
			}
			defer cancel()
			ctxs = append(ctxs, ctx)
			if err := ctx.Err(); err != nil {
				t.Fatalf("execution inherited canceled context: %v", err)
			}
			return errors.New("operation failed")
		},
	}
	root.AddCommand(cmd)
	root.SetArgs([]string{"fail"})
	for n := 0; n < 2; n++ {
		if err := executeCommand(root); err == nil {
			t.Fatal("failure swallowed")
		}
	}
	if cleaned != 2 || strings.Count(stderr.String(), "operation failed") != 2 ||
		strings.Count(stderr.String(), "cleanup failed") != 2 {
		t.Fatalf("cleanup or reporting failed: cleaned=%d, %q", cleaned, stderr.String())
	}
	for _, ctx := range ctxs {
		if !errors.Is(ctx.Err(), context.Canceled) {
			t.Fatalf("execution context not canceled: %v", ctx.Err())
		}
	}
}

func TestPagedFailureKeepsErrorsOnStderr(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	root := &cobra.Command{Use: "test", SilenceErrors: true, SilenceUsage: true}
	root.PersistentFlags().String("log-level", "off", "")
	root.PersistentFlags().String("log-file", "", "")
	root.PersistentFlags().Int64(logMaxBytesFlag, int64(diagnostics.DefaultMaxBytes), "")
	root.PersistentFlags().Duration(timeoutFlag, 0, "")
	out, err := os.CreateTemp(t.TempDir(), "stdout")
	if err != nil {
		t.Fatal(err)
	}
	defer out.Close()
	stderr, err := os.CreateTemp(t.TempDir(), "stderr")
	if err != nil {
		t.Fatal(err)
	}
	defer stderr.Close()
	root.SetOut(out)
	root.SetErr(stderr)
	cmd := &cobra.Command{
		Use: "paged", PreRunE: setPagedCobraOutput,
		RunE: func(cmd *cobra.Command, _ []string) error {
			if cmd.ErrOrStderr() != stderr {
				t.Fatal("pager redirected stderr")
			}
			return errors.New("failed before output")
		},
	}
	root.AddCommand(cmd)
	root.SetArgs([]string{"paged"})
	if err := executeCommand(root); err == nil {
		t.Fatal("command failure swallowed")
	}
	contents, err := os.ReadFile(stderr.Name())
	if err != nil || string(contents) != "failed before output\n" {
		t.Fatalf("error not on stderr: %q, %v", contents, err)
	}
	if contents, err := os.ReadFile(out.Name()); err != nil || len(contents) != 0 {
		t.Fatalf("error leaked into stdout: %q, %v", contents, err)
	}
	if _, err := out.WriteString("still open"); err != nil {
		t.Fatalf("error cleanup closed stdout: %v", err)
	}
}

func TestRootContextTimeout(t *testing.T) {
	cmd := &cobra.Command{Use: "test"}
	cmd.SetContext(context.Background())
	cmd.Flags().Duration(timeoutFlag, time.Nanosecond, "")
	ctx, cancel, err := RootContext(cmd)
	if err != nil {
		t.Fatal(err)
	}
	defer cancel()
	<-ctx.Done()
	if !errors.Is(ctx.Err(), context.DeadlineExceeded) {
		t.Fatalf("timeout not respected: %v", ctx.Err())
	}
}

func TestLaterLoggingFailureDoesNotFailCommand(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	path := filepath.Join(t.TempDir(), "diagnostics.log")
	root := &cobra.Command{Use: "test", SilenceErrors: true, SilenceUsage: true}
	root.PersistentFlags().String("log-level", "info", "")
	root.PersistentFlags().String("log-file", path, "")
	root.PersistentFlags().Int64(logMaxBytesFlag, int64(diagnostics.DefaultMaxBytes), "")
	root.PersistentFlags().Duration(timeoutFlag, 0, "")
	var stdout, stderr bytes.Buffer
	root.SetOut(&stdout)
	root.SetErr(&stderr)
	root.AddCommand(&cobra.Command{
		Use: "write",
		RunE: func(cmd *cobra.Command, _ []string) error {
			if err := os.Remove(path); err != nil {
				return err
			}
			if err := os.Mkdir(path, 0700); err != nil {
				return err
			}
			diagnostics.Logger(cmd.Context()).Info("diagnostic after storage failure")
			_, err := fmt.Fprintln(cmd.OutOrStdout(), "financial operation completed")
			return err
		},
	})
	root.SetArgs([]string{"write"})
	if err := executeCommand(root); err != nil {
		t.Fatalf("logging failure became command failure: %v", err)
	}
	if stdout.String() != "financial operation completed\n" ||
		strings.Count(stderr.String(), "diagnostic logging unavailable") != 1 {
		t.Fatalf("incorrect logging failure output: %q, %q", stdout.String(), stderr.String())
	}
}

func TestPagerCloseDoesNotCloseCallerOutput(t *testing.T) {
	file, err := os.CreateTemp(t.TempDir(), "output")
	if err != nil {
		t.Fatal(err)
	}

	defer file.Close()
	pager, err := NewPageWriteCloser(file, os.Stderr)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := io.WriteString(pager, "first\n"); err != nil {
		t.Fatal(err)
	}
	if err := pager.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := fmt.Fprintln(file, "still open"); err != nil {
		t.Fatalf("pager closed caller's output: %v", err)
	}
}

func TestPagerEarlyExit(t *testing.T) {
	if os.Getenv("BARONIAL_PAGER_TEST_CHILD") == "1" {
		if _, err := io.ReadFull(os.Stdin, make([]byte, 1)); err != nil {
			t.Fatal(err)
		}
		return
	}
	child := exec.Command(os.Args[0], "-test.run=^TestPagerEarlyExit$")
	child.Env = append(os.Environ(), "BARONIAL_PAGER_TEST_CHILD=1")
	pipe, err := child.StdinPipe()
	if err != nil {
		t.Fatal(err)
	}
	pager := &pageWriteCloser{childProc: child, handle: pipe}
	var writeErr error
	for n := 0; n < 100; n++ {
		if _, writeErr = pager.Write([]byte(strings.Repeat("x", 4096))); writeErr != nil {
			break
		}
	}
	if !errors.Is(writeErr, errPagerClosed) {
		t.Fatalf("early exit not recognized: %v", writeErr)
	}
	if err := pager.Close(); err != nil {
		t.Fatalf("successful pager exit reported as failure: %v", err)
	}
}
