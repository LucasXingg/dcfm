package i18n

import "fmt"

type Lang string

const (
	English Lang = "en"
	Chinese Lang = "zh"
)

type Messages struct {
	ConfigTitle                string
	ConfigSelectPrompt         string
	ConfigSelectAll            string
	ConfigSelectAPIKey         string
	ConfigSelectBaseURL        string
	ConfigSelectModel          string
	ConfigSelectLanguage       string
	ConfigAPIKeyPrompt         string
	ConfigBaseURLPrompt        string
	ConfigModelPrompt          string
	ConfigLanguagePrompt       string
	ConfigLanguageEnglish      string
	ConfigLanguageChinese      string
	ConfigCancelled            string
	ConfigSaved                string
	ConfigSaveError            string
	ConfigLoadWarning          string
	MainGenerating             string
	MainProposedCommand        string
	MainExecutePrompt          string
	MainExecuteEnter           string
	MainExecuteCancel          string
	MainExecuteEdit            string
	MainWarningEnv             string
	MainClipboardCopied        string
	MainClipboardCopiedRemote  string
	MainClipboardError         string
	MainCommandError           string
	MainAPIKeyMissing          string
	MainErrorLoadingConfig     string
	MainErrorGeneratingCommand string
	MainErrorParsingResponse   string
}

var messages = map[Lang]Messages{
	English: {
		ConfigTitle:                "dcfm Configuration",
		ConfigSelectPrompt:         "What would you like to configure?",
		ConfigSelectAll:            "Configure All Settings",
		ConfigSelectAPIKey:         "API Key",
		ConfigSelectBaseURL:        "Base URL",
		ConfigSelectModel:          "Model",
		ConfigSelectLanguage:       "Language",
		ConfigAPIKeyPrompt:         "Enter your API Key:",
		ConfigBaseURLPrompt:        "Enter Base URL (OpenAI compatible):",
		ConfigModelPrompt:          "Enter Model name:",
		ConfigLanguagePrompt:       "Select language:",
		ConfigLanguageEnglish:      "English",
		ConfigLanguageChinese:      "中文",
		ConfigCancelled:            "Configuration cancelled.",
		ConfigSaved:                "Configuration saved successfully.",
		ConfigSaveError:            "Error saving configuration: %v",
		ConfigLoadWarning:          "Warning: could not load existing config: %v",
		MainGenerating:             "Generating command...",
		MainProposedCommand:        "Proposed Command:",
		MainExecutePrompt:          "Execute (enter) / Cancel (q) / Edit (type message):",
		MainExecuteEnter:           "Cancelled.",
		MainExecuteCancel:          "Cancelled.",
		MainExecuteEdit:            "",
		MainWarningEnv:             "Warning: This command modifies the environment and cannot be executed directly by dcfm since changes won't persist to your shell.",
		MainClipboardCopied:        "The command has been copied to your clipboard. Paste it into your terminal to run it.",
		MainClipboardCopiedRemote:  "The command has been copied to your clipboard via OSC 52. Ensure your terminal emulator supports it. Paste it into your terminal to run it.",
		MainClipboardError:         "Failed to copy command to clipboard: %v",
		MainCommandError:           "Command finished with error: %v",
		MainAPIKeyMissing:          "API key is missing. Please run 'dcfm config' to set it",
		MainErrorLoadingConfig:     "Error loading config: %v",
		MainErrorGeneratingCommand: "Error generating command: %v",
		MainErrorParsingResponse:   "failed to parse JSON response from LLM: %w. Raw content: %s",
	},
	Chinese: {
		ConfigTitle:                "dcfm 配置",
		ConfigSelectPrompt:         "您想要配置什么？",
		ConfigSelectAPIKey:         "API 密钥",
		ConfigSelectBaseURL:        "基础 URL",
		ConfigSelectModel:          "模型",
		ConfigSelectLanguage:       "语言",
		ConfigSelectAll:            "配置所有设置",
		ConfigAPIKeyPrompt:         "请输入您的 API 密钥：",
		ConfigBaseURLPrompt:        "请输入基础 URL（兼容 OpenAI）：",
		ConfigModelPrompt:          "请输入模型名称：",
		ConfigLanguagePrompt:       "选择语言：",
		ConfigLanguageEnglish:      "English",
		ConfigLanguageChinese:      "中文",
		ConfigCancelled:            "配置已取消。",
		ConfigSaved:                "配置保存成功。",
		ConfigSaveError:            "保存配置时出错：%v",
		ConfigLoadWarning:          "警告：无法加载现有配置：%v",
		MainGenerating:             "正在生成命令...",
		MainProposedCommand:        "建议的命令：",
		MainExecutePrompt:          "执行（回车）/ 取消（q）/ 编辑（输入消息）：",
		MainExecuteEnter:           "已取消。",
		MainExecuteCancel:          "已取消。",
		MainExecuteEdit:            "",
		MainWarningEnv:             "警告：此命令会修改环境变量，dcfm 无法直接执行，因为更改不会持久保存到您的 shell 中。",
		MainClipboardCopied:        "命令已复制到剪贴板，请粘贴到终端运行。",
		MainClipboardCopiedRemote:  "命令已通过 OSC 52 复制到剪贴板，请确保您的终端模拟器支持此功能。粘贴到终端运行。",
		MainClipboardError:         "复制命令到剪贴板失败：%v",
		MainCommandError:           "命令执行出错：%v",
		MainAPIKeyMissing:          "API 密钥缺失，请运行 'dcfm config' 设置",
		MainErrorLoadingConfig:     "加载配置出错：%v",
		MainErrorGeneratingCommand: "生成命令出错：%v",
		MainErrorParsingResponse:   "解析 LLM JSON 响应失败：%w。原始内容：%s",
	},
}

func GetMessages(lang Lang) Messages {
	if msg, ok := messages[lang]; ok {
		return msg
	}
	return messages[English]
}

func GetLanguageName(lang Lang, displayLang Lang) string {
	msg := GetMessages(displayLang)
	if lang == English {
		return msg.ConfigLanguageEnglish
	}
	return msg.ConfigLanguageChinese
}

func Format(format string, args ...interface{}) string {
	return fmt.Sprintf(format, args...)
}