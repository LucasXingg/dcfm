package i18n

import "fmt"

type Lang string

const (
	English Lang = "en"
	Chinese Lang = "zh"
)

type Messages struct {
	ExplainInput           string
	ExplainEmpty           string
	ExplainGenerating      string
	ExplainError           string
	ExplainTitle           string
	ExplainSuspicion       string
	ExplainDisclaimer      string
	RootDescription        string
	ConfigFlag             string
	ExplainFlag            string
	HelpFlag               string
	HelpUsage              string
	HelpFlags              string
	SelectHint             string
	LLMTemplateParse       string
	LLMTemplateExecute     string
	LLMRequest             string
	LLMEmpty               string
	LLMIncomplete          string
	AgentFlag              string
	AgentMaxRoundsFlag     string
	AgentExploring         string
	AgentToolCall          string
	AgentInvalidRounds     string
	AgentRoundsRequireFlag string
	AgentEnvironmentError  string
	AgentInvalidResponse   string
	AgentLimitExceeded     string
	AgentTooManyTools      string
	AgentInvalidToolCall   string
	AgentNoExploration     string

	ConfigSelectProvider       string
	ConfigProviderPrompt       string
	ConfigProviderCustom       string
	ConfigProviderDeepSeek     string
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
		ExplainInput:           "Enter command to explain:",
		ExplainEmpty:           "No command provided.",
		ExplainGenerating:      "Generating explanation...",
		ExplainError:           "Error explaining command: %v",
		ExplainTitle:           "Explanation:",
		ExplainSuspicion:       "Your command may contain suspicious behavior:",
		ExplainDisclaimer:      "The model may be wrong; please double-check carefully.",
		RootDescription:        "dcfm translates natural language into shell commands",
		ConfigFlag:             "Configure API key, base URL, model, and language",
		ExplainFlag:            "Explain the command you pass in",
		HelpFlag:               "Show help",
		HelpUsage:              "Usage:",
		HelpFlags:              "Flags:",
		SelectHint:             "Use arrows to move, type to filter",
		LLMTemplateParse:       "Failed to parse system prompt template: %w",
		LLMTemplateExecute:     "Failed to render system prompt template: %w",
		LLMRequest:             "LLM request failed: %w",
		LLMEmpty:               "LLM returned no choices",
		LLMIncomplete:          "LLM response is incomplete (%s); no command will be executed",
		AgentFlag:              "Explore the environment with read-only tools before generating a command",
		AgentMaxRoundsFlag:     "Maximum exploration rounds (1-20; requires --agent)",
		AgentExploring:         "Exploring the environment (read-only)...",
		AgentToolCall:          "  Inspecting: %q %q",
		AgentInvalidRounds:     "Agent exploration rounds must be between 1 and 20",
		AgentRoundsRequireFlag: "--agent-max-rounds requires --agent",
		AgentEnvironmentError:  "Cannot inspect the working directory: %w",
		AgentInvalidResponse:   "Invalid final command from agent: %w",
		AgentLimitExceeded:     "The model continued requesting tools after the exploration limit; no command will be executed",
		AgentTooManyTools:      "The model requested more than 8 tools in one round",
		AgentInvalidToolCall:   "The model returned an invalid tool call (missing/duplicate ID or unsupported type)",
		AgentNoExploration:     "The model did not call any tools. Agent mode requires a model and API that support function calling; use normal mode otherwise",

		ConfigSelectProvider:       "Provider",
		ConfigProviderPrompt:       "Select provider:",
		ConfigProviderCustom:       "OpenAI / Custom (keep current settings)",
		ConfigProviderDeepSeek:     "DeepSeek (official, thinking enabled)",
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
		MainAPIKeyMissing:          "API key is missing. Please run 'dcfm -c' to set it",
		MainErrorLoadingConfig:     "Error loading config: %v",
		MainErrorGeneratingCommand: "Error generating command: %v",
		MainErrorParsingResponse:   "failed to parse JSON response from LLM: %v. Raw content: %s",
	},
	Chinese: {
		ExplainInput:           "请输入要解释的命令：",
		ExplainEmpty:           "未提供命令。",
		ExplainGenerating:      "正在生成解释...",
		ExplainError:           "解释命令出错：%v",
		ExplainTitle:           "命令解释：",
		ExplainSuspicion:       "您的命令可能包含可疑行为：",
		ExplainDisclaimer:      "模型可能出错，请仔细核查。",
		RootDescription:        "dcfm 将自然语言转换为 shell 命令",
		ConfigFlag:             "配置 API 密钥、基础 URL、模型和语言",
		ExplainFlag:            "解释传入的命令",
		HelpFlag:               "显示帮助",
		HelpUsage:              "用法：",
		HelpFlags:              "选项：",
		SelectHint:             "使用方向键移动，输入文字筛选",
		LLMTemplateParse:       "解析系统提示模板失败：%w",
		LLMTemplateExecute:     "渲染系统提示模板失败：%w",
		LLMRequest:             "模型请求失败：%w",
		LLMEmpty:               "模型未返回任何结果",
		LLMIncomplete:          "模型响应不完整（%s），不会执行命令",
		AgentFlag:              "先用只读工具探索环境，再生成命令",
		AgentMaxRoundsFlag:     "最大环境探索轮数（1-20；需同时使用 --agent）",
		AgentExploring:         "正在探索环境（只读）...",
		AgentToolCall:          "  正在查询：%q %q",
		AgentInvalidRounds:     "环境探索轮数必须在 1 到 20 之间",
		AgentRoundsRequireFlag: "--agent-max-rounds 需要同时使用 --agent",
		AgentEnvironmentError:  "无法检查当前工作目录：%w",
		AgentInvalidResponse:   "Agent 返回的最终命令无效：%w",
		AgentLimitExceeded:     "模型在达到探索上限后仍请求调用工具，不会执行命令",
		AgentTooManyTools:      "模型在一轮中请求了超过 8 次工具调用",
		AgentInvalidToolCall:   "模型返回了无效的工具调用（ID 缺失或重复，或类型不支持）",
		AgentNoExploration:     "模型未调用任何工具。Agent 模式需要模型和 API 支持函数调用，否则请使用普通模式",

		ConfigSelectProvider:       "提供商",
		ConfigProviderPrompt:       "选择提供商：",
		ConfigProviderCustom:       "OpenAI / 自定义（保留当前设置）",
		ConfigProviderDeepSeek:     "DeepSeek（官方，启用思考模式）",
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
		MainAPIKeyMissing:          "API 密钥缺失，请运行 'dcfm -c' 设置",
		MainErrorLoadingConfig:     "加载配置出错：%v",
		MainErrorGeneratingCommand: "生成命令出错：%v",
		MainErrorParsingResponse:   "解析 LLM JSON 响应失败：%v。原始内容：%s",
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
