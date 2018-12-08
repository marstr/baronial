package budget

import (
	"context"
	"testing"
	"time"
)

const (
	defaultTimeout = 30 * time.Second
)

func TestLoad(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	result, err := Load(ctx, "~/OneDrive/budget")
	if err != nil {
		t.Error(err)
		return
	}

	t.Log(result)
}
