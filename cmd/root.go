package cmd

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/atotto/clipboard"
	"github.com/lucas/dcfm/internal/config"
	"github.com/lucas/dcfm/internal/i18n"
	"github.com/lucas/dcfm/internal/llm"
	"github.com/lucas/dcfm/internal/shell"
	"github.com/spf13/cobra"
)

var configFlag bool
var explainFlag bool

type cmdResponse struct {
	Command     string `json:"command"`
	ModifiesEnv bool   `json:"modifies_env"`
}

func init() {
	rootCmd.Flags().BoolVarP(&configFlag, "config", "c", false, "Configure dcfm API Key, Base URL, and Model")
	rootCmd.Flags().BoolVarP(&explainFlag, "explain", "e", false, "Explain the command you pass in")
}

var rootCmd = &cobra.Command{
	Use:   "dcfm [prompt]",
	Short: "dcfm translates natural language into shell commands",
	Run: func(cmd *cobra.Command, args []string) {
		if configFlag {
			runConfig()
			return
		}

		if explainFlag {
			runExplain(args)
			return
		}

		if len(args) == 0 {
			cmd.Help()
			os.Exit(0)
		}

		prompt := strings.Join(args, " ")
		cfg, err := config.Load()
		if err != nil {
			fmt.Printf("Error loading config: %v\n", err)
			os.Exit(1)
		}

		msg := i18n.GetMessages(i18n.Lang(cfg.Language))
		shellCtx := shell.GetContextWithLanguage(cfg.Language)
		ctx := context.Background()

		for {
			fmt.Println(msg.MainGenerating)
			content, err := llm.GenerateCommand(ctx, prompt, cfg, shellCtx, llm.GenCmdPromptTpl)

			if err != nil {
				fmt.Printf(i18n.Format(msg.MainErrorGeneratingCommand, err))
				os.Exit(1)
			}

			var llmResp cmdResponse
			if err := json.Unmarshal([]byte(content), &llmResp); err != nil {
				fmt.Printf(i18n.Format(msg.MainErrorParsingResponse, err, content))
				os.Exit(1)
			}

			fmt.Printf("\n%s \033[1;36m%s\033[0m\n\n", msg.MainProposedCommand, llmResp.Command)

			var userInput string
			promptOpts := &survey.Input{
				Message: msg.MainExecutePrompt,
			}
			if err := survey.AskOne(promptOpts, &userInput); err != nil {
				fmt.Println(msg.MainExecuteCancel)
				os.Exit(0)
			}

			userInput = strings.TrimSpace(userInput)

			if strings.ToLower(userInput) == "q" {
				fmt.Println(msg.MainExecuteCancel)
				os.Exit(0)
			} else if userInput == "" {
				if llmResp.ModifiesEnv {
					fmt.Println("\n\033[1;33m" + msg.MainWarningEnv + "\033[0m")
					
					copyErr := clipboard.WriteAll(llmResp.Command)
					isRemote := os.Getenv("SSH_CLIENT") != "" || os.Getenv("SSH_TTY") != "" || os.Getenv("SSH_CONNECTION") != ""
					
					if isRemote {
						encoded := base64.StdEncoding.EncodeToString([]byte(llmResp.Command))
						fmt.Printf("\033]52;c;%s\a", encoded)
						copyErr = nil
					}

					if copyErr != nil {
						fmt.Printf(i18n.Format(msg.MainClipboardError, copyErr))
					} else if isRemote {
						fmt.Println("\033[1;32m" + msg.MainClipboardCopiedRemote + "\033[0m")
					} else {
						fmt.Println("\033[1;32m" + msg.MainClipboardCopied + "\033[0m")
					}
					os.Exit(0)
				}

				err := shell.Execute(llmResp.Command)
				if err != nil {
					fmt.Printf(i18n.Format(msg.MainCommandError, err))
					os.Exit(1)
				}
				os.Exit(0)
			} else {
				if cfg.Language == string(i18n.Chinese) {
					prompt = fmt.Sprintf("上一个请求：%s\n建议的命令：%s\n用户修改：%s\n请根据修改提供更新后的命令。", prompt, llmResp.Command, userInput)
				} else {
					prompt = fmt.Sprintf("Previous request: %s\nProposed command: %s\nUser refinement: %s\nPlease provide the updated command based on the refinement.", prompt, llmResp.Command, userInput)
				}
			}
		}
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
	if len(os.Args) == 1 {
		rootCmd.Help()
		os.Exit(0)
	}

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}