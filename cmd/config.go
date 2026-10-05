package cmd

import (
	"fmt"

	"github.com/AlecAivazis/survey/v2"
	"github.com/lucas/dcfm/internal/config"
	"github.com/lucas/dcfm/internal/i18n"
)

func runConfig() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Printf(i18n.GetMessages(i18n.Lang(cfg.Language)).ConfigLoadWarning+"\n", err)
	}

	isFirstTime := cfg.APIKey == ""

	if isFirstTime {
		runFirstTimeConfig(cfg)
	} else {
		runSettingsMenu(cfg)
	}
}

func runFirstTimeConfig(cfg config.Config) {
	msg := i18n.GetMessages(i18n.Lang(cfg.Language))

	fmt.Printf("\n=== %s ===\n\n", msg.ConfigTitle)

	err := promptProvider(&cfg, msg)
	if err != nil {
		fmt.Println(msg.ConfigCancelled)
		return
	}

	err = promptAPIKey(&cfg, msg)
	if err != nil {
		fmt.Println(msg.ConfigCancelled)
		return
	}

	err = promptBaseURL(&cfg, msg)
	if err != nil {
		fmt.Println(msg.ConfigCancelled)
		return
	}

	err = promptModel(&cfg, msg)
	if err != nil {
		fmt.Println(msg.ConfigCancelled)
		return
	}

	err = promptLanguage(&cfg, msg)
	if err != nil {
		fmt.Println(msg.ConfigCancelled)
		return
	}

	saveConfig(cfg)
}

func runSettingsMenu(cfg config.Config) {
	msg := i18n.GetMessages(i18n.Lang(cfg.Language))

	for {
		fmt.Printf("\n=== %s ===\n\n", msg.ConfigTitle)

		options := []string{
			msg.ConfigSelectAll,
			msg.ConfigSelectProvider,
			msg.ConfigSelectAPIKey,
			msg.ConfigSelectBaseURL,
			msg.ConfigSelectModel,
			msg.ConfigSelectLanguage,
		}

		var selection string
		prompt := &survey.Select{
			Message: msg.ConfigSelectPrompt,
			Options: options,
		}
		err := survey.AskOne(prompt, &selection)
		if err != nil {
			fmt.Println(msg.ConfigCancelled)
			return
		}

		switch selection {
		case msg.ConfigSelectProvider:
			if err := promptProvider(&cfg, msg); err != nil {
				fmt.Println(msg.ConfigCancelled)
				return
			}
			if cfg.Provider == config.ProviderDeepSeek {
				if err := promptAPIKey(&cfg, msg); err != nil {
					fmt.Println(msg.ConfigCancelled)
					return
				}
			}
			saveConfig(cfg)
			return
		case msg.ConfigSelectAll:
			err := promptProvider(&cfg, msg)
			if err != nil {
				fmt.Println(msg.ConfigCancelled)
				return
			}
			err = promptAPIKey(&cfg, msg)
			if err != nil {
				fmt.Println(msg.ConfigCancelled)
				return
			}
			err = promptBaseURL(&cfg, msg)
			if err != nil {
				fmt.Println(msg.ConfigCancelled)
				return
			}
			err = promptModel(&cfg, msg)
			if err != nil {
				fmt.Println(msg.ConfigCancelled)
				return
			}
			err = promptLanguage(&cfg, msg)
			if err != nil {
				fmt.Println(msg.ConfigCancelled)
				return
			}
			saveConfig(cfg)
			return
		case msg.ConfigSelectAPIKey:
			err := promptAPIKey(&cfg, msg)
			if err != nil {
				fmt.Println(msg.ConfigCancelled)
				return
			}
			saveConfig(cfg)
			return
		case msg.ConfigSelectBaseURL:
			err := promptBaseURL(&cfg, msg)
			if err != nil {
				fmt.Println(msg.ConfigCancelled)
				return
			}
			saveConfig(cfg)
			return
		case msg.ConfigSelectModel:
			err := promptModel(&cfg, msg)
			if err != nil {
				fmt.Println(msg.ConfigCancelled)
				return
			}
			saveConfig(cfg)
			return
		case msg.ConfigSelectLanguage:
			err := promptLanguage(&cfg, msg)
			if err != nil {
				fmt.Println(msg.ConfigCancelled)
				return
			}
			saveConfig(cfg)
			return
		}
	}
}

func promptAPIKey(cfg *config.Config, msg i18n.Messages) error {
	apiKeyPrompt := &survey.Password{
		Message: msg.ConfigAPIKeyPrompt,
	}
	var apiKey string
	err := survey.AskOne(apiKeyPrompt, &apiKey)
	if err != nil {
		return err
	}
	if apiKey != "" {
		cfg.APIKey = apiKey
	}
	return nil
}

func promptBaseURL(cfg *config.Config, msg i18n.Messages) error {
	baseURLPrompt := &survey.Input{
		Message: msg.ConfigBaseURLPrompt,
		Default: cfg.BaseURL,
	}
	var baseURL string
	err := survey.AskOne(baseURLPrompt, &baseURL)
	if err != nil {
		return err
	}
	cfg.BaseURL = baseURL
	return nil
}

func promptModel(cfg *config.Config, msg i18n.Messages) error {
	modelPrompt := &survey.Input{
		Message: msg.ConfigModelPrompt,
		Default: cfg.Model,
	}
	var model string
	err := survey.AskOne(modelPrompt, &model)
	if err != nil {
		return err
	}
	cfg.Model = model
	return nil
}

func promptLanguage(cfg *config.Config, msg i18n.Messages) error {
	languagePrompt := &survey.Select{
		Message: msg.ConfigLanguagePrompt,
		Options: []string{msg.ConfigLanguageEnglish, msg.ConfigLanguageChinese},
		Default: i18n.GetLanguageName(i18n.Lang(cfg.Language), i18n.Lang(cfg.Language)),
	}
	var language string
	err := survey.AskOne(languagePrompt, &language)
	if err != nil {
		return err
	}
	if language == msg.ConfigLanguageChinese {
		cfg.Language = string(i18n.Chinese)
	} else {
		cfg.Language = string(i18n.English)
	}
	return nil
}

func saveConfig(cfg config.Config) {
	msg := i18n.GetMessages(i18n.Lang(cfg.Language))
	if err := config.Save(cfg); err != nil {
		fmt.Printf(msg.ConfigSaveError, err)
		return
	}
	fmt.Println(msg.ConfigSaved)
}

// Choosing custom leaves the existing endpoint, model and key untouched.
func promptProvider(cfg *config.Config, msg i18n.Messages) error {
	defaultOption := msg.ConfigProviderCustom
	if cfg.Provider == config.ProviderDeepSeek {
		defaultOption = msg.ConfigProviderDeepSeek
	}
	var selection string
	if err := survey.AskOne(&survey.Select{
		Message: msg.ConfigProviderPrompt,
		Options: []string{msg.ConfigProviderCustom, msg.ConfigProviderDeepSeek},
		Default: defaultOption,
	}, &selection); err != nil {
		return err
	}
	if selection == msg.ConfigProviderDeepSeek {
		if cfg.Provider != config.ProviderDeepSeek {
			cfg.APIKey = ""
			cfg.BaseURL = config.DeepSeekBaseURL
			cfg.Model = config.DeepSeekModel
		}
		cfg.Provider = config.ProviderDeepSeek
	} else {
		cfg.Provider = ""
	}
	return nil
}
