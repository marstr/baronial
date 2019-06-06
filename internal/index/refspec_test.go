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

package index_test

import (
	"context"
	"io/ioutil"
	"os"
	"path"
	"testing"
	"time"

	"github.com/marstr/envelopes"
	"github.com/marstr/envelopes/persist"

	"github.com/marstr/baronial/internal/index"
)

func TestRefSpec_Transaction(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5 * time.Second)
	defer cancel()

	temp_dir, err := ioutil.TempDir("", "baronial_TestRefSpec_Transaction_")
	if err != nil {
		t.Error(err)
		return
	}
	defer os.RemoveAll(temp_dir)

	repo_loc := path.Join(temp_dir, index.RepoName)

	err = os.Mkdir(repo_loc, os.ModePerm)
	if err != nil {
		t.Error(err)
		return
	}

	t1 := envelopes.Transaction{
		Comment: "baronial1",
	}

	t2 := envelopes.Transaction{
		Comment: "baronial2",
		Parent: t1.ID(),
	}

	t3 := envelopes.Transaction{
		Comment: "baronial3",
		Parent: t2.ID(),
	}

	toWrite := []*envelopes.Transaction{&t1, &t2, &t3}

	fs := persist.FileSystem{Root: repo_loc}
	writer := persist.DefaultWriter{Stasher: fs}

	for _, entry := range toWrite {
		err = writer.Write(ctx, entry)
		if err != nil {
			t.Error(err)
			return
		}
	}

	testCases := []struct{
		text index.RefSpec
		expected envelopes.ID
	}{
		{index.RefSpec(t1.ID().String()), t1.ID()},
		{index.RefSpec(t2.ID().String() + `^`), t1.ID()},
		{index.RefSpec(t3.ID().String() + `~2`), t1.ID()},
		{index.RefSpec(t3.ID().String() + `^^`), t1.ID()},
		{index.RefSpec(t3.ID().String() + `~1~1`), t1.ID()},
	}

	for _, tc := range testCases {
		got, err := tc.text.Transaction(ctx, temp_dir)
		if err != nil {
			t.Error(err)
			continue
		}

		if !got.Equal(tc.expected) {
			t.Logf("%q\n\tgot:  %s\n\twant: %s", string(tc.text), got.String(), tc.expected.String())
			t.Fail()
		}
	}
}
