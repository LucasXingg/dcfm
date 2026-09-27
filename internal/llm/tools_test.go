package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"testing"

	"github.com/lucas/dcfm/internal/shell"
)

func testExplorer(t *testing.T) *explorer {
	t.Helper()
	e, err := newExplorer(shell.Context{PWD: t.TempDir(), OS: "linux", Shell: "/bin/bash"})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = e.Close() })
	return e
}

func callTool(t *testing.T, e *explorer, name string, args map[string]any) map[string]any {
	t.Helper()
	data, err := json.Marshal(args)
	if err != nil {
		t.Fatal(err)
	}
	var result map[string]any
	if err := json.Unmarshal([]byte(e.call(context.Background(), name, string(data))), &result); err != nil {
		t.Fatal(err)
	}
	return result
}

func writeFixture(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
		t.Fatal(err)
	}
}

func TestExplorerReadOnlyQueriesAndPagination(t *testing.T) {
	e := testExplorer(t)
	for i := 0; i < 105; i++ {
		writeFixture(t, filepath.Join(e.dir, fmt.Sprintf("file-%03d.txt", i)), "test")
	}
	first := callTool(t, e, "list_directory", map[string]any{})
	if first["truncated"] != true || first["next_offset"] != float64(100) || len(first["entries"].([]any)) != 100 {
		t.Fatalf("first page: %v", first)
	}
	second := callTool(t, e, "list_directory", map[string]any{"offset": 100})
	if second["truncated"] != false || len(second["entries"].([]any)) != 5 {
		t.Fatalf("second page: %v", second)
	}
	seen := make(map[string]bool)
	for _, page := range []map[string]any{first, second} {
		for _, item := range page["entries"].([]any) {
			name := item.(map[string]any)["name"].(string)
			if seen[name] {
				t.Errorf("repeated entry: %s", name)
			}
			seen[name] = true
		}
	}
	content := strings.Repeat("x", maxFileBytes-1) + "中文🙂" + strings.Repeat("y", 20)
	writeFixture(t, filepath.Join(e.dir, "unicode.txt"), content)
	first = callTool(t, e, "read_file", map[string]any{"path": "unicode.txt"})
	second = callTool(t, e, "read_file", map[string]any{"path": "unicode.txt", "offset": first["next_offset"]})
	if first["truncated"] != true || second["truncated"] != false || first["content"].(string)+second["content"].(string) != content {
		t.Fatalf("UTF-8 pagination lost data: %v, %v", first, second)
	}
	if got := callTool(t, e, "read_file", map[string]any{"path": "unicode.txt", "offset": len(content) + 1}); got["content"] != "" {
		t.Errorf("past EOF: %v", got)
	}
}

func TestExplorerRejectsEscapesCredentialsAndSpecialFiles(t *testing.T) {
	e := testExplorer(t)
	outside := t.TempDir()
	writeFixture(t, filepath.Join(outside, "secret"), "sensitive")
	if err := os.Symlink(outside, filepath.Join(e.dir, "escape")); err != nil {
		t.Fatal(err)
	}
	writeFixture(t, filepath.Join(e.dir, ".env"), "API_KEY=sensitive")
	if err := os.Symlink(".env", filepath.Join(e.dir, "innocent.txt")); err != nil {
		t.Fatal(err)
	}
	writeFixture(t, filepath.Join(e.dir, "binary"), "a\x00b")
	if err := syscall.Mkfifo(filepath.Join(e.dir, "pipe"), 0600); err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{outside, "../outside", "escape/secret", ".env", "innocent.txt", "binary", "pipe", ".", ".ssh/id_ed25519", ".config/dcfm/config.json"} {
		t.Run(path, func(t *testing.T) {
			result := callTool(t, e, "read_file", map[string]any{"path": path})
			if result["error"] == nil {
				t.Fatalf("unsafe read accepted: %v", result)
			}
			if strings.Contains(fmt.Sprint(result), "sensitive") {
				t.Fatal("leaked contents")
			}
		})
	}
	if result := callTool(t, e, "list_directory", map[string]any{"path": "escape"}); result["error"] == nil {
		t.Fatal("listed outside root")
	}
}

func TestExplorerLogicalPWDAndSymlinksWithinRoot(t *testing.T) {
	dir := t.TempDir()
	real := filepath.Join(dir, "real")
	alias := filepath.Join(dir, "alias")
	if err := os.Mkdir(real, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(real, alias); err != nil {
		t.Fatal(err)
	}
	writeFixture(t, filepath.Join(real, "data"), "contents")
	if err := os.Symlink("data", filepath.Join(real, "link")); err != nil {
		t.Fatal(err)
	}
	e, err := newExplorer(shell.Context{PWD: alias})
	if err != nil {
		t.Fatal(err)
	}
	defer e.Close()
	for _, path := range []string{"data", "link", filepath.Join(alias, "data"), filepath.Join(e.dir, "data")} {
		if result := callTool(t, e, "read_file", map[string]any{"path": path}); result["content"] != "contents" {
			t.Fatalf("%s: %v", path, result)
		}
	}
}

func TestExplorerRejectsMalformedArguments(t *testing.T) {
	e := testExplorer(t)
	for _, args := range []string{"", "null", "[]", "{} {}", `{"path":"x","command":"rm -rf ."}`, `{"offset":-1}`, `{"offset":1.2}`, `{"offset":10001}`, strings.Repeat("x", 4097)} {
		result := e.call(context.Background(), "list_directory", args)
		if !strings.Contains(result, `"error"`) {
			t.Errorf("accepted %q: %s", args, result)
		}
	}
	if got := e.call(context.Background(), "read_file", `{}`); !strings.Contains(got, `"error"`) {
		t.Fatal("accepted missing file path")
	}
	if got := e.call(context.Background(), "run_command", `{"command":"touch hacked"}`); !strings.Contains(got, "unknown tool") {
		t.Fatal(got)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if got := e.call(ctx, "list_directory", `{}`); !strings.Contains(got, "canceled") {
		t.Fatal(got)
	}
}

func TestExecutableLookupDoesNotRunProgramsOrLeakEnvironment(t *testing.T) {
	e := testExplorer(t)
	program := filepath.Join(e.dir, "example-program")
	marker := filepath.Join(e.dir, "executed")
	if err := os.WriteFile(program, []byte("#!/bin/sh\ntouch '"+marker+"'\n"), 0700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", e.dir)
	t.Setenv("DCFM_API_KEY", "do-not-send-this-key")
	got := callTool(t, e, "find_executable", map[string]any{"name": "example-program"})
	if got["found"] != true || got["path"] != program {
		t.Fatalf("lookup failed: %v", got)
	}
	if _, err := os.Stat(marker); !os.IsNotExist(err) {
		t.Fatal("lookup executed the program")
	}
	got = callTool(t, e, "find_executable", map[string]any{"name": "not-installed"})
	if got["found"] != false {
		t.Fatal(got)
	}
	for _, name := range []string{"", "sh -c ls", "/bin/sh", "../sh", "sh\nls"} {
		if got := callTool(t, e, "find_executable", map[string]any{"name": name}); got["error"] == nil {
			t.Fatalf("accepted %q", name)
		}
	}
	if strings.Contains(e.call(context.Background(), "get_environment", `{}`), "do-not-send-this-key") {
		t.Fatal("environment leaked API key")
	}
}

func TestExplorerProtectsItsOwnConfigAndLongDirectoryPages(t *testing.T) {
	e := testExplorer(t)
	configDir := filepath.Join(e.dir, "Library", "Application Support", "dcfm")
	if err := os.MkdirAll(configDir, 0700); err != nil {
		t.Fatal(err)
	}
	writeFixture(t, filepath.Join(configDir, "config.json"), `{"api_key":"private-key"}`)
	if got := callTool(t, e, "read_file", map[string]any{"path": filepath.Join(configDir, "config.json")}); got["error"] == nil {
		t.Fatal("read own API configuration")
	}
	configExplorer, err := newExplorer(shell.Context{PWD: configDir})
	if err != nil {
		t.Fatal(err)
	}
	defer configExplorer.Close()
	if got := callTool(t, configExplorer, "read_file", map[string]any{"path": "config.json"}); got["error"] == nil {
		t.Fatal("read config when starting inside config directory")
	}
	for i := 0; i < 100; i++ {
		writeFixture(t, filepath.Join(e.dir, fmt.Sprintf("%03d-%s", i, strings.Repeat("x", 240))), "")
	}
	seen := 0
	offset := float64(0)
	for {
		got := callTool(t, e, "list_directory", map[string]any{"offset": offset})
		if got["error"] != nil {
			t.Fatal(got)
		}
		seen += len(got["entries"].([]any))
		if got["truncated"] == false {
			break
		}
		if got["next_offset"].(float64) <= offset {
			t.Fatal("pagination did not advance")
		}
		offset = got["next_offset"].(float64)
	}
	if seen != 101 {
		t.Fatalf("listed %d entries, want 101", seen)
	}
}
