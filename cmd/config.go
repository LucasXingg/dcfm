package cmd

import (
	"fmt"

	"github.com/AlecAivazis/survey/v2"
	"github.com/lucas/dcfm/internal/config"
	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Configure dcfm API Key, Base URL, and Model",
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := config.Load()
		if err != nil {
			fmt.Printf("Warning: could not load existing config: %v\n", err)
		}

		// Prompt for API Key
		apiKeyPrompt := &survey.Password{
			Message: "Enter your API Key (hidden):",
		}
		var apiKey string
		err = survey.AskOne(apiKeyPrompt, &apiKey)
		if err != nil {
			fmt.Println("Configuration cancelled.")
			return
		}
		if apiKey != "" {
			cfg.APIKey = apiKey
		}

		// Prompt for Base URL
		baseURLPrompt := &survey.Input{
			Message: "Enter Base URL (OpenAI compatible):",
			Default: cfg.BaseURL,
		}
		var baseURL string
		err = survey.AskOne(baseURLPrompt, &baseURL)
		if err != nil {
			fmt.Println("Configuration cancelled.")
			return
		}
		cfg.BaseURL = baseURL

		// Prompt for Model
		modelPrompt := &survey.Input{
			Message: "Enter Model name:",
			Default: cfg.Model,
		}
		var model string
		err = survey.AskOne(modelPrompt, &model)
		if err != nil {
			fmt.Println("Configuration cancelled.")
			return
		}
		cfg.Model = model

		// Save Configuration
		if err := config.Save(cfg); err != nil {
			fmt.Printf("Error saving configuration: %v\n", err)
			return
		}

		fmt.Println("Configuration saved successfully.")
	},
}

func init() {
	rootCmd.AddCommand(configCmd)
}
