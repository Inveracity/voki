package local

// This package provides a local target for Voki, allowing users to run commands and copy files on the local machine.
//

import (
	"bytes"
	"context"
	"io"
	"os"
	"os/exec"

	"github.com/inveracity/voki/internal/targets"
)

// RunCommand executes a command on the local machine and returns the output.
func RunCommand(step targets.Step) (string, string, error) {
	ctx := context.Background()
	cmd := exec.CommandContext(ctx, step.Shell, "-c", step.Command)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		return "", stderr.String(), err
	}
	return stdout.String(), stderr.String(), nil
}

// CopyFile copies a file from the local machine to a specified destination.
func CopyFile(source, destination string) error {
	srcFile, err := os.Open(source)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	destFile, err := os.Create(destination)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, srcFile)
	if err != nil {
		return err
	}

	return nil
}
