package cmd

import (
	"bytes"
	"strings"
	"testing"

	"github.com/AlecAivazis/survey/v2"
	"github.com/lucas/dcfm/internal/i18n"
)

func TestChineseHelpAndMenuHints(t *testing.T) {
	defer localizeCLI(i18n.English)
	localizeCLI(i18n.Chinese)
	var output bytes.Buffer
	rootCmd.SetOut(&output)
	defer rootCmd.SetOut(nil)
	if err := rootCmd.Help(); err != nil {
		t.Fatal(err)
	}
	for _, text := range []string{"用法：", "选项：", "显示帮助", "解释传入的命令"} {
		if !strings.Contains(output.String(), text) {
			t.Fatalf("missing %q in %s", text, output.String())
		}
	}
	if strings.Contains(survey.SelectQuestionTemplate, "Use arrows") {
		t.Fatal("English menu hint remains")
	}
	localizeCLI(i18n.English)
	if !strings.Contains(survey.SelectQuestionTemplate, "Use arrows") {
		t.Fatal("English menu hint not restored")
	}
}
