//+build darwin linux

package cmd

import (
	"os/exec"
)

var pagingPrograms = []exec.Cmd{
	{
		Path: "less",
	},
	{
		Path: "more",
	},
}