package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"unicode/utf8"

	"github.com/lucas/dcfm/internal/shell"
	"github.com/sashabaranov/go-openai"
)

const maxFileBytes = 4096
const maxDirectoryEntries = 100
const maxToolResultBytes = 16384

func explorationTools() []openai.Tool {
	stringParam := func(description string) map[string]any {
		return map[string]any{"type": "string", "description": description}
	}
	offset := map[string]any{"type": "integer", "minimum": 0, "description": "Starting offset, default 0. Use next_offset from the previous result."}
	tool := func(name, description string, properties map[string]any, required []string) openai.Tool {
		return openai.Tool{Type: openai.ToolTypeFunction, Function: &openai.FunctionDefinition{
			Name: name, Description: description,
			Parameters: map[string]any{"type": "object", "properties": properties, "required": required, "additionalProperties": false},
		}}
	}
	return []openai.Tool{
		tool("get_environment", "Get detected OS, version, architecture, shell and working directory. Does not expose environment variable values.", map[string]any{}, []string{}),
		tool("list_directory", "List up to 100 entries in a directory under the starting working directory, including hidden entries. Never executes commands.", map[string]any{"path": stringParam("Relative or absolute directory path; defaults to ."), "offset": offset}, []string{}),
		tool("read_file", "Read up to 4096 bytes from a regular UTF-8 text file under the starting working directory. Common credential paths are blocked. Offset is in bytes.", map[string]any{"path": stringParam("Relative or absolute file path"), "offset": offset}, []string{"path"}),
		tool("find_executable", "Locate one executable by name on inherited PATH without running it. Does not detect aliases, functions or shell builtins.", map[string]any{"name": stringParam("Program name, such as git, python3, rg, or npm; no arguments or paths")}, []string{"name"}),
	}
}

type explorer struct {
	root        *os.Root
	dir         string
	environment shell.Context
}

func newExplorer(environment shell.Context) (*explorer, error) {
	if environment.PWD == "" || environment.PWD == "unknown" {
		return nil, fmt.Errorf("cannot determine working directory")
	}
	dir, err := filepath.EvalSymlinks(environment.PWD)
	if err != nil {
		return nil, err
	}
	dir, err = filepath.Abs(dir)
	if err != nil {
		return nil, err
	}
	root, err := os.OpenRoot(dir)
	if err != nil {
		return nil, err
	}
	return &explorer{root: root, dir: dir, environment: environment}, nil
}

func (e *explorer) Close() error { return e.root.Close() }

func (e *explorer) call(ctx context.Context, name, arguments string) string {
	result, err := e.run(ctx, name, arguments)
	if err != nil {
		result = map[string]string{"error": err.Error()}
	}
	data, err := json.Marshal(result)
	if err != nil || len(data) > maxToolResultBytes {
		return `{"error":"Tool output too large; request a narrower path or a later offset."}`
	}
	return string(data)
}

func decodeArguments(arguments string, target any) error {
	if len(arguments) > 4096 {
		return fmt.Errorf("tool arguments exceed 4096 bytes")
	}
	if !strings.HasPrefix(strings.TrimSpace(arguments), "{") {
		return fmt.Errorf("tool arguments must be a JSON object")
	}
	decoder := json.NewDecoder(strings.NewReader(arguments))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return fmt.Errorf("invalid tool arguments: %w", err)
	}
	if decoder.Decode(new(any)) != io.EOF {
		return fmt.Errorf("tool arguments must contain exactly one JSON object")
	}
	return nil
}

func (e *explorer) run(ctx context.Context, name, arguments string) (any, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	switch name {
	case "get_environment":
		if err := decodeArguments(arguments, &struct{}{}); err != nil {
			return nil, err
		}
		return e.environment, nil
	case "find_executable":
		var args struct {
			Name string `json:"name"`
		}
		if err := decodeArguments(arguments, &args); err != nil {
			return nil, err
		}
		if args.Name == "" || strings.ContainsAny(args.Name, "/\\ \t\r\n\x00") {
			return nil, fmt.Errorf("provide a program name without arguments or paths")
		}
		path, err := exec.LookPath(args.Name)
		if err != nil && !errors.Is(err, exec.ErrDot) {
			return map[string]any{"name": args.Name, "found": false}, nil
		}
		if !filepath.IsAbs(path) {
			path = filepath.Join(e.dir, path)
		}
		return map[string]any{"name": args.Name, "found": true, "path": path}, nil
	case "list_directory", "read_file":
		var args struct {
			Path   string `json:"path"`
			Offset int64  `json:"offset"`
		}
		if err := decodeArguments(arguments, &args); err != nil {
			return nil, err
		}
		if args.Offset < 0 {
			return nil, fmt.Errorf("offset must be non-negative")
		}
		if args.Path == "" {
			if name == "read_file" {
				return nil, fmt.Errorf("path is required")
			}
			args.Path = "."
		}
		path, err := e.resolvePath(args.Path)
		if err != nil {
			return nil, err
		}
		// Check before opening so FIFOs/devices cannot block a read. os.Root
		// additionally prevents traversal and symlink escapes at open time.
		info, err := e.root.Stat(path)
		if err != nil {
			return nil, err
		}
		if name == "list_directory" && !info.IsDir() {
			return nil, fmt.Errorf("path is not a directory")
		}
		if name == "read_file" && !info.Mode().IsRegular() {
			return nil, fmt.Errorf("only regular text files can be read")
		}
		file, err := e.root.OpenFile(path, os.O_RDONLY|syscall.O_NONBLOCK, 0)
		if err != nil {
			return nil, err
		}
		defer file.Close()
		info, err = file.Stat()
		if err != nil {
			return nil, err
		}
		if name == "read_file" && !info.Mode().IsRegular() {
			return nil, fmt.Errorf("only regular text files can be read")
		}
		if name == "list_directory" {
			return listDirectory(ctx, file, args.Offset)
		}
		return readTextFile(file, args.Offset)
	default:
		return nil, fmt.Errorf("unknown tool %q", name)
	}
}

func (e *explorer) resolvePath(path string) (string, error) {
	// os.Getwd may preserve a logical (symlinked) PWD. Accept absolute paths
	// expressed using that same spelling, then operate on the pinned real root.
	if filepath.IsAbs(path) && e.environment.PWD != e.dir {
		if rel, err := filepath.Rel(e.environment.PWD, path); err == nil && filepath.IsLocal(rel) {
			path = filepath.Join(e.dir, rel)
		}
	}
	if !filepath.IsAbs(path) {
		path = filepath.Join(e.dir, path)
	}
	check := func(path string) (string, error) {
		rel, err := filepath.Rel(e.dir, path)
		if err != nil || !filepath.IsLocal(rel) {
			return "", fmt.Errorf("path must stay within the starting working directory")
		}
		if sensitivePath(path) {
			return "", fmt.Errorf("credential or private configuration paths cannot be inspected")
		}
		return rel, nil
	}
	if _, err := check(path); err != nil {
		return "", err
	}
	resolved, err := filepath.EvalSymlinks(path)
	if err != nil {
		return "", err
	}
	return check(resolved)
}

func sensitivePath(path string) bool {
	parts := strings.Split(filepath.ToSlash(path), "/")
	for i, part := range parts {
		part = strings.ToLower(part)
		// macOS keeps dcfm's API key outside .config. Also protect it when
		// the working directory is the configuration directory itself.
		if part == "config.json" && i > 0 && strings.EqualFold(parts[i-1], "dcfm") {
			return true
		}
		switch part {
		case ".ssh", ".aws", ".azure", ".gnupg", ".kube", ".config", ".codex", ".git", ".netrc", ".npmrc", ".pypirc", ".git-credentials", "credentials", "credentials.json", "secrets.json", "id_rsa", "id_dsa", "id_ecdsa", "id_ed25519":
			return true
		}
		if part == ".env" || strings.HasPrefix(part, ".env.") || strings.HasSuffix(part, ".pem") || strings.HasSuffix(part, ".key") {
			return true
		}
	}
	return false
}

func listDirectory(ctx context.Context, file *os.File, offset int64) (any, error) {
	// Bound both memory and traversal work even if the model supplies a huge offset.
	if offset > 10000 {
		return nil, fmt.Errorf("directory offset must be at most 10000; choose a subdirectory")
	}
	for skipped := int64(0); skipped < offset; {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		entries, err := file.ReadDir(int(min(offset-skipped, maxDirectoryEntries)))
		skipped += int64(len(entries))
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
	}
	entries, err := file.ReadDir(maxDirectoryEntries + 1)
	if err != nil && err != io.EOF {
		return nil, err
	}
	truncated := len(entries) > maxDirectoryEntries
	if truncated {
		entries = entries[:maxDirectoryEntries]
	}
	type entry struct {
		Name string `json:"name"`
		Type string `json:"type"`
	}
	result := make([]entry, 0, len(entries))
	resultBytes := 256 // reserve space for the result envelope
	for _, item := range entries {
		kind := "file"
		if item.IsDir() {
			kind = "directory"
		} else if item.Type()&os.ModeSymlink != 0 {
			kind = "symlink"
		}
		itemResult := entry{Name: item.Name(), Type: kind}
		encoded, err := json.Marshal(itemResult)
		if err != nil {
			return nil, err
		}
		if resultBytes+len(encoded)+1 > maxToolResultBytes {
			truncated = true
			break
		}
		resultBytes += len(encoded) + 1
		result = append(result, itemResult)
	}
	return map[string]any{"entries": result, "truncated": truncated, "next_offset": offset + int64(len(result))}, nil
}

func readTextFile(file *os.File, offset int64) (any, error) {
	if _, err := file.Seek(offset, io.SeekStart); err != nil {
		return nil, err
	}
	data, err := io.ReadAll(io.LimitReader(file, maxFileBytes+1))
	if err != nil {
		return nil, err
	}
	truncated := len(data) > maxFileBytes
	if truncated {
		data = data[:maxFileBytes]
		// Do not split a multi-byte rune between pages.
		for cut := 0; cut < utf8.UTFMax-1 && !utf8.Valid(data) && len(data) > 0; cut++ {
			data = data[:len(data)-1]
		}
	}
	if !utf8.Valid(data) || bytes.IndexByte(data, 0) >= 0 {
		return nil, fmt.Errorf("file is not UTF-8 text (or offset splits a UTF-8 character)")
	}
	return map[string]any{"content": string(data), "truncated": truncated, "next_offset": offset + int64(len(data))}, nil
}
