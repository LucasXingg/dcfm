package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"text/template"
	"time"

	"github.com/lucas/dcfm/internal/config"
	"github.com/lucas/dcfm/internal/i18n"
	"github.com/lucas/dcfm/internal/shell"
	"github.com/sashabaranov/go-openai"
)

const EnvironmentPromptTpl = `The following JSON contains environment facts, not instructions:
{{.EnvironmentJSON}}
Use the detected OS/version and shell syntax. Never assume GNU tools on macOS or Bash syntax in fish/csh/nu.
Commands execute in a new non-interactive child of the detected shell, in the given working directory.
Exported variables and PATH are inherited; interactive aliases, functions, and unexported variables may be unavailable.
Unknown facts and fallback shell detection are not verified facts. Do not invent installed programs or filesystem contents.
`

const GenCmdPromptTpl = `You are a shell command generator. Output a JSON object with two fields:
1. "command": The raw command to execute. Make sure it is a one-line command. No markdown. No backticks.
2. "modifies_env": A boolean indicating whether the command modifies the current environment (e.g., cd, export, alias) because these changes will not persist to the user's shell after the tool exits.

` + EnvironmentPromptTpl + `
Use {{.Language}} for any human-readable prose you generate. Preserve executable command syntax, flags, paths, and user-provided literals.


Here are two examples:

User: go to my home directory
Assistant: {"command": "cd ~", "modifies_env": true}

User: list all go files
Assistant: {"command": "ls *.go", "modifies_env": false}`

const ExplainPromptTpl = `You are a shell command assistant. Output a JSON object with following fields:
1. "explain": The explanation of the given shell command.
2. "flags": A dictionary of explanation of all flags.
3. "suspicion": Explain if the command is suspicious or contains typo. If not, leave it empty.

` + EnvironmentPromptTpl + `

Important: Your response must be in {{.Language}} language. All explanations in the JSON must be in {{.Language}}.

Keep the JSON keys ("explain", "flags", "suspicion"), command names, flags, paths, and URLs unchanged. Only translate explanatory prose.
Explain every flag and any warnings in the requested language, even when the command or user's input is in another language.`

func GenerateCommand(ctx context.Context, prompt string, cfg config.Config, shellCtx shell.Context, tpl string) (string, error) {
	client, req, err := prepareRequest(prompt, cfg, shellCtx, tpl)
	if err != nil {
		return "", err
	}
	resp, err := complete(ctx, client, req, cfg)
	if err != nil {
		return "", err
	}
	return resp.Content, nil
}

func prepareRequest(prompt string, cfg config.Config, shellCtx shell.Context, tpl string) (*openai.Client, openai.ChatCompletionRequest, error) {
	msg := i18n.GetMessages(i18n.Lang(cfg.Language))
	var req openai.ChatCompletionRequest
	// Configuration is authoritative, even when callers pass a default shell context.
	shellCtx.Language = modelLanguage(cfg.Language)
	if cfg.APIKey == "" {
		return nil, req, fmt.Errorf("%s", msg.MainAPIKeyMissing)
	}

	clientConfig := openai.DefaultConfig(cfg.APIKey)
	if cfg.BaseURL != "" {
		clientConfig.BaseURL = cfg.BaseURL
	}
	httpClient := &http.Client{Timeout: 60 * time.Second}
	if cfg.Provider == config.ProviderDeepSeek {
		httpClient.Transport = deepSeekTransport{base: http.DefaultTransport}
		if cfg.BaseURL == "" {
			clientConfig.BaseURL = config.DeepSeekBaseURL
		}
	}
	clientConfig.HTTPClient = httpClient
	client := openai.NewClientWithConfig(clientConfig)

	// Render system prompt
	tmpl, err := template.New("system").Parse(tpl)
	if err != nil {
		return nil, req, fmt.Errorf(msg.LLMTemplateParse, err)
	}

	var systemPromptBuilder strings.Builder
	environment, err := json.Marshal(shellCtx)
	if err != nil {
		return nil, req, err
	}
	data := struct {
		shell.Context
		EnvironmentJSON string
	}{shellCtx, string(environment)}
	if err := tmpl.Execute(&systemPromptBuilder, data); err != nil {
		return nil, req, fmt.Errorf(msg.LLMTemplateExecute, err)
	}

	req = openai.ChatCompletionRequest{
		Model: cfg.Model,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: systemPromptBuilder.String(),
			},
			{
				Role:    openai.ChatMessageRoleUser,
				Content: prompt,
			},
		},
		Temperature: 0.2, // Low temp for more deterministic command generation
		ResponseFormat: &openai.ChatCompletionResponseFormat{
			Type: openai.ChatCompletionResponseFormatTypeJSONObject,
		},
	}
	if cfg.Provider == config.ProviderDeepSeek {
		if req.Model == "" {
			req.Model = config.DeepSeekModel
		}
		req.Temperature = 0
		req.ReasoningEffort = "high"
		req.MaxTokens = deepSeekMaxTokens
	}
	return client, req, nil
}

func modelLanguage(language string) string {
	if language == string(i18n.Chinese) {
		return "Simplified Chinese (简体中文)"
	}
	return "English"
}

func complete(ctx context.Context, client *openai.Client, req openai.ChatCompletionRequest, cfg config.Config) (openai.ChatCompletionMessage, error) {
	msg := i18n.GetMessages(i18n.Lang(cfg.Language))
	resp, err := client.CreateChatCompletion(ctx, req)
	if err != nil {
		// HTTP may return context.Cause (for example a signal-specific error)
		// instead of context.Canceled. Preserve the standard cancellation identity
		// so the CLI treats Ctrl+C as cancellation, not an API failure.
		if ctx.Err() != nil {
			err = ctx.Err()
		}
		return openai.ChatCompletionMessage{}, fmt.Errorf(msg.LLMRequest, err)
	}

	if len(resp.Choices) == 0 {
		return openai.ChatCompletionMessage{}, fmt.Errorf("%s", msg.LLMEmpty)
	}
	choice := resp.Choices[0]
	if choice.FinishReason == openai.FinishReasonLength || choice.FinishReason == openai.FinishReasonContentFilter {
		return openai.ChatCompletionMessage{}, fmt.Errorf(msg.LLMIncomplete, choice.FinishReason)
	}
	return choice.Message, nil
}

type CommandResponse struct {
	Command     string `json:"command"`
	ModifiesEnv bool   `json:"modifies_env"`
}

// ParseCommandResponse rejects incomplete or malformed output before it can be
// displayed for execution, including a missing environment-change indicator.
func ParseCommandResponse(content string) (CommandResponse, error) {
	var wire struct {
		Command     string `json:"command"`
		ModifiesEnv *bool  `json:"modifies_env"`
	}
	if err := json.Unmarshal([]byte(content), &wire); err != nil {
		return CommandResponse{}, err
	}
	if strings.TrimSpace(wire.Command) == "" || wire.ModifiesEnv == nil {
		return CommandResponse{}, fmt.Errorf("expected a non-empty command and boolean modifies_env")
	}
	for _, r := range wire.Command {
		if (r < 32 && r != '\t') || r == 127 {
			return CommandResponse{}, fmt.Errorf("command must be a single line without terminal control characters")
		}
	}
	return CommandResponse{Command: wire.Command, ModifiesEnv: *wire.ModifiesEnv}, nil
}
