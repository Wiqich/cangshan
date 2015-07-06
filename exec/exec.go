package exec

import (
	"bytes"
	"fmt"
	"github.com/yangchenxing/cangshan/logging"
	"io"
	"os/exec"
	"strings"
	"time"
)

func Exec(path string, params []string, stdin []byte, timeout time.Duration) (bool, []byte, []byte, error) {
	var err error
	command := exec.Command(path, params...)
	commandText := path
	if len(params) > 0 {
		commandText += " " + strings.Join(params, " ")
	}
	logging.Debug("execute command: %s", commandText)
	var stdinWriter io.WriteCloser
	if stdin != nil {
		if stdinWriter, err = command.StdinPipe(); err != nil {
			return false, nil, nil, fmt.Errorf("get stdin pipe of command `%s` fail: %s",
				commandText, err.Error())
		}
	}
	stdout := &bytes.Buffer{}
	command.Stdout = stdout
	stderr := &bytes.Buffer{}
	command.Stderr = stderr
	if err := command.Start(); err != nil {
		return false, nil, nil, fmt.Errorf("start execution of command `%s` fail: %s",
			commandText, err.Error())
	}
	if stdin != nil {
		if _, err := stdinWriter.Write(stdin); err != nil {
			return false, nil, nil, fmt.Errorf("write stdin to execution of command `%s` fail: %s",
				commandText, err.Error())
		}
		stdinWriter.Close()
	}
	done := make(chan error, 1)
	go func() {
		done <- command.Wait()
	}()
	select {
	case <-time.After(timeout):
		if err := command.Process.Kill(); err != nil {
			return false, nil, nil, fmt.Errorf("kill timeout execution of command `%s` fail: %s",
				commandText, err.Error())
		}
		return false, nil, nil, fmt.Errorf("execute command `%s` timeout: %s",
			commandText, timeout.String())
	case err := <-done:
		if err != nil {
			return false, nil, nil, fmt.Errorf("execute command `%s` fail: %s",
				commandText, err.Error())
		}
	}
	return command.ProcessState.Success(), stdout.Bytes(), stderr.Bytes(), nil
}
