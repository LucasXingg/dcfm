package shell

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestDetectShellPrefersActiveShellOverLoginShell(t *testing.T) {
	lookup := func(name string) (string, error) {
		if name == "missing" {
			return "", os.ErrNotExist
		}
		if filepath.IsAbs(name) {
			return name, nil
		}
		return "/bin/" + name, nil
	}
	for _, test := range []struct {
		name, login, want, source string
		processes                 map[int]process
	}{
		{"nested bash", "/bin/zsh", "/bin/bash", "parent process", map[int]process{40: {30, "/bin/bash"}}},
		{"login shell name", "/bin/bash", "/bin/zsh", "parent process", map[int]process{40: {30, "-zsh"}}},
		{"wrapper", "/bin/zsh", "/bin/fish", "parent process", map[int]process{40: {30, "go"}, 30: {20, "fish"}}},
		{"process unavailable", "/bin/bash", "/bin/bash", "SHELL fallback", nil},
		{"no login shell", "", "/bin/sh", "POSIX fallback", nil},
		{"invalid login shell", "missing", "/bin/sh", "POSIX fallback", nil},
		{"self parent", "/bin/zsh", "/bin/zsh", "SHELL fallback", map[int]process{40: {40, "go"}}},
	} {
		t.Run(test.name, func(t *testing.T) {
			got, source := detectShell(40, test.login, func(pid int) (process, error) {
				if proc, ok := test.processes[pid]; ok {
					return proc, nil
				}
				return process{}, os.ErrNotExist
			}, lookup)
			if got != test.want || !strings.Contains(source, test.source) {
				t.Fatalf("shell=%s source=%s", got, source)
			}
		})
	}
}

func TestGetContextIgnoresStalePWD(t *testing.T) {
	dir, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	t.Chdir(dir)
	t.Setenv("PWD", "/stale/not-the-current-directory")
	got := GetContextWithLanguage("zh")
	if got.PWD != dir || got.OS != runtime.GOOS || got.Arch != runtime.GOARCH || got.Language != "zh" || got.Shell == "" || got.ShellSource == "" {
		t.Fatalf("incorrect context: %+v", got)
	}
}

func TestOSReleaseParsing(t *testing.T) {
	for input, want := range map[string]string{
		"NAME=Ubuntu\nPRETTY_NAME=\"Ubuntu 24.04.1 LTS\"\n": "Ubuntu 24.04.1 LTS",
		"PRETTY_NAME='Alpine Linux v3.21'":                  "Alpine Linux v3.21",
		"PRETTY_NAME=Fedora":                                "Fedora",
		"NAME=Ubuntu":                                       "",
	} {
		if got := parseOSRelease(input); got != want {
			t.Errorf("%q => %q, want %q", input, got, want)
		}
	}
}

func TestExecutionUsesCapturedShellAndWorkingDirectory(t *testing.T) {
	dir := t.TempDir()
	// A shell wrapper marks execution, proving the supplied shell is used even
	// when SHELL changes between generation and confirmation.
	shellPath := filepath.Join(dir, "captured-shell")
	if err := os.WriteFile(shellPath, []byte("#!/bin/sh\nprintf captured > used-shell\nexec /bin/sh \"$@\"\n"), 0700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("SHELL", "/does/not/exist")
	err := ExecuteWithContext(context.Background(), `printf '%s' "$PWD" > actual-pwd`, Context{Shell: shellPath, PWD: dir})
	if err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(filepath.Join(dir, "actual-pwd"))
	if err != nil {
		t.Fatal(err)
	}
	resolved, err := filepath.EvalSymlinks(string(got))
	if err != nil {
		t.Fatal(err)
	}
	want, err := filepath.EvalSymlinks(dir)
	if err != nil {
		t.Fatal(err)
	}
	if resolved != want {
		t.Errorf("ran in %q, want %q", resolved, want)
	}
	if data, err := os.ReadFile(filepath.Join(dir, "used-shell")); err != nil || string(data) != "captured" {
		t.Fatal("did not use captured shell")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := ExecuteWithContext(ctx, "true", Context{Shell: shellPath, PWD: dir}); !errors.Is(err, context.Canceled) {
		t.Fatalf("ignored cancellation: %v", err)
	}
}

func TestActiveShellDetectionInSubprocess(t *testing.T) {
	if os.Getenv("DCFM_TEST_SHELL_HELPER") == "1" {
		_ = json.NewEncoder(os.Stdout).Encode(GetContext())
		os.Exit(0)
	}
	for _, shellPath := range []string{"/bin/bash", "/bin/sh", "/bin/zsh"} {
		if _, err := os.Stat(shellPath); err != nil {
			continue
		}
		t.Run(shellPath, func(t *testing.T) {
			// The final ':' keeps the launching shell alive instead of letting it
			// optimize away the subprocess with exec. SHELL is intentionally stale.
			command := exec.Command(shellPath, "-c", `"$1" -test.run=^TestActiveShellDetectionInSubprocess$; :`, "dcfm-test", os.Args[0])
			command.Env = append(os.Environ(), "DCFM_TEST_SHELL_HELPER=1", "SHELL=/not/the/current/shell")
			output, err := command.Output()
			if err != nil {
				t.Fatal(err)
			}
			var got Context
			if err := json.Unmarshal(output, &got); err != nil {
				t.Fatalf("%v: %s", err, output)
			}
			if got.ShellSource != "parent process" {
				t.Fatalf("did not detect active shell: %+v", got)
			}
			// /bin/sh may be a symlink to dash or bash on Linux.
			want, err := filepath.EvalSymlinks(shellPath)
			if err != nil {
				t.Fatal(err)
			}
			actual, err := filepath.EvalSymlinks(got.Shell)
			if err != nil {
				t.Fatal(err)
			}
			if actual != want {
				t.Errorf("detected %s, want %s", actual, want)
			}
		})
	}
}
