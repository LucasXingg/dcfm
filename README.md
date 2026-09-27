# dcfm (Do Command For Me)

[English](README.md) | [中文](README_zh.md)

`dcfm` is a cross-platform CLI tool written in Go that translates natural language prompts into executable shell commands using an LLM. It is environment-aware (detects OS, Shell, and current directory) and provide suitable commands based on the context.

**Supported Platforms:** macOS and Linux. Windows is not supported.

## Features

- **Natural Language to Shell**: Converts commands like "find all big files" into the exact bash or zsh command you need.
- **Environment-Aware**: Sends your OS, shell type, and current working directory to the LLM for perfectly tailored commands.
- **Agent Mode**: Uses multiple rounds of read-only tool calls to inspect your project and available programs before proposing a command.
- **Native Terminal Support**: Executes commands by attaching your terminal's standard input/output. This means interactive commands like `vim`, `top`, or `htop` work exactly as if you typed them yourself.
- **Custom OpenAI-Compatible API**: Works out of the box with OpenAI's `gpt-4o`, but easily configurable to point to custom API endpoints (like LM Studio, Ollama, or Azure) by changing the Base URL.

## Installation

### One-line Installation (macOS & Linux)

The easiest way to install `dcfm` is using our install script:

```bash
curl -sL https://raw.githubusercontent.com/LucasXingg/dcfm/main/install.sh | bash
```

### Install from Source

If you prefer, you can build from source:

1. Ensure you have [Go](https://go.dev/) 1.26.1 or later installed (see `go.mod`).
2. Clone the repository and build:
   ```bash
   git clone https://github.com/LucasXingg/dcfm.git
   cd dcfm
   go build -o dcfm
   ```
3. Move the binary to your path, for example:
   ```bash
   sudo mv dcfm /usr/local/bin/
   ```

## Configuration

Before using `dcfm`, you need to set up your API configuration.

```bash
dcfm -c
```
You will be interactively prompted for:
- **API Key**: Your OpenAI API key (or custom provider key).
- **Base URL**: Defaults to OpenAI, but can be set to any OpenAI-compatible API endpoint.
- **Model Name**: Defaults to `gpt-4o`.

*Note: Your configuration is stored with `0600` permissions in the OS user configuration directory: `~/Library/Application Support/dcfm/config.json` on macOS; `$XDG_CONFIG_HOME/dcfm/config.json` on Linux, or `~/.config/dcfm/config.json` when that variable is unset.*

### Environment Variables
You can also override the configuration at runtime using environment variables:
- `DCFM_API_KEY`
- `DCFM_BASE_URL`
- `DCFM_MODEL`
- `DCFM_LANGUAGE` (`en` / `zh`)

## Usage

Simply run `dcfm` followed by your prompt:

```bash
dcfm list all json files in this directory and sort by size
```

**Example Output:**
```
Generating command...

Proposed Command: ls -lhS *.json

? Execute (enter) / Cancel (q) / Edit (type message):

```

If the command isn't quite right, provide a refinement (e.g., "just show the top 5"). The tool will generate a new command based on your feedback.

### Agent mode

Run from the directory you want to inspect:

```bash
dcfm --agent "Look at this project and give me the command to run its tests"
dcfm -a "Find a suitable command to search the source files in this directory"
dcfm -a --agent-max-rounds 4 "How do I build this project?"
```

The model selects tools, receives their results, and can request further inspection before returning a command. Each tool call is shown in the terminal. The resulting command uses the same execute / cancel / refine prompt as normal mode. Refining a request starts a fresh exploration with your previous request, proposed command, and feedback.

| Tool | Information available |
| --- | --- |
| `get_environment` | OS/version, architecture, detected shell and detection source, working directory |
| `list_directory` | Directory entries, including hidden entries, in pages of up to 100 |
| `read_file` | Regular UTF-8 text files, in pages of up to 4 KiB |
| `find_executable` | Whether a program exists on inherited PATH, and its executable path |

Exploration never executes shell commands or programs. File access stays inside the starting directory, including checks against symlink escapes; common credential paths such as `.env`, `.ssh`, `.aws`, `.config`, and private key files are blocked. These checks are not a general secret detector: relevant source files can still contain private information. Inspected filenames and file contents are sent to your configured model provider, so use agent mode only in directories you are comfortable sharing with that provider. Environment variable values are not dumped.

The default budget is 8 exploration rounds, with one additional final request if the budget is reached. `--agent-max-rounds` accepts 1–20. Limits also apply to tool calls (8 per response, 32 total), tool output, and time (60 seconds per API request, 3 minutes per exploration). Press Ctrl+C to cancel generation. The API and model must support Chat Completions function calling (`tools` / `tool_choice`) and JSON output; unsupported requests fail visibly. Omit `--agent` to use normal mode. `--agent` cannot be combined with `--explain` or `--config`.

### Environment detection

Both modes detect the nearest recognized shell in the parent process chain, falling back to `$SHELL` and then `/bin/sh` when detection is unavailable. `$SHELL` alone describes the login shell and may be stale after switching shells. Prompts include the detection source, OS version/distribution, CPU architecture, and the current directory from the operating system. Unknown values are labelled explicitly.

Execution uses the same shell and directory that were supplied to the model. Commands run in a new child shell, so interactive aliases, functions, and unexported variables may not be available. Commands that change the parent environment, such as `cd` and `export`, continue to use the existing clipboard workflow.

## Development

```bash
go test -race -timeout 2m ./...
go vet ./...
go build ./...
```

Tests use a local mock API; no model credentials or paid requests are needed. CI runs the tests on Linux and macOS.

## License

MIT License
