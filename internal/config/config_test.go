package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestEnvOr(t *testing.T) {
	key := "PINCHTAB_TEST_ENV"
	fallback := "default"

	_ = os.Unsetenv(key)
	if got := envOr(key, fallback); got != fallback {
		t.Errorf("envOr() = %v, want %v", got, fallback)
	}

	val := "set"
	_ = os.Setenv(key, val)
	defer func() { _ = os.Unsetenv(key) }()
	if got := envOr(key, fallback); got != val {
		t.Errorf("envOr() = %v, want %v", got, val)
	}
}

func TestEnvIntOr(t *testing.T) {
	key := "PINCHTAB_TEST_INT"
	fallback := 42

	_ = os.Unsetenv(key)
	if got := envIntOr(key, fallback); got != fallback {
		t.Errorf("envIntOr() = %v, want %v", got, fallback)
	}

	_ = os.Setenv(key, "100")
	if got := envIntOr(key, fallback); got != 100 {
		t.Errorf("envIntOr() = %v, want %v", got, 100)
	}

	_ = os.Setenv(key, "invalid")
	if got := envIntOr(key, fallback); got != fallback {
		t.Errorf("envIntOr() = %v, want %v", got, fallback)
	}
}

func TestEnvBoolOr(t *testing.T) {
	key := "PINCHTAB_TEST_BOOL"
	fallback := true

	_ = os.Unsetenv(key)
	if got := envBoolOr(key, fallback); got != fallback {
		t.Errorf("envBoolOr() = %v, want %v", got, fallback)
	}

	tests := []struct {
		val  string
		want bool
	}{
		{"1", true}, {"true", true}, {"yes", true}, {"on", true},
		{"0", false}, {"false", false}, {"no", false}, {"off", false},
		{"garbage", true}, // should return fallback
	}

	for _, tt := range tests {
		_ = os.Setenv(key, tt.val)
		if got := envBoolOr(key, fallback); got != tt.want {
			t.Errorf("envBoolOr(%q) = %v, want %v", tt.val, got, tt.want)
		}
	}
}

func TestMaskToken(t *testing.T) {
	tests := []struct {
		token string
		want  string
	}{
		{"", "(none)"},
		{"short", "***"},
		{"very-long-token-secret", "very...cret"},
	}

	for _, tt := range tests {
		if got := MaskToken(tt.token); got != tt.want {
			t.Errorf("MaskToken(%q) = %v, want %v", tt.token, got, tt.want)
		}
	}
}

func TestLoadConfigDefaults(t *testing.T) {
	clearConfigEnvVars(t)
	// Point to non-existent config to test pure defaults
	_ = os.Setenv("PINCHTAB_CONFIG", filepath.Join(t.TempDir(), "nonexistent.json"))
	defer func() { _ = os.Unsetenv("PINCHTAB_CONFIG") }()

	cfg := Load()
	if cfg.Port != "9867" {
		t.Errorf("default Port = %v, want 9867", cfg.Port)
	}
	if cfg.Bind != "127.0.0.1" {
		t.Errorf("default Bind = %v, want 127.0.0.1", cfg.Bind)
	}
	if cfg.AllowEvaluate {
		t.Errorf("default AllowEvaluate = %v, want false", cfg.AllowEvaluate)
	}
	if cfg.Strategy != "simple" {
		t.Errorf("default Strategy = %v, want simple", cfg.Strategy)
	}
	if cfg.AllocationPolicy != "fcfs" {
		t.Errorf("default AllocationPolicy = %v, want fcfs", cfg.AllocationPolicy)
	}
	if cfg.TabEvictionPolicy != "reject" {
		t.Errorf("default TabEvictionPolicy = %v, want reject", cfg.TabEvictionPolicy)
	}
}

func TestLoadConfigEnvOverrides(t *testing.T) {
	clearConfigEnvVars(t)
	// Point to non-existent config to isolate env var testing
	_ = os.Setenv("PINCHTAB_CONFIG", filepath.Join(t.TempDir(), "nonexistent.json"))
	_ = os.Setenv("PINCHTAB_PORT", "1234")
	_ = os.Setenv("PINCHTAB_ALLOW_EVALUATE", "1")
	_ = os.Setenv("PINCHTAB_STRATEGY", "explicit")
	defer func() {
		_ = os.Unsetenv("PINCHTAB_CONFIG")
		_ = os.Unsetenv("PINCHTAB_PORT")
		_ = os.Unsetenv("PINCHTAB_ALLOW_EVALUATE")
		_ = os.Unsetenv("PINCHTAB_STRATEGY")
	}()

	cfg := Load()
	if cfg.Port != "1234" {
		t.Errorf("env Port = %v, want 1234", cfg.Port)
	}
	if !cfg.AllowEvaluate {
		t.Errorf("env AllowEvaluate = %v, want true", cfg.AllowEvaluate)
	}
	if cfg.Strategy != "explicit" {
		t.Errorf("env Strategy = %v, want explicit", cfg.Strategy)
	}
}

func TestLegacyBridgeEnvFallback(t *testing.T) {
	clearConfigEnvVars(t)
	// Point to non-existent config to isolate env var testing
	_ = os.Setenv("PINCHTAB_CONFIG", filepath.Join(t.TempDir(), "nonexistent.json"))
	_ = os.Setenv("BRIDGE_PORT", "5555")
	_ = os.Setenv("BRIDGE_ALLOW_EVALUATE", "true")
	defer func() {
		_ = os.Unsetenv("PINCHTAB_CONFIG")
		_ = os.Unsetenv("BRIDGE_PORT")
		_ = os.Unsetenv("BRIDGE_ALLOW_EVALUATE")
	}()

	cfg := Load()
	if cfg.Port != "5555" {
		t.Errorf("legacy fallback Port = %v, want 5555", cfg.Port)
	}
	if !cfg.AllowEvaluate {
		t.Errorf("legacy fallback AllowEvaluate = %v, want true", cfg.AllowEvaluate)
	}
}

func TestPinchtabEnvTakesPrecedence(t *testing.T) {
	clearConfigEnvVars(t)
	// Point to non-existent config to isolate env var testing
	_ = os.Setenv("PINCHTAB_CONFIG", filepath.Join(t.TempDir(), "nonexistent.json"))
	_ = os.Setenv("PINCHTAB_PORT", "7777")
	_ = os.Setenv("BRIDGE_PORT", "8888")
	defer func() {
		_ = os.Unsetenv("PINCHTAB_CONFIG")
		_ = os.Unsetenv("PINCHTAB_PORT")
		_ = os.Unsetenv("BRIDGE_PORT")
	}()

	cfg := Load()
	if cfg.Port != "7777" {
		t.Errorf("precedence Port = %v, want 7777 (PINCHTAB_ should win)", cfg.Port)
	}
}

func TestDefaultFileConfig(t *testing.T) {
	fc := DefaultFileConfig()
	if fc.Server.Port != "9867" {
		t.Errorf("DefaultFileConfig.Server.Port = %v, want 9867", fc.Server.Port)
	}
	if fc.Server.Bind != "127.0.0.1" {
		t.Errorf("DefaultFileConfig.Server.Bind = %v, want 127.0.0.1", fc.Server.Bind)
	}
	if *fc.Chrome.Headless != true {
		t.Errorf("DefaultFileConfig.Chrome.Headless = %v, want true", *fc.Chrome.Headless)
	}
	if fc.Orchestrator.Strategy != "simple" {
		t.Errorf("DefaultFileConfig.Orchestrator.Strategy = %v, want simple", fc.Orchestrator.Strategy)
	}
}

// TestLoadNestedConfig tests loading the new nested config format.
func TestLoadNestedConfig(t *testing.T) {
	clearConfigEnvVars(t)

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.json")
	_ = os.Setenv("PINCHTAB_CONFIG", configPath)
	defer func() { _ = os.Unsetenv("PINCHTAB_CONFIG") }()

	// Create nested config file
	nestedConfig := `{
		"server": {
			"port": "8888",
			"bind": "0.0.0.0",
			"token": "secret123"
		},
		"chrome": {
			"headless": false,
			"maxTabs": 50,
			"stealthLevel": "full",
			"tabEvictionPolicy": "close_oldest"
		},
		"security": {
			"allowEvaluate": true
		},
		"orchestrator": {
			"strategy": "explicit",
			"allocationPolicy": "round_robin"
		},
		"timeouts": {
			"actionSec": 60,
			"navigateSec": 120
		}
	}`
	if err := os.WriteFile(configPath, []byte(nestedConfig), 0644); err != nil {
		t.Fatal(err)
	}

	cfg := Load()

	// Server
	if cfg.Port != "8888" {
		t.Errorf("nested Port = %v, want 8888", cfg.Port)
	}
	if cfg.Bind != "0.0.0.0" {
		t.Errorf("nested Bind = %v, want 0.0.0.0", cfg.Bind)
	}
	if cfg.Token != "secret123" {
		t.Errorf("nested Token = %v, want secret123", cfg.Token)
	}

	// Chrome
	if cfg.Headless != false {
		t.Errorf("nested Headless = %v, want false", cfg.Headless)
	}
	if cfg.MaxTabs != 50 {
		t.Errorf("nested MaxTabs = %v, want 50", cfg.MaxTabs)
	}
	if cfg.StealthLevel != "full" {
		t.Errorf("nested StealthLevel = %v, want full", cfg.StealthLevel)
	}
	if cfg.TabEvictionPolicy != "close_oldest" {
		t.Errorf("nested TabEvictionPolicy = %v, want close_oldest", cfg.TabEvictionPolicy)
	}

	// Security
	if cfg.AllowEvaluate != true {
		t.Errorf("nested AllowEvaluate = %v, want true", cfg.AllowEvaluate)
	}

	// Orchestrator
	if cfg.Strategy != "explicit" {
		t.Errorf("nested Strategy = %v, want explicit", cfg.Strategy)
	}
	if cfg.AllocationPolicy != "round_robin" {
		t.Errorf("nested AllocationPolicy = %v, want round_robin", cfg.AllocationPolicy)
	}

	// Timeouts
	if cfg.ActionTimeout != 60*time.Second {
		t.Errorf("nested ActionTimeout = %v, want 60s", cfg.ActionTimeout)
	}
	if cfg.NavigateTimeout != 120*time.Second {
		t.Errorf("nested NavigateTimeout = %v, want 120s", cfg.NavigateTimeout)
	}
}

// TestLoadLegacyFlatConfig tests backward compatibility with flat config files.
func TestLoadLegacyFlatConfig(t *testing.T) {
	clearConfigEnvVars(t)

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.json")
	_ = os.Setenv("PINCHTAB_CONFIG", configPath)
	defer func() { _ = os.Unsetenv("PINCHTAB_CONFIG") }()

	// Create legacy flat config file (old format)
	legacyConfig := `{
		"port": "7777",
		"headless": false,
		"maxTabs": 30,
		"allowEvaluate": true,
		"timeoutSec": 45,
		"navigateSec": 90
	}`
	if err := os.WriteFile(configPath, []byte(legacyConfig), 0644); err != nil {
		t.Fatal(err)
	}

	cfg := Load()

	if cfg.Port != "7777" {
		t.Errorf("legacy flat Port = %v, want 7777", cfg.Port)
	}
	if cfg.Headless != false {
		t.Errorf("legacy flat Headless = %v, want false", cfg.Headless)
	}
	if cfg.MaxTabs != 30 {
		t.Errorf("legacy flat MaxTabs = %v, want 30", cfg.MaxTabs)
	}
	if cfg.AllowEvaluate != true {
		t.Errorf("legacy flat AllowEvaluate = %v, want true", cfg.AllowEvaluate)
	}
	if cfg.ActionTimeout != 45*time.Second {
		t.Errorf("legacy flat ActionTimeout = %v, want 45s", cfg.ActionTimeout)
	}
	if cfg.NavigateTimeout != 90*time.Second {
		t.Errorf("legacy flat NavigateTimeout = %v, want 90s", cfg.NavigateTimeout)
	}
}

// TestEnvOverridesNestedConfig verifies env vars take precedence over nested config.
func TestEnvOverridesNestedConfig(t *testing.T) {
	clearConfigEnvVars(t)

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.json")
	_ = os.Setenv("PINCHTAB_CONFIG", configPath)
	_ = os.Setenv("PINCHTAB_PORT", "9999")
	_ = os.Setenv("PINCHTAB_STRATEGY", "simple-autorestart")
	defer func() {
		_ = os.Unsetenv("PINCHTAB_CONFIG")
		_ = os.Unsetenv("PINCHTAB_PORT")
		_ = os.Unsetenv("PINCHTAB_STRATEGY")
	}()

	// Config file says port 8888 and strategy explicit
	nestedConfig := `{
		"server": {
			"port": "8888"
		},
		"orchestrator": {
			"strategy": "explicit"
		}
	}`
	if err := os.WriteFile(configPath, []byte(nestedConfig), 0644); err != nil {
		t.Fatal(err)
	}

	cfg := Load()

	// Env var should win
	if cfg.Port != "9999" {
		t.Errorf("env should override file: Port = %v, want 9999", cfg.Port)
	}
	if cfg.Strategy != "simple-autorestart" {
		t.Errorf("env should override file: Strategy = %v, want simple-autorestart", cfg.Strategy)
	}
}

func TestListenAddr(t *testing.T) {
	cfg := &RuntimeConfig{Bind: "127.0.0.1", Port: "9867"}
	if got := cfg.ListenAddr(); got != "127.0.0.1:9867" {
		t.Errorf("expected 127.0.0.1:9867, got %s", got)
	}

	cfg = &RuntimeConfig{Bind: "0.0.0.0", Port: "8080"}
	if got := cfg.ListenAddr(); got != "0.0.0.0:8080" {
		t.Errorf("expected 0.0.0.0:8080, got %s", got)
	}
}

// TestIsLegacyConfig tests the format detection logic.
func TestIsLegacyConfig(t *testing.T) {
	tests := []struct {
		name     string
		json     string
		isLegacy bool
	}{
		{
			name:     "nested format with server",
			json:     `{"server": {"port": "9867"}}`,
			isLegacy: false,
		},
		{
			name:     "nested format with chrome",
			json:     `{"chrome": {"headless": true}}`,
			isLegacy: false,
		},
		{
			name:     "legacy format with port",
			json:     `{"port": "9867"}`,
			isLegacy: true,
		},
		{
			name:     "legacy format with headless",
			json:     `{"headless": true}`,
			isLegacy: true,
		},
		{
			name:     "empty object",
			json:     `{}`,
			isLegacy: false,
		},
		{
			name:     "mixed - nested wins",
			json:     `{"server": {"port": "8888"}, "port": "7777"}`,
			isLegacy: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isLegacyConfig([]byte(tt.json))
			if got != tt.isLegacy {
				t.Errorf("isLegacyConfig(%s) = %v, want %v", tt.json, got, tt.isLegacy)
			}
		})
	}
}

// TestConvertLegacyConfig tests the legacy to nested conversion.
func TestConvertLegacyConfig(t *testing.T) {
	h := false
	maxTabs := 25
	lc := &legacyFileConfig{
		Port:          "7777",
		Headless:      &h,
		MaxTabs:       &maxTabs,
		AllowEvaluate: boolPtr(true),
		TimeoutSec:    45,
		NavigateSec:   90,
	}

	fc := convertLegacyConfig(lc)

	if fc.Server.Port != "7777" {
		t.Errorf("converted Server.Port = %v, want 7777", fc.Server.Port)
	}
	if *fc.Chrome.Headless != false {
		t.Errorf("converted Chrome.Headless = %v, want false", *fc.Chrome.Headless)
	}
	if *fc.Chrome.MaxTabs != 25 {
		t.Errorf("converted Chrome.MaxTabs = %v, want 25", *fc.Chrome.MaxTabs)
	}
	if *fc.Security.AllowEvaluate != true {
		t.Errorf("converted Security.AllowEvaluate = %v, want true", *fc.Security.AllowEvaluate)
	}
	if fc.Timeouts.ActionSec != 45 {
		t.Errorf("converted Timeouts.ActionSec = %v, want 45", fc.Timeouts.ActionSec)
	}
	if fc.Timeouts.NavigateSec != 90 {
		t.Errorf("converted Timeouts.NavigateSec = %v, want 90", fc.Timeouts.NavigateSec)
	}
}

// TestDefaultFileConfigJSON tests that DefaultFileConfig serializes correctly.
func TestDefaultFileConfigJSON(t *testing.T) {
	fc := DefaultFileConfig()
	data, err := json.MarshalIndent(fc, "", "  ")
	if err != nil {
		t.Fatalf("failed to marshal DefaultFileConfig: %v", err)
	}

	// Verify it can be parsed back
	var parsed FileConfig
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("failed to unmarshal DefaultFileConfig output: %v", err)
	}

	if parsed.Server.Port != "9867" {
		t.Errorf("round-trip Server.Port = %v, want 9867", parsed.Server.Port)
	}
	if *parsed.Chrome.Headless != true {
		t.Errorf("round-trip Chrome.Headless = %v, want true", *parsed.Chrome.Headless)
	}
}

// Helper functions

func boolPtr(b bool) *bool {
	return &b
}

// clearConfigEnvVars unsets all config-related env vars for clean tests.
func clearConfigEnvVars(t *testing.T) {
	t.Helper()
	envVars := []string{
		"PINCHTAB_PORT", "PINCHTAB_BIND", "PINCHTAB_TOKEN", "PINCHTAB_CONFIG",
		"PINCHTAB_STATE_DIR", "PINCHTAB_PROFILE_DIR", "PINCHTAB_HEADLESS",
		"PINCHTAB_MAX_TABS", "PINCHTAB_ALLOW_EVALUATE", "PINCHTAB_ALLOW_MACRO",
		"PINCHTAB_ALLOW_SCREENCAST", "PINCHTAB_ALLOW_DOWNLOAD", "PINCHTAB_ALLOW_UPLOAD",
		"PINCHTAB_STRATEGY", "PINCHTAB_ALLOCATION_POLICY", "PINCHTAB_TAB_EVICTION_POLICY",
		"PINCHTAB_STEALTH", "PINCHTAB_NO_RESTORE", "PINCHTAB_NO_ANIMATIONS",
		"PINCHTAB_TIMEOUT", "PINCHTAB_NAV_TIMEOUT",
		"BRIDGE_PORT", "BRIDGE_BIND", "BRIDGE_TOKEN", "BRIDGE_CONFIG",
		"BRIDGE_STATE_DIR", "BRIDGE_PROFILE", "BRIDGE_HEADLESS",
		"BRIDGE_MAX_TABS", "BRIDGE_ALLOW_EVALUATE",
		"BRIDGE_STEALTH", "BRIDGE_NO_RESTORE", "BRIDGE_NO_ANIMATIONS",
		"BRIDGE_TIMEOUT", "BRIDGE_NAV_TIMEOUT",
		"CDP_URL", "CHROME_BIN", "CHROME_BINARY", "CHROME_FLAGS",
	}
	for _, v := range envVars {
		_ = os.Unsetenv(v)
	}
}
