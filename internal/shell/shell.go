package shell

import (
	"os"
	"os/exec"
	"runtime"
)

type Context struct {
	OS       string
	Shell    string
	PWD      string
	Language string
}

func GetContext() Context {
	pwd, _ := os.Getwd()
	
	shell := os.Getenv("SHELL")
	if shell == "" {
		shell = "/bin/sh"
	}

	return Context{
		OS:       runtime.GOOS,
		Shell:    shell,
		PWD:      pwd,
		Language: "en",
	}
}

func GetContextWithLanguage(language string) Context {
	ctx := GetContext()
	ctx.Language = language
	return ctx
}

// Execute runs the given command string in the native shell.
func Execute(command string) error {
	ctx := GetContext()

	cmd := exec.Command(ctx.Shell, "-c", command)

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
