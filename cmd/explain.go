package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/lucas/dcfm/internal/config"
	"github.com/lucas/dcfm/internal/i18n"
	"github.com/lucas/dcfm/internal/llm"
	"github.com/lucas/dcfm/internal/shell"
)

type explainResponse struct {
	Explain   string            `json:"explain"`
	Flags     map[string]string `json:"flags"`
	Suspicion string            `json:"suspicion"`
}

func runExplain(args []string) {
	cfg, err := config.Load()
	msg := i18n.GetMessages(i18n.Lang(cfg.Language))
	if err != nil {
		fmt.Printf(msg.MainErrorLoadingConfig+"\n", err)
		os.Exit(1)
	}

	var prompt string

	if len(args) != 0 {
		prompt = strings.Join(args, " ")
	} else {
		cmdPrompt := &survey.Input{
			Message: msg.ExplainInput,
		}

		err := survey.AskOne(cmdPrompt, &prompt)
		if err != nil {
			fmt.Println(msg.ExplainEmpty)
			return
		}

	}

	if strings.TrimSpace(prompt) == "" {
		fmt.Println(msg.ExplainEmpty)
		return
	}
	fmt.Println(msg.ExplainGenerating)
	shellCtx := shell.GetContextWithLanguage(cfg.Language)
	ctx := context.Background()

	content, err := llm.GenerateCommand(ctx, prompt, cfg, shellCtx, llm.ExplainPromptTpl)

	if err != nil {
		fmt.Printf(msg.ExplainError+"\n", err)
		os.Exit(1)
	}

	var llmResp explainResponse
	if err := json.Unmarshal([]byte(content), &llmResp); err != nil {
		fmt.Printf(msg.MainErrorParsingResponse+"\n", err, content)
		os.Exit(1)
	}

	fmt.Printf("\n%s\n\n\033[1;36m%s\033[0m\n", msg.ExplainTitle, llmResp.Explain)

	for flag, desc := range llmResp.Flags {
		fmt.Printf("\033[1;33m%s\033[0m: %s\n", flag, desc)
	}

	if llmResp.Suspicion != "" {
		fmt.Printf("\n\n\033[1;31m%s\033[0m\n%s\n", msg.ExplainSuspicion, llmResp.Suspicion)
		fmt.Println("\n\033[4;90m" + msg.ExplainDisclaimer + "\033[0m")
	}
}
