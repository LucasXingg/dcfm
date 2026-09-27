package cmd

import (
	"io"
	"strings"
	"testing"

	"github.com/lucas/dcfm/internal/i18n"
)

func TestRejectsConflictingModesAndInvalidAgentLimits(t *testing.T) {
	t.Setenv("DCFM_LANGUAGE", "en")
	rootCmd.SetOut(io.Discard)
	rootCmd.SetErr(io.Discard)
	defer func() { rootCmd.SetArgs(nil); rootCmd.SetOut(nil); rootCmd.SetErr(nil); localizeCLI(i18n.English) }()
	reset := func() {
		for _, name := range []string{"agent", "config", "explain", "agent-max-rounds"} {
			flag := rootCmd.Flags().Lookup(name)
			if err := flag.Value.Set(flag.DefValue); err != nil {
				t.Fatal(err)
			}
			flag.Changed = false
		}
	}
	defer reset()
	for _, test := range []struct {
		args []string
		want string
	}{
		{[]string{"--agent", "--explain", "ls"}, "none of the others"},
		{[]string{"--agent", "--config"}, "none of the others"},
		{[]string{"--agent-max-rounds", "3", "test"}, "requires --agent"},
		{[]string{"--agent", "--agent-max-rounds", "0", "test"}, "between 1 and 20"},
		{[]string{"--agent", "--agent-max-rounds", "21", "test"}, "between 1 and 20"},
	} {
		reset()
		rootCmd.SetArgs(test.args)
		_, err := rootCmd.ExecuteC()
		if err == nil || !strings.Contains(err.Error(), test.want) {
			t.Errorf("%v: %v", test.args, err)
		}
	}
}
