# dcfm (Do Command For Me)

[English](README.md) | [中文](README_zh.md)

`dcfm` 是一个用 Go 编写的跨平台命令行工具。它能通过 LLM 将自然语言提示翻译成可执行的 Shell 命令。它能够感知运行环境（自动检测操作系统、Shell 类型以及当前目录），并提供合适的命令。

**支持平台：** macOS 和 Linux。不支持 Windows。

## 特性

- **自然语言转换为 Shell 命令**：将诸如"查找所有大文件"之类的指令转换为您所需的精准 bash 或 zsh 命令。
- **环境感知**：将您的操作系统、Shell 类型和当前工作目录等信息发送给 LLM，从而生成完全量身定制的命令。
- **Agent 模式**：通过多轮只读工具调用，先检查项目结构和可用程序，再给出命令。
- **原生终端支持**：通过附加终端的标准输入/输出执行命令。这意味着像 `vim`、`top` 或 `htop` 这样的交互式命令会就像您自己手动输入一样完美运行。
- **自定义兼容 OpenAI 的 API**：开箱即用地支持 OpenAI 的 `gpt-4o`，但通过更改 Base URL，也可以轻松配置为指向自定义的 API 端点（如 LM Studio、Ollama 或 Azure）。

## 安装

### 一键安装 (macOS & Linux)

安装 `dcfm` 最简单的方法是使用我们的安装脚本：

```bash
curl -sL https://raw.githubusercontent.com/LucasXingg/dcfm/main/install.sh | bash
```

### 从源码安装

如果您更愿意，可以从源码进行安装：

1. 确保您已安装 [Go](https://go.dev/) 1.26.1 或更新版本（以 `go.mod` 为准）。
2. 克隆仓库并构建：
   ```bash
   git clone https://github.com/LucasXingg/dcfm.git
   cd dcfm
   go build -o dcfm
   ```
3. 将二进制文件移动到您的系统 PATH 中，例如：
   ```bash
   sudo mv dcfm /usr/local/bin/
   ```

## 配置

在使用 `dcfm` 之前，您需要设置您的 API 配置。

```bash
dcfm -c
```
将会通过交互式方式提示您输入以下内容：
- **API Key**：您的 OpenAI API 密钥（或自定义提供商密钥）。
- **Base URL**：默认为 OpenAI，但可以设置为任何兼容 OpenAI 的 API 端点。
- **Model Name**：默认为 `gpt-4o`。

*注意：配置文件权限为 `0600`，使用操作系统的用户配置目录：macOS 为 `~/Library/Application Support/dcfm/config.json`；Linux 为 `$XDG_CONFIG_HOME/dcfm/config.json`，未设置该变量时为 `~/.config/dcfm/config.json`。*

### 环境变量
您也可以在运行时通过使用环境变量覆盖配置：
- `DCFM_API_KEY`
- `DCFM_BASE_URL`
- `DCFM_MODEL`
- `DCFM_LANGUAGE` (`en` / `zh`)

## 用法

只需运行 `dcfm`，紧接着输入您的提示词即可：

```bash
dcfm 列出此目录中的所有 json 文件并按大小排序
```

**示例输出：**
```
Generating command...

Proposed Command: ls -lhS *.json

? Execute (enter) / Cancel (q) / Edit (type message):

```

如果生成的命令不太准确，您可以继续提供改进意见（例如：“只显示前 5 个”）。该工具将根据您的反馈生成一条新的命令。

### Agent 模式

在希望检查的目录中运行：

```bash
dcfm --agent "看看这个项目，告诉我运行测试应该用什么命令"
dcfm -a "找到适合搜索当前目录源代码的命令"
dcfm -a --agent-max-rounds 4 "这个项目应该怎么构建？"
```

模型会自主选择工具、读取返回结果，并根据需要继续查询，最后给出命令。终端会显示每次工具调用。最终仍使用普通模式的「执行 / 取消 / 修改」交互；输入修改意见后，会携带原请求、上一条命令和反馈重新探索。

| 工具 | 可获取的信息 |
| --- | --- |
| `get_environment` | 操作系统及版本、架构、检测到的 shell 及检测来源、当前目录 |
| `list_directory` | 目录条目（包含隐藏项），每页最多 100 项 |
| `read_file` | 普通 UTF-8 文本文件，每页最多 4 KiB |
| `find_executable` | 程序是否位于继承的 PATH 中，以及可执行文件路径 |

探索阶段不会执行 shell 命令或程序。文件读取范围限定在启动目录内部，并检查符号链接是否越界；禁止读取 `.env`、`.ssh`、`.aws`、`.config`、私钥等常见凭据路径。这些检查并非通用的秘密信息检测器，源文件仍可能包含私密内容。读取的文件名和文件内容会发送给您配置的模型服务，请在适合分享给该服务的目录中使用 Agent 模式。工具不会导出全部环境变量的值。

默认最多探索 8 轮；达到上限时，额外请求一次最终命令。`--agent-max-rounds` 可设置为 1–20。每轮最多调用 8 个工具，总计最多 32 次；工具输出大小也有限制。单次 API 请求超时为 60 秒，一次探索总超时为 3 分钟。生成过程中可按 Ctrl+C 取消。模型和 API 需要支持 Chat Completions 函数调用（`tools` / `tool_choice`）及 JSON 输出，不支持时会明确报错；去掉 `--agent` 即使用普通模式。`--agent` 不能与 `--explain`、`--config` 同时使用。

### 环境信息检测

两种模式都会优先识别父进程链中最近的已知 shell；无法识别时依次回退到 `$SHELL` 和 `/bin/sh`。`$SHELL` 通常表示登录 shell，在切换 shell 后可能不准确。提示词会包含检测来源、系统版本或 Linux 发行版、CPU 架构，以及从操作系统获取的当前工作目录；未知值会明确标记。

执行命令时使用传给模型的同一个 shell 和目录。命令在新的子 shell 中运行，因此交互式别名、函数和未导出的变量可能不可用。`cd`、`export` 等需要修改父 shell 环境的命令，仍沿用复制到剪贴板的流程。

## 开发与测试

```bash
go test -race -timeout 2m ./...
go vet ./...
go build ./...
```

测试使用本地模拟 API，不需要模型密钥，也不会产生付费请求。CI 会在 Linux 和 macOS 上运行测试。

## 许可证

MIT License
