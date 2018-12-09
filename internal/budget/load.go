package budget

import (
	"context"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/marstr/envelopes"
)

const (
	cashName = "cash.txt"
)

func Load(ctx context.Context, dirname string) (retval envelopes.Budget, err error) {
	var entries []os.FileInfo
	var children map[string]envelopes.Budget

	select {
	case <-ctx.Done():
		err = ctx.Err()
		return
	default:
		// Intentionally Left Blank
	}

	entries, err = ioutil.ReadDir(dirname)
	if err != nil {
		return
	}

	for _, e := range entries {
		fullEntryName := filepath.Join(dirname, e.Name())
		if e.IsDir() {
			var child envelopes.Budget

			if strings.HasPrefix(e.Name(), ".") {
				continue
			}

			child, err = Load(ctx, fullEntryName)
			if err != nil {
				return
			}

			if children == nil {
				children = make(map[string]envelopes.Budget)
			}
			children[e.Name()] = child
		} else if e.Name() == cashName {
			var reader io.Reader
			var contents []byte
			var bal envelopes.Balance

			reader, err = os.Open(fullEntryName)
			if err != nil {
				return
			}
			reader = io.LimitReader(reader, 2*1024)

			contents, err = ioutil.ReadAll(reader)
			if err != nil {
				return
			}

			trimmed := strings.TrimSpace(string(contents))
			bal, err = envelopes.ParseAmount(trimmed)
			if err != nil {
				return
			}

			retval = retval.WithBalance(bal)
		}
	}

	if children != nil {
		retval = retval.WithChildren(children)
	}

	return retval, nil
}
