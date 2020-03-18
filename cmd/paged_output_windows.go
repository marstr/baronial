// +build windows

package cmd

import (
	"os/exec"
)

var pagingPrograms = []exec.Cmd{
	{
		Path: "more",
	},
}