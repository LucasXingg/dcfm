package cmd

import (
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/lucas/dcfm/internal/i18n"
)

var defaultSelectTemplate = survey.SelectQuestionTemplate

// Configure translations before Cobra handles help or interactive prompts.
func localizeCLI(lang i18n.Lang) {
	msg := i18n.GetMessages(lang)
	rootCmd.Short = msg.RootDescription
	rootCmd.Flags().Lookup("config").Usage = msg.ConfigFlag
	rootCmd.Flags().Lookup("explain").Usage = msg.ExplainFlag
	rootCmd.InitDefaultHelpFlag()
	rootCmd.Flags().Lookup("help").Usage = msg.HelpFlag
	rootCmd.CompletionOptions.DisableDefaultCmd = true
	rootCmd.SetHelpTemplate(`{{.Short}}

` + msg.HelpUsage + `
  {{.UseLine}}

` + msg.HelpFlags + `
{{.LocalFlags.FlagUsages}}`)
	survey.SelectQuestionTemplate = strings.ReplaceAll(defaultSelectTemplate,
		"Use arrows to move, type to filter", msg.SelectHint)
}
