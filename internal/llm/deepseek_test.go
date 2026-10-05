package llm

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/lucas/dcfm/internal/config"
	"github.com/lucas/dcfm/internal/shell"
	"github.com/sashabaranov/go-openai"
)

func TestDeepSeekThinkingRequests(t *testing.T) {
	for _, provider := range []string{"", config.ProviderDeepSeek} {
		for _, mode := range []string{"generate", "explain", "agent"} {
			t.Run(provider+"/"+mode, func(t *testing.T) {
				step := 0
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					step++
					var wire struct {
						openai.ChatCompletionRequest
						Thinking *struct {
							Type string `json:"type"`
						} `json:"thinking"`
					}
					if err := json.NewDecoder(r.Body).Decode(&wire); err != nil {
						t.Error(err)
					}
					req := wire.ChatCompletionRequest
					if provider == config.ProviderDeepSeek {
						if wire.Thinking == nil || wire.Thinking.Type != "enabled" || req.ReasoningEffort != "high" || req.MaxTokens != deepSeekMaxTokens || req.Temperature != 0 {
							t.Errorf("incorrect thinking request: %+v", wire)
						}
					} else if wire.Thinking != nil || req.ReasoningEffort != "" || req.Temperature != 0.2 {
						t.Errorf("legacy request changed: %+v", wire)
					}
					response := openai.ChatCompletionMessage{Role: "assistant", Content: `{"command":"pwd","modifies_env":false}`}
					if mode == "agent" {
						if step <= 2 {
							choice := "auto"
							if step == 1 && provider == "" {
								choice = "required"
							}
							if req.ToolChoice != choice || len(req.Tools) != 4 || req.ResponseFormat != nil {
								t.Errorf("incorrect exploration request: %+v", req)
							}
							response.Content = ""
							response.ReasoningContent = "inspect environment"
							response.ToolCalls = []openai.ToolCall{toolCall("env", "get_environment", `{}`)}
						} else {
							if provider == config.ProviderDeepSeek {
								if req.ToolChoice != nil || len(req.Tools) != 0 {
									t.Error("final thinking request forces tool choice")
								}
							} else if req.ToolChoice != "none" || len(req.Tools) != 4 {
								t.Error("legacy final request changed")
							}
							if req.ResponseFormat == nil {
								t.Error("final JSON format missing")
							}
						}
						if step > 1 {
							for i := 2; i < 2+(step-1)*2; i += 2 {
								if req.Messages[i].ReasoningContent != "inspect environment" || req.Messages[i+1].ToolCallID != "env" {
									t.Error("reasoning/tool history lost")
								}
							}
						}
					} else if req.ResponseFormat == nil {
						t.Error("JSON format missing")
					}
					w.Header().Set("Content-Type", "application/json")
					_ = json.NewEncoder(w).Encode(openai.ChatCompletionResponse{Choices: []openai.ChatCompletionChoice{{Message: response}}})
				}))
				defer server.Close()
				cfg := config.Config{APIKey: "test", Provider: provider, BaseURL: server.URL, Model: "test"}
				var err error
				if mode == "agent" {
					_, err = GenerateCommandAgent(context.Background(), "pwd", cfg, shell.Context{PWD: t.TempDir()}, AgentOptions{MaxRounds: 2})
				} else {
					tpl := GenCmdPromptTpl
					if mode == "explain" {
						tpl = ExplainPromptTpl
					}
					_, err = GenerateCommand(context.Background(), "pwd", cfg, shell.Context{}, tpl)
				}
				if err != nil {
					t.Fatal(err)
				}
				if mode == "agent" && step != 3 {
					t.Fatalf("requests = %d", step)
				}
			})
		}
	}
}
