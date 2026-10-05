package config

import "testing"

func TestLoadWithoutFile(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)
	t.Setenv("XDG_CONFIG_HOME", dir)
	for _, key := range []string{"DCFM_API_KEY", "DCFM_BASE_URL", "DCFM_MODEL", "DCFM_LANGUAGE", "DCFM_PROVIDER"} {
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

func TestDeepSeekProviderDefaultsAndOverrides(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)
	t.Setenv("XDG_CONFIG_HOME", dir)
	for _, key := range []string{"DCFM_API_KEY", "DCFM_BASE_URL", "DCFM_MODEL", "DCFM_LANGUAGE", "DCFM_PROVIDER"} {
		t.Setenv(key, "")
	}
	if err := Save(Config{Provider: ProviderDeepSeek, APIKey: "test", Language: "zh"}); err != nil {
		t.Fatal(err)
	}
	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Provider != ProviderDeepSeek || cfg.BaseURL != DeepSeekBaseURL || cfg.Model != DeepSeekModel || cfg.Language != "zh" {
		t.Fatalf("incorrect defaults: %+v", cfg)
	}
	t.Setenv("DCFM_BASE_URL", "https://example.test/v1")
	t.Setenv("DCFM_MODEL", "custom-model")
	cfg, err = Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.BaseURL != "https://example.test/v1" || cfg.Model != "custom-model" {
		t.Fatalf("overrides lost: %+v", cfg)
	}
}

func TestLegacyProviderConfigStaysUnchanged(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)
	t.Setenv("XDG_CONFIG_HOME", dir)
	for _, key := range []string{"DCFM_API_KEY", "DCFM_BASE_URL", "DCFM_MODEL", "DCFM_LANGUAGE", "DCFM_PROVIDER"} {
		t.Setenv(key, "")
	}
	want := Config{APIKey: "legacy-key", BaseURL: "https://example.test/v1", Model: "legacy-model", Language: "zh"}
	if err := Save(want); err != nil {
		t.Fatal(err)
	}
	got, err := Load()
	if err != nil || got != want {
		t.Fatalf("legacy config changed: %+v, %v", got, err)
	}
	t.Setenv("DCFM_PROVIDER", ProviderDeepSeek)
	got, err = Load()
	if err != nil || got.Provider != ProviderDeepSeek || got.BaseURL != want.BaseURL || got.Model != want.Model || got.APIKey != want.APIKey {
		t.Fatalf("provider override changed explicit values: %+v, %v", got, err)
	}
}
