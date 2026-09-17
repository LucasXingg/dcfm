package llm

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/lucas/dcfm/internal/config"
	"github.com/lucas/dcfm/internal/shell"
	"github.com/sashabaranov/go-openai"
)

func TestConfiguredLanguageReachesModel(t *testing.T) {
	for _, lang := range []string{"zh", "en"} {
		for name, tpl := range map[string]string{"generate": GenCmdPromptTpl, "explain": ExplainPromptTpl} {
			t.Run(lang+"/"+name, func(t *testing.T) {
				var request openai.ChatCompletionRequest
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
						t.Error(err)
					}
					w.Header().Set("Content-Type", "application/json")
					_, _ = w.Write([]byte(`{"choices":[{"message":{"role":"assistant","content":"{}"}}]}`))
				}))
				defer server.Close()
				// Deliberately supply a contradictory shell language: config must win.
				cfg := config.Config{APIKey: "test", BaseURL: server.URL, Model: "test", Language: lang}
				_, err := GenerateCommand(context.Background(), "explain ls -l", cfg, shell.Context{Language: "wrong"}, tpl)
				if err != nil {
					t.Fatal(err)
				}
				expected := "English"
				if lang == "zh" {
					expected = "Simplified Chinese (简体中文)"
				}
				prompt := request.Messages[0].Content
				if !strings.Contains(prompt, expected) || strings.Contains(prompt, "wrong") {
					t.Fatalf("incorrect prompt language: %s", prompt)
				}
				if name == "explain" && strings.Contains(prompt, `"explain": "go to my home directory"`) {
					t.Fatal("English example biases translated output")
				}
			})
		}
	}
}
