package cmd

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/atotto/clipboard"
	"github.com/lucas/dcfm/internal/config"
	"github.com/lucas/dcfm/internal/llm"
	"github.com/lucas/dcfm/internal/shell"
	"github.com/spf13/cobra"
)

var configFlag bool

func init() {
	rootCmd.Flags().BoolVarP(&configFlag, "config", "c", false, "Configure dcfm API Key, Base URL, and Model")
}

var rootCmd = &cobra.Command{
	Use:   "dcfm [prompt]",
	Short: "dcfm translates natural language into shell commands",
	Run: func(cmd *cobra.Command, args []string) {
		if configFlag {
			runConfig()
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

		shellCtx := shell.GetContext()
		ctx := context.Background()

		for {
			fmt.Println("Generating command...")
			llmResp, err := llm.GenerateCommand(ctx, prompt, cfg, shellCtx)
			if err != nil {
				fmt.Printf("Error generating command: %v\n", err)
				os.Exit(1)
			}

			fmt.Printf("\nProposed Command: \033[1;36m%s\033[0m\n\n", llmResp.Command)

			var userInput string
			promptOpts := &survey.Input{
				Message: "Execute (enter) / Cancel (q) / Edit (type message):",
			}
			if err := survey.AskOne(promptOpts, &userInput); err != nil {
				fmt.Println("Cancelled.")
				os.Exit(0)
			}

			userInput = strings.TrimSpace(userInput)

			if strings.ToLower(userInput) == "q" {
				fmt.Println("Cancelled.")
				os.Exit(0)
			} else if userInput == "" {
				if llmResp.ModifiesEnv {
					fmt.Println("\n\033[1;33mWarning: This command modifies the environment and cannot be executed directly by dcfm since changes won't persist to your shell.\033[0m")
					
					copyErr := clipboard.WriteAll(llmResp.Command)
					isRemote := os.Getenv("SSH_CLIENT") != "" || os.Getenv("SSH_TTY") != "" || os.Getenv("SSH_CONNECTION") != ""
					
					if isRemote {
						encoded := base64.StdEncoding.EncodeToString([]byte(llmResp.Command))
						fmt.Printf("\033]52;c;%s\a", encoded)
						copyErr = nil
					}

					if copyErr != nil {
						fmt.Printf("Failed to copy command to clipboard: %v\n", copyErr)
					} else if isRemote {
						fmt.Println("\033[1;32mThe command has been copied to your clipboard via OSC 52. Ensure your terminal emulator supports it. Paste it into your terminal to run it.\033[0m")
					} else {
						fmt.Println("\033[1;32mThe command has been copied to your clipboard. Paste it into your terminal to run it.\033[0m")
					}
					os.Exit(0)
				}

				err := shell.Execute(llmResp.Command)
				if err != nil {
					fmt.Printf("Command finished with error: %v\n", err)
					os.Exit(1)
				}
				os.Exit(0)
			} else {
				prompt = fmt.Sprintf("Previous request: %s\nProposed command: %s\nUser refinement: %s\nPlease provide the updated command based on the refinement.", prompt, llmResp.Command, userInput)
				// loop continues
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
