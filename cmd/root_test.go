package cmd

import (
	"io"
	"strings"
	"testing"

	"github.com/lucas/dcfm/internal/config"
	"github.com/lucas/dcfm/internal/i18n"
	"github.com/spf13/cobra"
)

func TestRejectsConflictingModesAndInvalidAgentLimits(t *testing.T) {
	t.Setenv("DCFM_LANGUAGE", "en")
	rootCmd.SetOut(io.Discard)
	rootCmd.SetErr(io.Discard)
	defer func() { rootCmd.SetArgs(nil); rootCmd.SetOut(nil); rootCmd.SetErr(nil); localizeCLI(i18n.English) }()
	reset := func() {
		for _, name := range []string{"agent", "direct", "config", "explain", "agent-max-rounds"} {
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
		{[]string{"--agent", "--direct", "test"}, "none of the others"},
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

func TestSavedModeAndOverrides(t *testing.T) {
	for _, test := range []struct {
		mode string
		args []string
		want bool
	}{
		{config.ModeDirect, nil, false},
		{config.ModeAgent, nil, true},
		{config.ModeDirect, []string{"--agent"}, true},
		{config.ModeAgent, []string{"--direct"}, false},
		{config.ModeAgent, []string{"--agent=false"}, false},
		{config.ModeAgent, []string{"--config"}, false},
		{config.ModeAgent, []string{"--explain"}, false},
	} {
		cmd := &cobra.Command{}
		for _, name := range []string{"agent", "direct", "config", "explain"} {
			cmd.Flags().Bool(name, false, "")
		}
		if err := cmd.ParseFlags(test.args); err != nil {
			t.Fatal(err)
		}
		if got := selectedAgentMode(cmd, config.Config{DefaultMode: test.mode}); got != test.want {
			t.Errorf("mode %s, args %v: got %v", test.mode, test.args, got)
		}
	}
}

func TestAgentLimitUsesSavedMode(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)
	t.Setenv("XDG_CONFIG_HOME", dir)
	if err := config.Save(config.Config{DefaultMode: config.ModeAgent}); err != nil {
		t.Fatal(err)
	}
	flag := rootCmd.Flags().Lookup("agent-max-rounds")
	defer func() { _ = flag.Value.Set(flag.DefValue); flag.Changed = false }()
	if err := rootCmd.Flags().Set("agent-max-rounds", "3"); err != nil {
		t.Fatal(err)
	}
	if err := rootCmd.PreRunE(rootCmd, nil); err != nil {
		t.Fatal(err)
	}
	if err := rootCmd.Flags().Set("agent-max-rounds", "0"); err != nil {
		t.Fatal(err)
	}
	if err := rootCmd.PreRunE(rootCmd, nil); err == nil {
		t.Fatal("accepted invalid rounds")
	}
}
