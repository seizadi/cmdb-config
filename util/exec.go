package util

import (
	"bytes"
	"os/exec"
)

func CmdExec(cmd string, path string) error {
	var out bytes.Buffer
	// TODO - Use Stream Buffer to do this
	echo_cmd := exec.Command("echo", cmd)
	echo_cmd.Stdout = &out
	err := echo_cmd.Run()
	if err != nil {
		return err
	}

	err = CopyBufferContents(out.Bytes(), path)
	if err != nil {
		return err
	}

	out.Reset()
	exec_cmd := exec.Command("bash", path)
	exec_cmd.Stdout = &out
	err = exec_cmd.Run()
	if err != nil {
		return err
	}

	return nil
}
