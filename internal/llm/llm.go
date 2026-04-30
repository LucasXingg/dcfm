package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"text/template"

	"github.com/lucas/dcfm/internal/config"
	"github.com/lucas/dcfm/internal/shell"
	"github.com/sashabaranov/go-openai"
)

type Response struct {
	Command     string `json:"command"`
	ModifiesEnv bool   `json:"modifies_env"`
}

const systemPromptTpl = `You are a shell command generator. Output a JSON object with two fields:
1. "command": The raw command to execute. Make sure it is a one-line command. No markdown. No backticks.
2. "modifies_env": A boolean indicating whether the command modifies the current environment (e.g., cd, export, alias) because these changes will not persist to the user's shell after the tool exits.

OS: {{.OS}}, Shell: {{.Shell}}, PWD: {{.PWD}}.

Here are two examples:

User: go to my home directory
Assistant: {"command": "cd ~", "modifies_env": true}

User: list all go files
Assistant: {"command": "ls *.go", "modifies_env": false}`

func GenerateCommand(ctx context.Context, prompt string, cfg config.Config, shellCtx shell.Context) (*Response, error) {
	if cfg.APIKey == "" {
		return nil, fmt.Errorf("API key is missing. Please run 'dcfm config' to set it")
	}

	clientConfig := openai.DefaultConfig(cfg.APIKey)
	if cfg.BaseURL != "" {
		clientConfig.BaseURL = cfg.BaseURL
	}
	client := openai.NewClientWithConfig(clientConfig)

	// Render system prompt
	tmpl, err := template.New("system").Parse(systemPromptTpl)
	if err != nil {
		return nil, fmt.Errorf("failed to parse system prompt template: %w", err)
	}

	var systemPromptBuilder strings.Builder
	if err := tmpl.Execute(&systemPromptBuilder, shellCtx); err != nil {
		return nil, fmt.Errorf("failed to execute system prompt template: %w", err)
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
		return nil, fmt.Errorf("LLM request failed: %w", err)
	}

	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("LLM returned no choices")
	}

	content := resp.Choices[0].Message.Content
	
	var llmResp Response
	if err := json.Unmarshal([]byte(content), &llmResp); err != nil {
		return nil, fmt.Errorf("failed to parse JSON response from LLM: %w. Raw content: %s", err, content)
	}

	return &llmResp, nil
}
