package config

import "testing"

func TestLoadWithoutFile(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)
	t.Setenv("XDG_CONFIG_HOME", dir)
	for _, key := range []string{"DCFM_API_KEY", "DCFM_BASE_URL", "DCFM_MODEL", "DCFM_LANGUAGE"} {
		t.Setenv(key, "")
	}
	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Language != "en" || cfg.Model != "gpt-4o" || cfg.BaseURL == "" {
		t.Fatalf("missing defaults: %+v", cfg)
	}
	t.Setenv("DCFM_LANGUAGE", "zh")
	t.Setenv("DCFM_API_KEY", "test")
	cfg, err = Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Language != "zh" || cfg.APIKey != "test" {
		t.Fatalf("environment ignored: %+v", cfg)
	}
}
