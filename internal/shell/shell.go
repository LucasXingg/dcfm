package shell

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"
)

type Context struct {
	OS          string `json:"os"`
	OSVersion   string `json:"os_version"`
	Arch        string `json:"arch"`
	Shell       string `json:"shell"`
	ShellSource string `json:"shell_source"`
	PWD         string `json:"pwd"`
	Language    string `json:"language"`
}

func GetContext() Context {
	pwd, err := os.Getwd()
	if err != nil {
		pwd = "unknown"
	}
	// SHELL is the login shell, not necessarily the shell that launched dcfm.
	shell, source := detectShell(os.Getppid(), os.Getenv("SHELL"), readProcess, exec.LookPath)

	return Context{
		OS:          runtime.GOOS,
		OSVersion:   osDescription(),
		Arch:        runtime.GOARCH,
		Shell:       shell,
		ShellSource: source,
		PWD:         pwd,
		Language:    "en",
	}
}

func GetContextWithLanguage(language string) Context {
	ctx := GetContext()
	ctx.Language = language
	return ctx
}

// Execute runs the given command string in the native shell.
func Execute(command string) error {
	return ExecuteWithContext(context.Background(), command, GetContext())
}

// ExecuteWithContext uses exactly the shell and directory shown to the model.
// Like the original Execute, this starts a non-interactive child shell; aliases
// and unexported variables from the calling shell are not inherited.
func ExecuteWithContext(ctx context.Context, command string, shellCtx Context) error {
	if shellCtx.PWD == "" || shellCtx.PWD == "unknown" {
		return fmt.Errorf("cannot determine working directory")
	}
	cmd := exec.CommandContext(ctx, shellCtx.Shell, "-c", command)
	cmd.Dir = shellCtx.PWD

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

type process struct {
	parent int
	name   string
}

func readProcess(pid int) (process, error) {
	output, err := systemOutput("ps", "-p", strconv.Itoa(pid), "-o", "ppid=", "-o", "comm=")
	if err != nil {
		return process{}, err
	}
	fields := strings.Fields(output)
	if len(fields) < 2 {
		return process{}, fmt.Errorf("missing process information")
	}
	parent, err := strconv.Atoi(fields[0])
	name := strings.TrimSpace(strings.TrimPrefix(output, fields[0]))
	// On Linux comm can be truncated. /proc also preserves the actual executable
	// when PATH contains another program with the same name.
	if executable, linkErr := os.Readlink(fmt.Sprintf("/proc/%d/exe", pid)); linkErr == nil {
		name = executable
	}
	return process{parent: parent, name: name}, err
}

func detectShell(pid int, loginShell string, inspect func(int) (process, error), lookPath func(string) (string, error)) (string, string) {
	for depth := 0; pid > 1 && depth < 8; depth++ {
		proc, err := inspect(pid)
		if err != nil {
			break
		}
		name := strings.TrimPrefix(proc.name, "-") // login shells may be named -zsh
		switch filepath.Base(name) {
		case "sh", "bash", "zsh", "fish", "dash", "ash", "ksh", "ksh93", "mksh", "csh", "tcsh", "nu", "elvish":
			if path, err := lookPath(name); err == nil {
				return path, "parent process"
			}
		}
		if proc.parent == pid {
			break
		}
		pid = proc.parent
	}
	if loginShell != "" {
		if path, err := lookPath(loginShell); err == nil {
			return path, "SHELL fallback (active shell unknown)"
		}
	}
	return "/bin/sh", "POSIX fallback (active shell unknown)"
}

func systemOutput(name string, args ...string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	output, err := exec.CommandContext(ctx, name, args...).Output()
	return strings.TrimSpace(string(output)), err
}

func osDescription() string {
	switch runtime.GOOS {
	case "darwin":
		if version, err := systemOutput("sw_vers", "-productVersion"); err == nil {
			return "macOS " + version
		}
	case "linux":
		for _, path := range []string{"/etc/os-release", "/usr/lib/os-release"} {
			if data, err := os.ReadFile(path); err == nil {
				if name := parseOSRelease(string(data)); name != "" {
					return name
				}
			}
		}
	}
	return "unknown"
}

func parseOSRelease(data string) string {
	for _, line := range strings.Split(data, "\n") {
		key, value, ok := strings.Cut(strings.TrimSpace(line), "=")
		if ok && key == "PRETTY_NAME" {
			return strings.Trim(strings.TrimSpace(value), "\"'")
		}
	}
	return ""
}

// CheckCommand uses bash/cmd to just parse/check if the command looks superficially runnable.
// This is optional but can be useful.
func CheckCommand(command string) error {
	// Not fully implemented, just a stub
	return nil
}
