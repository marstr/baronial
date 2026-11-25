package diagnostics

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"syscall"
	"time"

	"github.com/marstr/units/data"
)

// DefaultMaxBytes is the maximum log-file size unless the caller overrides it.
const DefaultMaxBytes = 5 * data.Megabyte

type contextKey struct{}

var disabledLogger = slog.New(slog.NewTextHandler(io.Discard, nil))

// WithLogger associates diagnostics with one command invocation.
func WithLogger(ctx context.Context, logger *slog.Logger) context.Context {
	return context.WithValue(ctx, contextKey{}, logger)
}

// Logger returns the invocation's logger, or a silent logger for standalone helpers.
func Logger(ctx context.Context) *slog.Logger {
	if logger, ok := ctx.Value(contextKey{}).(*slog.Logger); ok && logger != nil {
		return logger
	}
	return disabledLogger
}

// ParseLevel validates a diagnostic level; off disables storage entirely.
func ParseLevel(value string) (slog.Level, bool, error) {
	switch strings.ToLower(value) {
	case "debug":
		return slog.LevelDebug, true, nil
	case "info":
		return slog.LevelInfo, true, nil
	case "warn":
		return slog.LevelWarn, true, nil
	case "error":
		return slog.LevelError, true, nil
	case "off":
		return slog.LevelError, false, nil
	default:
		return 0, false, fmt.Errorf("invalid log level %q: use debug, info, warn, error, or off", value)
	}
}

// DefaultPath selects per-user diagnostic storage, outside finance repositories.
func DefaultPath() (string, error) {
	if runtime.GOOS == "windows" {
		if base := os.Getenv("LOCALAPPDATA"); base != "" {
			return filepath.Join(base, "baronial", "logs", "baronial.log"), nil
		}
		return "", errors.New("LOCALAPPDATA is not set")
	}
	if runtime.GOOS != "darwin" {
		if base := os.Getenv("XDG_STATE_HOME"); filepath.IsAbs(base) {
			return filepath.Join(base, "baronial", "baronial.log"), nil
		}
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	if runtime.GOOS == "darwin" {
		return filepath.Join(home, "Library", "Logs", "baronial", "baronial.log"), nil
	}
	return filepath.Join(home, ".local", "state", "baronial", "baronial.log"), nil
}

// ErrorKind retains useful failure classification without serializing financial input.
func ErrorKind(err error) string {
	if err == nil {
		return ""
	}
	if joined, ok := err.(interface{ Unwrap() []error }); ok {
		var kinds []string
		for _, cause := range joined.Unwrap() {
			kinds = append(kinds, ErrorKind(cause))
		}
		return strings.Join(kinds, "; ")
	}
	for _, candidate := range []struct {
		err  error
		name string
	}{
		{context.Canceled, "canceled"},
		{context.DeadlineExceeded, "deadline exceeded"},
		{fs.ErrNotExist, "not found"},
		{fs.ErrPermission, "permission denied"},
		{fs.ErrExist, "already exists"},
		{fs.ErrInvalid, "invalid"},
		{io.ErrShortWrite, "short write"},
		{io.ErrUnexpectedEOF, "unexpected EOF"},
		{io.EOF, "EOF"},
	} {
		if errors.Is(err, candidate.err) {
			return candidate.name
		}
	}
	if code, ok := err.(syscall.Errno); ok {
		return code.Error()
	}
	if cause := errors.Unwrap(err); cause != nil {
		return reflect.TypeOf(err).String() + ": " + ErrorKind(cause)
	}
	return reflect.TypeOf(err).String()
}

func safeAttribute(_ []string, attr slog.Attr) slog.Attr {
	if attr.Key != "error" {
		return attr
	}
	err, ok := attr.Value.Any().(error)
	if !ok {
		return slog.String("error", "unclassified failure")
	}
	return slog.String("error", ErrorKind(err))
}

// Open creates a JSON logger. Storage failures warn once and disable diagnostics,
// without becoming failures of the financial command.
func Open(path string, level slog.Level, maxBytes int64, warn func(error)) (*slog.Logger, io.Closer) {
	writer := &rotatingWriter{
		path: path, maxBytes: maxBytes, backups: 3,
		maxAge: 7 * 24 * time.Hour, warn: warn,
	}
	if maxBytes <= 0 {
		writer.fail(errors.New("maximum diagnostic file size must be positive"))
	} else if err := writer.initialize(); err != nil {
		writer.fail(err)
	}
	logger := slog.New(slog.NewJSONHandler(writer, &slog.HandlerOptions{
		Level: level, ReplaceAttr: safeAttribute,
	}))
	return logger, writer
}
