// Copyright © 2019 Martin Strobel
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
	"io"
	"os"
	"path/filepath"
	"sort"

	"github.com/marstr/envelopes"
)

// WriteBudget takes the memoized Budget and commits it to the current baronial index.
func WriteBudget(ctx context.Context, targetDir string, budget envelopes.Budget) error {
	targetFile := filepath.Join(targetDir, cashName)
	handle, err := os.Create(targetFile)
	if err != nil {
		return err
	}
	defer handle.Close()

	return writeBalance(ctx, handle, budget.Balance)
}

func writeBalance(_ context.Context, output io.Writer, bal envelopes.Balance) error {
	var err error
	assetTypes := make([]string, 0, len(bal))

	for k := range bal {
		assetTypes = append(assetTypes, string(k))
	}

	sort.Strings(assetTypes)

	for i := range assetTypes {
		_, err = fmt.Fprintf(output, "%s %s\n", assetTypes[i], bal[envelopes.AssetType(assetTypes[i])].FloatString(3))
		if err != nil {
			return err
		}
	}
	return nil
}
