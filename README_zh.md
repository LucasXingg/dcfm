# dcfm (Do Command For Me)

[English](README.md) | [中文](README_zh.md)

`dcfm` 是一个用 Go 编写的跨平台命令行工具。它能通过 LLM 将自然语言提示翻译成可执行的 Shell 命令。它能够感知运行环境（自动检测操作系统、Shell 类型以及当前目录），并提供合适的命令。

## 特性

- **自然语言转换为 Shell 命令**：将诸如“查找所有大文件”之类的指令转换为您所需的精准 bash、zsh 或 PowerShell 命令。
- **环境感知**：将您的操作系统、Shell 类型和当前工作目录等信息发送给 LLM，从而生成完全量身定制的命令。
- **原生终端支持**：通过附加终端的标准输入/输出执行命令。这意味着像 `vim`、`top` 或 `htop` 这样的交互式命令会就像您自己手动输入一样完美运行。
- **自定义兼容 OpenAI 的 API**：开箱即用地支持 OpenAI 的 `gpt-4o`，但通过更改 Base URL，也可以轻松配置为指向自定义的 API 端点（如 LM Studio、Ollama 或 Azure）。

## 安装

目前，您可以从源码进行安装：

1. 确保您已安装 [Go](https://go.dev/)。
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

*注意：您的配置被安全地存储在 `~/.config/dcfm/config.json`（Windows 下为 `%AppData%/dcfm/config.json`）中，文件权限为 `0600`。*

### 环境变量
您也可以在运行时通过使用环境变量覆盖配置：
- `DCFM_API_KEY`
- `DCFM_BASE_URL`
- `DCFM_MODEL`

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

## 许可证

MIT License
