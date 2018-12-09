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

package index

import (
	"context"
	"fmt"
	"github.com/marstr/envelopes"
	"os"
	"path/filepath"
)

// WriteBudget takes the memoized Budget and commits it to the current baronial index.
func WriteBudget(ctx context.Context, targetDir string, budget envelopes.Budget) error {
	targetFile := filepath.Join(targetDir, cashName)
	handle, err := os.Create(targetFile)
	if err != nil {
		return err
	}
	defer handle.Close()

	payload := envelopes.FormatAmount(budget.Balance())

	_, err = fmt.Fprintln(handle, payload)
	return nil
}
