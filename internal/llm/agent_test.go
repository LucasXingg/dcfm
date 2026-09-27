package llm

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/lucas/dcfm/internal/config"
	"github.com/lucas/dcfm/internal/shell"
	"github.com/sashabaranov/go-openai"
)

func mockModel(t *testing.T, handler func(openai.ChatCompletionRequest, int) openai.ChatCompletionMessage) (config.Config, *atomic.Int32) {
	t.Helper()
	var count atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req openai.ChatCompletionRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Error(err)
		}
		if r.URL.Path != "/chat/completions" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Bearer test-key" {
			t.Error("API key missing")
		}
		message := handler(req, int(count.Add(1)))
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(openai.ChatCompletionResponse{Choices: []openai.ChatCompletionChoice{{Message: message}}})
	}))
	t.Cleanup(server.Close)
	return config.Config{APIKey: "test-key", BaseURL: server.URL, Model: "test-model", Language: "en"}, &count
}

func toolCall(id, name, arguments string) openai.ToolCall {
	return openai.ToolCall{ID: id, Type: openai.ToolTypeFunction, Function: openai.FunctionCall{Name: name, Arguments: arguments}}
}

func TestAgentExploresMultipleRoundsAndPreservesHistory(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "package.json"), []byte(`{"scripts":{"test":"vitest"}}`), 0600); err != nil {
		t.Fatal(err)
	}
	final := `{"command":"npm test","modifies_env":false}`
	cfg, count := mockModel(t, func(req openai.ChatCompletionRequest, step int) openai.ChatCompletionMessage {
		if req.Model != "test-model" || len(req.Tools) != 4 {
			t.Errorf("incorrect model/tools: %+v", req)
		}
		if req.ResponseFormat != nil {
			t.Error("tool requests should not force JSON response format")
		}
		switch step {
		case 1:
			if req.ToolChoice != "required" || len(req.Messages) != 2 {
				t.Error("first turn must require exploration")
			}
			return openai.ChatCompletionMessage{Role: "assistant", ReasoningContent: "inspect project", ToolCalls: []openai.ToolCall{
				toolCall("dir-1", "list_directory", `{}`), toolCall("env-1", "get_environment", `{}`),
			}}
		case 2:
			if req.ToolChoice != "auto" || len(req.Messages) != 5 {
				t.Errorf("history size = %d", len(req.Messages))
				return openai.ChatCompletionMessage{}
			}
			if req.Messages[2].ReasoningContent != "inspect project" || len(req.Messages[2].ToolCalls) != 2 {
				t.Error("assistant history was lost")
			}
			for i, id := range []string{"dir-1", "env-1"} {
				m := req.Messages[i+3]
				if m.Role != "tool" || m.ToolCallID != id || !json.Valid([]byte(m.Content)) {
					t.Errorf("invalid tool response: %+v", m)
				}
			}
			if !strings.Contains(req.Messages[3].Content, "package.json") {
				t.Error("directory result missing")
			}
			return openai.ChatCompletionMessage{ToolCalls: []openai.ToolCall{toolCall("file-1", "read_file", `{"path":"package.json"}`)}}
		case 3:
			if len(req.Messages) != 7 || !strings.Contains(req.Messages[6].Content, "vitest") {
				t.Error("read result missing from next turn")
			}
			return openai.ChatCompletionMessage{Content: final}
		default:
			t.Error("unexpected extra request")
			return openai.ChatCompletionMessage{}
		}
	})
	var progress []string
	result, err := GenerateCommandAgent(context.Background(), "run tests", cfg, shell.Context{PWD: dir, Shell: "/bin/sh"}, AgentOptions{OnToolCall: func(name, _ string) { progress = append(progress, name) }})
	if err != nil || result != final {
		t.Fatalf("result = %q, err = %v", result, err)
	}
	if count.Load() != 3 || len(progress) != 3 {
		t.Fatalf("requests = %d, progress = %v", count.Load(), progress)
	}
}

func TestAgentFeedsToolErrorsBackToModel(t *testing.T) {
	cfg, _ := mockModel(t, func(req openai.ChatCompletionRequest, step int) openai.ChatCompletionMessage {
		if step == 1 {
			return openai.ChatCompletionMessage{ToolCalls: []openai.ToolCall{
				toolCall("bad-json", "read_file", `{"path":`),
				toolCall("unknown", "execute_shell", `{"command":"touch should-not-exist"}`),
				toolCall("missing", "read_file", `{"path":"absent"}`),
			}}
		}
		for _, m := range req.Messages[3:] {
			if m.Role != "tool" || !strings.Contains(m.Content, `"error"`) {
				t.Errorf("missing error result: %+v", m)
			}
		}
		return openai.ChatCompletionMessage{Content: `{"command":"ls","modifies_env":false}`}
	})
	_, err := GenerateCommandAgent(context.Background(), "list files", cfg, shell.Context{PWD: t.TempDir()}, AgentOptions{})
	if err != nil {
		t.Fatal(err)
	}
}

func TestAgentBoundedExploration(t *testing.T) {
	for _, test := range []struct {
		name                                        string
		rounds, callsPerRound, requests, executions int
	}{
		{"round limit", 1, 1, 2, 1},
		{"total tool limit", 20, 7, 6, 32},
	} {
		t.Run(test.name, func(t *testing.T) {
			cfg, count := mockModel(t, func(req openai.ChatCompletionRequest, step int) openai.ChatCompletionMessage {
				if step == test.requests {
					if req.ToolChoice != "none" || req.ResponseFormat == nil {
						t.Error("final request must disable tools and require JSON")
					}
					return openai.ChatCompletionMessage{Content: `{"command":"pwd","modifies_env":false}`}
				}
				var calls []openai.ToolCall
				for i := 0; i < test.callsPerRound; i++ {
					calls = append(calls, toolCall(fmt.Sprintf("%d-%d", step, i), "get_environment", `{}`))
				}
				return openai.ChatCompletionMessage{ToolCalls: calls}
			})
			executions := 0
			_, err := GenerateCommandAgent(context.Background(), "pwd", cfg, shell.Context{PWD: t.TempDir()}, AgentOptions{MaxRounds: test.rounds, OnToolCall: func(_, _ string) { executions++ }})
			if err != nil {
				t.Fatal(err)
			}
			if int(count.Load()) != test.requests || executions != test.executions {
				t.Fatalf("requests=%d, executions=%d", count.Load(), executions)
			}
		})
	}
}

func TestAgentRejectsInvalidModelResponses(t *testing.T) {
	for _, test := range []struct {
		name     string
		response openai.ChatCompletionMessage
		want     string
	}{
		{"no exploration", openai.ChatCompletionMessage{Content: `{"command":"ls","modifies_env":false}`}, "did not call any tools"},
		{"missing id", openai.ChatCompletionMessage{ToolCalls: []openai.ToolCall{toolCall("", "get_environment", `{}`)}}, "invalid tool call"},
		{"duplicate id", openai.ChatCompletionMessage{ToolCalls: []openai.ToolCall{toolCall("same", "get_environment", `{}`), toolCall("same", "get_environment", `{}`)}}, "invalid tool call"},
		{"too many calls", openai.ChatCompletionMessage{ToolCalls: make([]openai.ToolCall, 9)}, "more than 8"},
	} {
		t.Run(test.name, func(t *testing.T) {
			cfg, count := mockModel(t, func(openai.ChatCompletionRequest, int) openai.ChatCompletionMessage { return test.response })
			_, err := GenerateCommandAgent(context.Background(), "test", cfg, shell.Context{PWD: t.TempDir()}, AgentOptions{})
			if err == nil || !strings.Contains(err.Error(), test.want) || count.Load() != 1 {
				t.Fatalf("err=%v, calls=%d", err, count.Load())
			}
		})
	}
	for _, content := range []string{"", "not json", `{}`, `{"command":"ls"}`, `{"command":"ls\nrm file","modifies_env":false}`} {
		t.Run("invalid final "+content, func(t *testing.T) {
			cfg, _ := mockModel(t, func(_ openai.ChatCompletionRequest, step int) openai.ChatCompletionMessage {
				if step == 1 {
					return openai.ChatCompletionMessage{ToolCalls: []openai.ToolCall{toolCall("env", "get_environment", `{}`)}}
				}
				return openai.ChatCompletionMessage{Content: content}
			})
			if _, err := GenerateCommandAgent(context.Background(), "test", cfg, shell.Context{PWD: t.TempDir()}, AgentOptions{}); err == nil {
				t.Fatal("accepted invalid final command")
			}
		})
	}
}

func TestAgentStopsWhenProviderIgnoresToolLimit(t *testing.T) {
	cfg, count := mockModel(t, func(openai.ChatCompletionRequest, int) openai.ChatCompletionMessage {
		return openai.ChatCompletionMessage{ToolCalls: []openai.ToolCall{toolCall("env", "get_environment", `{}`)}}
	})
	_, err := GenerateCommandAgent(context.Background(), "pwd", cfg, shell.Context{PWD: t.TempDir()}, AgentOptions{MaxRounds: 1})
	if err == nil || count.Load() != 2 || !strings.Contains(err.Error(), "exploration limit") {
		t.Fatalf("err=%v, calls=%d", err, count.Load())
	}
}

func TestAgentCancellationAndAPIErrors(t *testing.T) {
	t.Run("signal cancellation cause", func(t *testing.T) {
		ctx, cancel := context.WithCancelCause(context.Background())
		defer cancel(nil)
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cancel(errors.New("interrupt signal received"))
		}))
		defer server.Close()
		_, err := GenerateCommandAgent(ctx, "test", config.Config{APIKey: "test", BaseURL: server.URL, Model: "test"}, shell.Context{PWD: t.TempDir()}, AgentOptions{})
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("lost cancellation identity: %v", err)
		}
	})
	t.Run("cancelled before request", func(t *testing.T) {
		cfg, count := mockModel(t, func(openai.ChatCompletionRequest, int) openai.ChatCompletionMessage {
			return openai.ChatCompletionMessage{}
		})
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		_, err := GenerateCommandAgent(ctx, "test", cfg, shell.Context{PWD: t.TempDir()}, AgentOptions{})
		if !errors.Is(err, context.Canceled) || count.Load() != 0 {
			t.Fatalf("err=%v, calls=%d", err, count.Load())
		}
	})
	t.Run("deadline while waiting", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			select {
			case <-r.Context().Done():
			case <-time.After(time.Second):
			}
		}))
		defer server.Close()
		cfg := config.Config{APIKey: "test", BaseURL: server.URL, Model: "test"}
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()
		_, err := GenerateCommandAgent(ctx, "test", cfg, shell.Context{PWD: t.TempDir()}, AgentOptions{})
		if !errors.Is(err, context.DeadlineExceeded) {
			t.Fatalf("err=%v", err)
		}
	})
	for _, test := range []struct {
		name, body string
		status     int
	}{
		{"API failure", `{"error":{"message":"tools unsupported","type":"invalid_request_error"}}`, 400},
		{"empty choices", `{"choices":[]}`, 200},
		{"truncated", `{"choices":[{"finish_reason":"length","message":{"content":"{}"}}]}`, 200},
		{"filtered", `{"choices":[{"finish_reason":"content_filter","message":{"content":"{}"}}]}`, 200},
	} {
		t.Run(test.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(test.status)
				_, _ = w.Write([]byte(test.body))
			}))
			defer server.Close()
			_, err := GenerateCommandAgent(context.Background(), "test", config.Config{APIKey: "test", BaseURL: server.URL, Model: "test"}, shell.Context{PWD: t.TempDir()}, AgentOptions{})
			if err == nil {
				t.Fatal("expected request failure")
			}
		})
	}
}
