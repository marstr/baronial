// Copyright © 2018 Martin Strobel
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

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
