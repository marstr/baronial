package cmd

import (
	"context"
	"errors"
	"io"
	"os"
	"os/exec"

	"golang.org/x/crypto/ssh/terminal"
)

var pagedOutput io.Writer = os.Stdout


func getPageWriter(ctx context.Context) (io.Writer, error) {
	if terminal.IsTerminal(int(os.Stdout.Fd())) {
		var err error

		if pagingPrograms == nil || len(pagingPrograms) == 0 {
			return os.Stdout, errors.New("unrecognized platform, skipping paging")
		}

		for _, cmd := range pagingPrograms {
			_, err = exec.LookPath(cmd.Path);
			if err != nil {
				continue
			}
			var retval io.Writer
			wrappedCmd := exec.CommandContext(ctx, cmd.Path, cmd.Args...)
			wrappedCmd.Stdout = os.Stdout
			retval, err = wrappedCmd.StdinPipe()
			if err != nil {
				return os.Stdout, err
			}
			wrappedCmd.Start()

			return retval, nil
		}

		return os.Stdout, errors.New("no paging programs found")
	}

	return os.Stdout, nil
}