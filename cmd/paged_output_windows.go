//go:build windows
// +build windows

package cmd

import (
	"errors"
	"os/exec"

	"golang.org/x/sys/windows"
)

var pagingPrograms = []exec.Cmd{
	{
		Path: "more",
	},
}

func isPagerClosedPipe(err error) bool {
	return errors.Is(err, windows.ERROR_BROKEN_PIPE) || errors.Is(err, windows.ERROR_NO_DATA)
}
