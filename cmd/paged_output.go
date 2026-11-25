package cmd

import (
	"errors"
	"io"
	"os"
	"os/exec"
	"sync"

	"github.com/spf13/cobra"
	"golang.org/x/term"
)

var errPagerClosed = errors.New("pager closed")

func setPagedCobraOutput(cmd *cobra.Command, _ []string) error {
	outFile, ok := cmd.OutOrStdout().(*os.File)
	if !ok {
		return nil
	}
	errFile, ok := cmd.ErrOrStderr().(*os.File)
	if !ok {
		return nil
	}
	output, err := NewPageWriteCloser(outFile, errFile)
	if err != nil {
		return err
	}
	cmd.SetOut(output)
	if run, ok := cmd.Context().Value(executionKey{}).(*execution); ok {
		run.cleanups = append(run.cleanups, func() error {
			cmd.SetOut(outFile)
			return output.Close()
		})
	} else {
		cmd.SetOut(outFile)
		return errors.Join(errors.New("pager requires command execution lifecycle"), output.Close())
	}
	return nil
}

// NewPageWriteCloser creates a stream writer that copies its output into a program like `less` or `more` to show output
// slowly, allowing a human to page through. Do remember to call `Close`, or you may mess up where Stdout is pointing in
// your shell.
func NewPageWriteCloser(outFile *os.File, errFile *os.File) (io.WriteCloser, error) {
	retval := &pageWriteCloser{}
	if term.IsTerminal(int(outFile.Fd())) {
		var err error
		if len(pagingPrograms) == 0 {
			return nil, errors.New("unrecognized platform, skipping paging")
		}

		for _, cmd := range pagingPrograms {
			if _, err = exec.LookPath(cmd.Path); err != nil {
				continue
			}

			retval.childProc = exec.Command(cmd.Path, cmd.Args...)
			retval.childProc.Stdout = outFile
			retval.childProc.Stderr = errFile
			retval.handle, err = retval.childProc.StdinPipe()
			if err != nil {
				return nil, err
			}
			break
		}
		if retval.handle == nil {
			retval.handle = outFile
		}
	} else {
		retval.handle = outFile
	}

	return retval, nil
}

type pageWriteCloser struct {
	procStart sync.Once
	handle    io.WriteCloser
	childProc *exec.Cmd
	started   bool
	startErr  error
}

func (pwc *pageWriteCloser) Write(p []byte) (int, error) {
	pwc.procStart.Do(func() {
		if pwc.childProc != nil {
			pwc.startErr = pwc.childProc.Start()
			pwc.started = pwc.startErr == nil
		}
	})
	if pwc.startErr != nil {
		return 0, pwc.startErr
	}
	n, err := pwc.handle.Write(p)
	if pwc.childProc != nil && isPagerClosedPipe(err) {
		return n, errPagerClosed
	}
	return n, err
}

func (pwc *pageWriteCloser) Close() error {
	if pwc.childProc == nil {
		return nil
	}
	if !pwc.started {
		return pwc.handle.Close()
	}
	err := pwc.handle.Close()
	return errors.Join(err, pwc.childProc.Wait())
}
