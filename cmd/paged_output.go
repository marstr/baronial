package cmd

import (
	"errors"
	"io"
	"os"
	"os/exec"
	"sync"

	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh/terminal"
)

func setPagedCobraOutput(cmd *cobra.Command, _ []string) error {
	output, err := NewPageWriteCloser(os.Stdout, os.Stderr)
	if err != nil {
		return err
	}
	cmd.SetOutput(output)
	cmd.PostRunE = func(cmd *cobra.Command, args []string) error {
		return output.Close()
	}
	return nil
}

// NewPageWriteCloser creates a stream writer that copies its output into a program like `less` or `more` to show output
// slowly, allowing a human to page through. Do remember to call `Close`, or you may mess up where Stdout is pointing in
// your shell.
func NewPageWriteCloser(outFile *os.File, errFile *os.File) (io.WriteCloser, error) {
	retval := &pageWriteCloser{}
	if terminal.IsTerminal(int(outFile.Fd())) {
		var err error
		if pagingPrograms == nil || len(pagingPrograms) == 0 {
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
	} else {
		retval.handle = outFile
	}

	return retval, nil
}

type pageWriteCloser struct {
	procStart sync.Once
	handle io.WriteCloser
	childProc *exec.Cmd
}

func (pwc *pageWriteCloser) Write(p []byte) (int, error) {
	var err error
	pwc.procStart.Do(func(){
		if pwc.childProc != nil{
			err = pwc.childProc.Start();
		}
	})
	if err != nil {
		return 0, err
	}
	return pwc.handle.Write(p)
}

func (pwc *pageWriteCloser) Close() error {
	err := pwc.handle.Close()
	if err != nil {
		return err
	}
	if pwc.childProc != nil {
		err = pwc.childProc.Wait()
		if err != nil {
			return err
		}
	}
	return nil
}