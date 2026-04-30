package shell

import (
	"os"
	"os/exec"
	"runtime"
)

type Context struct {
	OS    string
	Shell string
	PWD   string
}

func GetContext() Context {
	pwd, _ := os.Getwd()
	
	shell := os.Getenv("SHELL")
	if runtime.GOOS == "windows" {
		shell = os.Getenv("COMSPEC")
		if shell == "" {
			shell = "cmd.exe"
		}
	} else if shell == "" {
		shell = "/bin/sh"
	}

	return Context{
		OS:    runtime.GOOS,
		Shell: shell,
		PWD:   pwd,
	}
}

// Execute runs the given command string in the native shell.
func Execute(command string) error {
	ctx := GetContext()

	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		// Try to handle powershell vs cmd
		// For simplicity, we use cmd.exe /c
		cmd = exec.Command("cmd", "/c", command)
	} else {
		cmd = exec.Command(ctx.Shell, "-c", command)
	}

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

// CheckCommand uses bash/cmd to just parse/check if the command looks superficially runnable.
// This is optional but can be useful.
func CheckCommand(command string) error {
	// Not fully implemented, just a stub
	return nil
}
