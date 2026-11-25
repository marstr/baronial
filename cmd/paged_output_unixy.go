//go:build darwin || linux
// +build darwin linux

package cmd

import (
	"errors"
	"os/exec"
	"syscall"
)

var pagingPrograms = []exec.Cmd{
	{
		Path: "more",
	},
}

func isPagerClosedPipe(err error) bool {
	return errors.Is(err, syscall.EPIPE)
}
