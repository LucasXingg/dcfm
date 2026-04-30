# `dcfm` (Do Command For Me)

`dcfm` is a cross-platform CLI tool written in Go that translates natural language prompts into executable shell commands using an LLM. It is environment-aware (detects OS, Shell, and current directory) and features a seamless "Review-Edit-Execute" loop.

## Features

- **Natural Language to Shell**: Converts commands like "find all big files" into the exact bash, zsh, or PowerShell command you need.
- **Environment-Aware**: Sends your OS, shell type, and current working directory to the LLM for perfectly tailored commands (e.g., handles Conda activation correctly).
- **Interactive Review Loop**: Presents the proposed command and allows you to `Run`, `Edit` (refine your prompt), or `Cancel`.
- **Native Terminal Support**: Executes commands by attaching your terminal's standard input/output. This means interactive commands like `vim`, `top`, or `htop` work exactly as if you typed them yourself.
- **Custom OpenAI-Compatible API**: Works out of the box with OpenAI's `gpt-4o`, but easily configurable to point to custom API endpoints (like LM Studio, Ollama, or Azure) by changing the Base URL.

## Installation

Currently, you can install from source:

1. Ensure you have [Go](https://go.dev/) installed.
2. Clone the repository and build:
   ```bash
   git clone https://github.com/lucas/dcfm.git
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
dcfm config
```
You will be interactively prompted for:
- **API Key**: Your OpenAI API key (or custom provider key).
- **Base URL**: Defaults to OpenAI, but can be set to any OpenAI-compatible API endpoint.
- **Model Name**: Defaults to `gpt-4o`.

*Note: Your configuration is stored securely with `0600` permissions in `~/.config/dcfm/config.json` (or `%AppData%/dcfm/config.json` on Windows).*

### Environment Variables
You can also override the configuration at runtime using environment variables:
- `DCFM_API_KEY`
- `DCFM_BASE_URL`
- `DCFM_MODEL`

## Usage

Simply run `dcfm` followed by your prompt:

```bash
dcfm list all json files in this directory and sort by size
```

**Example Output:**
```
Generating command...

Proposed Command: ls -lhS *.json

? What would you like to do?  [Use arrows to move, type to filter]
> Run
  Edit
  Cancel
```

If the command isn't quite right, select `Edit` to provide a refinement (e.g., "just show the top 5"). The tool will generate a new command based on your feedback.

## License

MIT License
