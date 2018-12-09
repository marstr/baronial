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
	"testing"
	"time"

	"github.com/mitchellh/go-homedir"
)

const (
	//defaultTimeout = 30 * time.Second
	defaultTimeout = 4 * 24 * time.Hour
)

func TestLoadBudget(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	loc, err := homedir.Expand("~/OneDrive/finances/budget")
	if err != nil {
		t.Error(err)
		return
	}

	result, err := LoadBudget(ctx, loc)
	if err != nil {
		t.Error(err)
		return
	}

	t.Log(result)
}
