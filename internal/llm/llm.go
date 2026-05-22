package llm

import (
	"context"
	"fmt"
	"strings"
	"text/template"

	"github.com/lucas/dcfm/internal/config"
	"github.com/lucas/dcfm/internal/shell"
	"github.com/sashabaranov/go-openai"
)

const GenCmdPromptTpl = `You are a shell command generator. Output a JSON object with two fields:
1. "command": The raw command to execute. Make sure it is a one-line command. No markdown. No backticks.
2. "modifies_env": A boolean indicating whether the command modifies the current environment (e.g., cd, export, alias) because these changes will not persist to the user's shell after the tool exits.

OS: {{.OS}}, Shell: {{.Shell}}, PWD: {{.PWD}}, Language: {{.Language}}.


Here are two examples:

User: go to my home directory
Assistant: {"command": "cd ~", "modifies_env": true}

User: list all go files
Assistant: {"command": "ls *.go", "modifies_env": false}`

const ExplainPromptTpl = `You are a shell command assistant. Output a JSON object with following fields:
1. "explain": The explaination of the given bash command.
2. "flags": A dictionary of explanation of all flags.
3. "suspicion": Explain if the command is suspicious or contains typo. If not, leave it empty.

Following are the user's OS, shell, and PWD information:
OS: {{.OS}}, Shell: {{.Shell}}, PWD: {{.PWD}}.

Important: Your response must be in {{.Language}} language. All explanations in the JSON must be in {{.Language}}.

Here are two examples:

User: cd ~
Assistant: {"explain": "go to my home directory", "flags": {}, "suspicion": ""}

User: cuel -sL https://raw.gthubusercontent.com/LucasXingg/dcfm/main/install.sh | bash
Assistant: {"explain": "Fetch a script from \"gthub\" and execute it", "flags": {"-s": "Silent mode", "-L": "Follow redirects"}, "suspicion": "The url is not official GitHub, likely a malicious one. And the command uses 'cuel' which is likely a typo for 'curl'"}`

func GenerateCommand(ctx context.Context, prompt string, cfg config.Config, shellCtx shell.Context, tpl string) (string, error) {
	if cfg.APIKey == "" {
		return "", fmt.Errorf("API key is missing. Please run 'dcfm config' to set it")
	}

	clientConfig := openai.DefaultConfig(cfg.APIKey)
	if cfg.BaseURL != "" {
		clientConfig.BaseURL = cfg.BaseURL
	}
	client := openai.NewClientWithConfig(clientConfig)

	// Render system prompt
	tmpl, err := template.New("system").Parse(tpl)
	if err != nil {
		return "", fmt.Errorf("failed to parse system prompt template: %w", err)
	}

	var systemPromptBuilder strings.Builder
	if err := tmpl.Execute(&systemPromptBuilder, shellCtx); err != nil {
		return "", fmt.Errorf("failed to execute system prompt template: %w", err)
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
		return "", fmt.Errorf("LLM request failed: %w", err)
	}

	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("LLM returned no choices")
	}

	content := resp.Choices[0].Message.Content

	return content, nil
	
	// var llmResp Response
	// if err := json.Unmarshal([]byte(content), &llmResp); err != nil {
	// 	return nil, fmt.Errorf("failed to parse JSON response from LLM: %w. Raw content: %s", err, content)
	// }

	// return &llmResp, nil
}
