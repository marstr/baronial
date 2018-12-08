package budget

import (
	"context"
	"github.com/mitchellh/go-homedir"
	"testing"
	"time"
)

const (
	//defaultTimeout = 30 * time.Second
	defaultTimeout = 4 * 24 * time.Hour
)

func TestLoad(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	loc, err := homedir.Expand("~/OneDrive/budget")
	if err != nil {
		t.Error(err)
		return
	}

	result, err := Load(ctx, loc)
	if err != nil {
		t.Error(err)
		return
	}

	t.Log(result)
}
