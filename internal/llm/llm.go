package llm

import (
	"context"
	"fmt"
	"strings"
	"text/template"

	"github.com/lucas/dcfm/internal/config"
	"github.com/lucas/dcfm/internal/i18n"
	"github.com/lucas/dcfm/internal/shell"
	"github.com/sashabaranov/go-openai"
)

const GenCmdPromptTpl = `You are a shell command generator. Output a JSON object with two fields:
1. "command": The raw command to execute. Make sure it is a one-line command. No markdown. No backticks.
2. "modifies_env": A boolean indicating whether the command modifies the current environment (e.g., cd, export, alias) because these changes will not persist to the user's shell after the tool exits.

OS: {{.OS}}, Shell: {{.Shell}}, PWD: {{.PWD}}, Language: {{.Language}}.
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

Following are the user's OS, shell, and PWD information:
OS: {{.OS}}, Shell: {{.Shell}}, PWD: {{.PWD}}.

Important: Your response must be in {{.Language}} language. All explanations in the JSON must be in {{.Language}}.

Keep the JSON keys ("explain", "flags", "suspicion"), command names, flags, paths, and URLs unchanged. Only translate explanatory prose.
Explain every flag and any warnings in the requested language, even when the command or user's input is in another language.`

func GenerateCommand(ctx context.Context, prompt string, cfg config.Config, shellCtx shell.Context, tpl string) (string, error) {
	msg := i18n.GetMessages(i18n.Lang(cfg.Language))
	// Configuration is authoritative, even when callers pass a default shell context.
	shellCtx.Language = "English"
	if cfg.Language == string(i18n.Chinese) {
		shellCtx.Language = "Simplified Chinese (简体中文)"
	}
	if cfg.APIKey == "" {
		return "", fmt.Errorf("%s", msg.MainAPIKeyMissing)
	}

	clientConfig := openai.DefaultConfig(cfg.APIKey)
	if cfg.BaseURL != "" {
		clientConfig.BaseURL = cfg.BaseURL
	}
	client := openai.NewClientWithConfig(clientConfig)

	// Render system prompt
	tmpl, err := template.New("system").Parse(tpl)
	if err != nil {
		return "", fmt.Errorf(msg.LLMTemplateParse, err)
	}

	var systemPromptBuilder strings.Builder
	if err := tmpl.Execute(&systemPromptBuilder, shellCtx); err != nil {
		return "", fmt.Errorf(msg.LLMTemplateExecute, err)
	}

	req := openai.ChatCompletionRequest{
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

	resp, err := client.CreateChatCompletion(ctx, req)
	if err != nil {
		return "", fmt.Errorf(msg.LLMRequest, err)
	}

	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("%s", msg.LLMEmpty)
	}

	content := resp.Choices[0].Message.Content

	return content, nil
}
