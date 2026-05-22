package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/lucas/dcfm/internal/config"
	"github.com/lucas/dcfm/internal/llm"
	"github.com/lucas/dcfm/internal/shell"
)

type explainResponse struct {
	Explain     string `json:"explain"`
	Flags       map[string]string `json:"flags"`
	Suspicion   string `json:"suspicion"`
}

func runExplain(args []string) {

	var prompt string

	if len(args) != 0 {
		prompt = strings.Join(args, " ")
	} else {
		cmdPrompt := &survey.Input{
			Message: "Enter Command to Explain:",
		}
		
		err := survey.AskOne(cmdPrompt, &prompt)
		if err != nil {
			fmt.Println("No command provided.")
			return
		}

	}

	fmt.Println("Generating Explanation...")

	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		os.Exit(1)
	}

	shellCtx := shell.GetContext()
	ctx := context.Background()

	content, err := llm.GenerateCommand(ctx, prompt, cfg, shellCtx, llm.ExplainPromptTpl)

	if err != nil {
		fmt.Printf("Error explaining command: %v\n", err)
		os.Exit(1)
	}

	var llmResp explainResponse
	if err := json.Unmarshal([]byte(content), &llmResp); err != nil {
		fmt.Printf("failed to parse JSON response from LLM: %w. Raw content: %s", err, content)
		os.Exit(1)
	}

	fmt.Printf("\nExplanation: \n\n\033[1;36m%s\033[0m\n", llmResp.Explain)

	for flag, desc := range llmResp.Flags {
		fmt.Printf("\033[1;33m%s\033[0m: %s\n", flag, desc)
	}


	if llmResp.Suspicion != "" {
		fmt.Printf("\n\n\033[1;31mYour command may contains suspicious behavior:\033[0m \n%s\n", llmResp.Suspicion)
		fmt.Println("\n\033[4;90mLLM may be wrong, please doublecheck carefully.\033[0m")
	}
}
