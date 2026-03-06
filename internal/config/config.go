package config

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// RuntimeConfig holds all runtime settings used throughout the application.
// This is the single source of truth for configuration at runtime.
type RuntimeConfig struct {
	// Server settings
	Bind              string
	Port              string
	InstancePortStart int // Starting port for instances (default 9868)
	InstancePortEnd   int // Ending port for instances (default 9968)
	CdpURL            string
	Token             string
	StateDir          string

	// Security settings
	AllowEvaluate   bool
	AllowMacro      bool
	AllowScreencast bool
	AllowDownload   bool
	AllowUpload     bool

	// Chrome settings
	Headless          bool
	NoRestore         bool
	ProfileDir        string
	ChromeVersion     string
	Timezone          string
	BlockImages       bool
	BlockMedia        bool
	BlockAds          bool
	MaxTabs           int
	MaxParallelTabs   int // 0 = auto-detect from runtime.NumCPU
	ChromeBinary      string
	ChromeExtraFlags  string
	ExtensionPaths    []string
	UserAgent         string
	NoAnimations      bool
	StealthLevel      string
	TabEvictionPolicy string // "reject" (default), "close_oldest", "close_lru"

	// Timeout settings
	ActionTimeout   time.Duration
	NavigateTimeout time.Duration
	ShutdownTimeout time.Duration
	WaitNavDelay    time.Duration

	// Orchestrator settings (dashboard mode only)
	Strategy         string // "simple" (default), "explicit", or "simple-autorestart"
	AllocationPolicy string // "fcfs" (default), "round_robin", "random"
}

// --- Nested FileConfig structure (PR #91 model) ---

// FileConfig is the persistent configuration written to disk.
// Uses nested sections for organization: Server, Chrome, Orchestrator, Timeouts.
type FileConfig struct {
	Server       ServerConfig       `json:"server,omitempty"`
	Chrome       ChromeConfig       `json:"chrome,omitempty"`
	Security     SecurityConfig     `json:"security,omitempty"`
	Orchestrator OrchestratorConfig `json:"orchestrator,omitempty"`
	Timeouts     TimeoutsConfig     `json:"timeouts,omitempty"`
}

// ServerConfig holds server/network settings.
type ServerConfig struct {
	Port              string `json:"port,omitempty"`
	Bind              string `json:"bind,omitempty"`
	Token             string `json:"token,omitempty"`
	StateDir          string `json:"stateDir,omitempty"`
	CdpURL            string `json:"cdpUrl,omitempty"`
	InstancePortStart *int   `json:"instancePortStart,omitempty"`
	InstancePortEnd   *int   `json:"instancePortEnd,omitempty"`
}

// ChromeConfig holds browser/Chrome settings.
type ChromeConfig struct {
	Headless          *bool    `json:"headless,omitempty"`
	NoRestore         *bool    `json:"noRestore,omitempty"`
	ProfileDir        string   `json:"profileDir,omitempty"`
	ChromeVersion     string   `json:"chromeVersion,omitempty"`
	Timezone          string   `json:"timezone,omitempty"`
	BlockImages       *bool    `json:"blockImages,omitempty"`
	BlockMedia        *bool    `json:"blockMedia,omitempty"`
	BlockAds          *bool    `json:"blockAds,omitempty"`
	MaxTabs           *int     `json:"maxTabs,omitempty"`
	MaxParallelTabs   *int     `json:"maxParallelTabs,omitempty"`
	ChromeBinary      string   `json:"chromeBinary,omitempty"`
	ChromeExtraFlags  string   `json:"chromeExtraFlags,omitempty"`
	ExtensionPaths    []string `json:"extensionPaths,omitempty"`
	UserAgent         string   `json:"userAgent,omitempty"`
	NoAnimations      *bool    `json:"noAnimations,omitempty"`
	StealthLevel      string   `json:"stealthLevel,omitempty"`
	TabEvictionPolicy string   `json:"tabEvictionPolicy,omitempty"`
}

// SecurityConfig holds security/permission settings.
type SecurityConfig struct {
	AllowEvaluate   *bool `json:"allowEvaluate,omitempty"`
	AllowMacro      *bool `json:"allowMacro,omitempty"`
	AllowScreencast *bool `json:"allowScreencast,omitempty"`
	AllowDownload   *bool `json:"allowDownload,omitempty"`
	AllowUpload     *bool `json:"allowUpload,omitempty"`
}

// OrchestratorConfig holds orchestrator/strategy settings.
type OrchestratorConfig struct {
	Strategy         string `json:"strategy,omitempty"`
	AllocationPolicy string `json:"allocationPolicy,omitempty"`
}

// TimeoutsConfig holds timeout settings.
type TimeoutsConfig struct {
	ActionSec   int `json:"actionSec,omitempty"`
	NavigateSec int `json:"navigateSec,omitempty"`
	ShutdownSec int `json:"shutdownSec,omitempty"`
	WaitNavMs   int `json:"waitNavMs,omitempty"`
}

// --- Legacy flat FileConfig for backward compatibility ---

// legacyFileConfig is the old flat structure for backward compatibility.
type legacyFileConfig struct {
	Port              string `json:"port"`
	InstancePortStart *int   `json:"instancePortStart,omitempty"`
	InstancePortEnd   *int   `json:"instancePortEnd,omitempty"`
	CdpURL            string `json:"cdpUrl,omitempty"`
	Token             string `json:"token,omitempty"`
	AllowEvaluate     *bool  `json:"allowEvaluate,omitempty"`
	AllowMacro        *bool  `json:"allowMacro,omitempty"`
	AllowScreencast   *bool  `json:"allowScreencast,omitempty"`
	AllowDownload     *bool  `json:"allowDownload,omitempty"`
	AllowUpload       *bool  `json:"allowUpload,omitempty"`
	StateDir          string `json:"stateDir"`
	ProfileDir        string `json:"profileDir"`
	Headless          *bool  `json:"headless,omitempty"`
	NoRestore         bool   `json:"noRestore"`
	MaxTabs           *int   `json:"maxTabs,omitempty"`
	TimeoutSec        int    `json:"timeoutSec,omitempty"`
	NavigateSec       int    `json:"navigateSec,omitempty"`
}

// convertLegacyConfig converts flat config to nested structure.
func convertLegacyConfig(lc *legacyFileConfig) *FileConfig {
	fc := &FileConfig{}

	// Server
	fc.Server.Port = lc.Port
	fc.Server.CdpURL = lc.CdpURL
	fc.Server.Token = lc.Token
	fc.Server.StateDir = lc.StateDir
	fc.Server.InstancePortStart = lc.InstancePortStart
	fc.Server.InstancePortEnd = lc.InstancePortEnd

	// Chrome
	fc.Chrome.Headless = lc.Headless
	fc.Chrome.ProfileDir = lc.ProfileDir
	fc.Chrome.MaxTabs = lc.MaxTabs
	if lc.NoRestore {
		b := true
		fc.Chrome.NoRestore = &b
	}

	// Security
	fc.Security.AllowEvaluate = lc.AllowEvaluate
	fc.Security.AllowMacro = lc.AllowMacro
	fc.Security.AllowScreencast = lc.AllowScreencast
	fc.Security.AllowDownload = lc.AllowDownload
	fc.Security.AllowUpload = lc.AllowUpload

	// Timeouts
	fc.Timeouts.ActionSec = lc.TimeoutSec
	fc.Timeouts.NavigateSec = lc.NavigateSec

	return fc
}

// isLegacyConfig detects if JSON is flat (legacy) or nested (new).
// Returns true if it looks like legacy format.
func isLegacyConfig(data []byte) bool {
	var probe map[string]json.RawMessage
	if err := json.Unmarshal(data, &probe); err != nil {
		return false
	}

	// If "server" or "chrome" keys exist, it's new format
	if _, hasServer := probe["server"]; hasServer {
		return false
	}
	if _, hasChrome := probe["chrome"]; hasChrome {
		return false
	}

	// If "port" or "headless" exist at top level, it's legacy
	if _, hasPort := probe["port"]; hasPort {
		return true
	}
	if _, hasHeadless := probe["headless"]; hasHeadless {
		return true
	}

	// Default to new format for empty/unknown
	return false
}

// --- Environment variable helpers ---

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func envIntOr(key string, fallback int) int {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	n, err := strconv.Atoi(v)
	if err != nil || n < 0 {
		return fallback
	}
	return n
}

// splitCommaPaths splits a comma-separated string into non-empty trimmed paths.
func splitCommaPaths(s string) []string {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func envBoolOr(key string, fallback bool) bool {
	v, ok := os.LookupEnv(key)
	if !ok {
		return fallback
	}
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "1", "true", "yes", "on":
		return true
	case "0", "false", "no", "off":
		return false
	default:
		return fallback
	}
}

func envMigrate(newKey, oldKey string) string {
	if v := os.Getenv(newKey); v != "" {
		return v
	}
	if v := os.Getenv(oldKey); v != "" {
		slog.Warn("deprecated env var, use "+newKey+" instead", "var", oldKey)
		return v
	}
	return ""
}

func envOrMigrate(newKey, oldKey, fallback string) string {
	if v := envMigrate(newKey, oldKey); v != "" {
		return v
	}
	return fallback
}

func envIntOrMigrate(newKey, oldKey string, fallback int) int {
	v := envMigrate(newKey, oldKey)
	if v == "" {
		return fallback
	}
	n, err := strconv.Atoi(v)
	if err != nil || n < 0 {
		return fallback
	}
	return n
}

func envBoolOrMigrate(newKey, oldKey string, fallback bool) bool {
	if v, ok := os.LookupEnv(newKey); ok {
		switch strings.ToLower(strings.TrimSpace(v)) {
		case "1", "true", "yes", "on":
			return true
		case "0", "false", "no", "off":
			return false
		default:
			return fallback
		}
	}
	if v, ok := os.LookupEnv(oldKey); ok {
		slog.Warn("deprecated env var, use "+newKey+" instead", "var", oldKey)
		switch strings.ToLower(strings.TrimSpace(v)) {
		case "1", "true", "yes", "on":
			return true
		case "0", "false", "no", "off":
			return false
		default:
			return fallback
		}
	}
	return fallback
}

func envMigrateIsSet(newKey, oldKey string) bool {
	if os.Getenv(newKey) != "" {
		return true
	}
	return os.Getenv(oldKey) != ""
}

// homeDir returns the user's home directory, checking $HOME first for container compatibility
func homeDir() string {
	if home := os.Getenv("HOME"); home != "" {
		return home
	}
	h, _ := os.UserHomeDir()
	return h
}

// userConfigDir returns the OS-appropriate app config directory:
// - macOS: ~/Library/Application Support/pinchtab
// - Linux: ~/.config/pinchtab (or $XDG_CONFIG_HOME/pinchtab)
// - Windows: %APPDATA%\pinchtab
//
// For backwards compatibility, if ~/.pinchtab exists and the new location
// doesn't, it returns ~/.pinchtab (allowing seamless migration).
func userConfigDir() string {
	home := homeDir()
	legacyPath := filepath.Join(home, ".pinchtab")

	// Try to get OS-appropriate config directory
	configDir, err := os.UserConfigDir()
	if err != nil {
		// Fallback to legacy location if UserConfigDir fails
		return legacyPath
	}

	newPath := filepath.Join(configDir, "pinchtab")

	// Backwards compatibility: if legacy location exists and new doesn't, use legacy
	legacyExists := dirExists(legacyPath)
	newExists := dirExists(newPath)

	if legacyExists && !newExists {
		return legacyPath
	}

	return newPath
}

// dirExists checks if a directory exists
func dirExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}

func (c *RuntimeConfig) ListenAddr() string {
	return c.Bind + ":" + c.Port
}

// Load returns the RuntimeConfig with precedence: env vars > config file > defaults.
func Load() *RuntimeConfig {
	cfg := &RuntimeConfig{
		// Server defaults + env vars
		Bind:              envOrMigrate("PINCHTAB_BIND", "BRIDGE_BIND", "127.0.0.1"),
		Port:              envOrMigrate("PINCHTAB_PORT", "BRIDGE_PORT", "9867"),
		InstancePortStart: envIntOr("PINCHTAB_INSTANCE_PORT_START", 9868),
		InstancePortEnd:   envIntOr("PINCHTAB_INSTANCE_PORT_END", 9968),
		CdpURL:            os.Getenv("CDP_URL"),
		Token:             envMigrate("PINCHTAB_TOKEN", "BRIDGE_TOKEN"),
		StateDir:          envOrMigrate("PINCHTAB_STATE_DIR", "BRIDGE_STATE_DIR", userConfigDir()),

		// Security defaults + env vars
		AllowEvaluate:   envBoolOrMigrate("PINCHTAB_ALLOW_EVALUATE", "BRIDGE_ALLOW_EVALUATE", false),
		AllowMacro:      envBoolOrMigrate("PINCHTAB_ALLOW_MACRO", "BRIDGE_ALLOW_MACRO", false),
		AllowScreencast: envBoolOrMigrate("PINCHTAB_ALLOW_SCREENCAST", "BRIDGE_ALLOW_SCREENCAST", false),
		AllowDownload:   envBoolOrMigrate("PINCHTAB_ALLOW_DOWNLOAD", "BRIDGE_ALLOW_DOWNLOAD", false),
		AllowUpload:     envBoolOrMigrate("PINCHTAB_ALLOW_UPLOAD", "BRIDGE_ALLOW_UPLOAD", false),

		// Chrome defaults + env vars
		Headless:          envBoolOrMigrate("PINCHTAB_HEADLESS", "BRIDGE_HEADLESS", true),
		NoRestore:         envBoolOrMigrate("PINCHTAB_NO_RESTORE", "BRIDGE_NO_RESTORE", false),
		ProfileDir:        envOrMigrate("PINCHTAB_PROFILE_DIR", "BRIDGE_PROFILE", filepath.Join(userConfigDir(), "chrome-profile")),
		ChromeVersion:     envOrMigrate("PINCHTAB_CHROME_VERSION", "BRIDGE_CHROME_VERSION", "144.0.7559.133"),
		Timezone:          envMigrate("PINCHTAB_TIMEZONE", "BRIDGE_TIMEZONE"),
		BlockImages:       envBoolOrMigrate("PINCHTAB_BLOCK_IMAGES", "BRIDGE_BLOCK_IMAGES", false),
		BlockMedia:        envBoolOrMigrate("PINCHTAB_BLOCK_MEDIA", "BRIDGE_BLOCK_MEDIA", false),
		BlockAds:          envBoolOrMigrate("PINCHTAB_BLOCK_ADS", "BRIDGE_BLOCK_ADS", false),
		MaxTabs:           envIntOrMigrate("PINCHTAB_MAX_TABS", "BRIDGE_MAX_TABS", 20),
		MaxParallelTabs:   envIntOr("PINCHTAB_MAX_PARALLEL_TABS", 0),
		ChromeBinary:      envOr("CHROME_BIN", os.Getenv("CHROME_BINARY")),
		ChromeExtraFlags:  os.Getenv("CHROME_FLAGS"),
		ExtensionPaths:    splitCommaPaths(os.Getenv("CHROME_EXTENSION_PATHS")),
		UserAgent:         envMigrate("PINCHTAB_USER_AGENT", "BRIDGE_USER_AGENT"),
		NoAnimations:      envBoolOrMigrate("PINCHTAB_NO_ANIMATIONS", "BRIDGE_NO_ANIMATIONS", false),
		StealthLevel:      envOrMigrate("PINCHTAB_STEALTH", "BRIDGE_STEALTH", "light"),
		TabEvictionPolicy: envOr("PINCHTAB_TAB_EVICTION_POLICY", "reject"),

		// Timeout defaults
		ActionTimeout:   30 * time.Second,
		NavigateTimeout: 60 * time.Second,
		ShutdownTimeout: 10 * time.Second,
		WaitNavDelay:    1 * time.Second,

		// Orchestrator defaults + env vars
		Strategy:         envOr("PINCHTAB_STRATEGY", "simple"),
		AllocationPolicy: envOr("PINCHTAB_ALLOCATION_POLICY", "fcfs"),
	}

	// Load config file (supports both legacy flat and new nested format)
	configPath := envOrMigrate("PINCHTAB_CONFIG", "BRIDGE_CONFIG", filepath.Join(userConfigDir(), "config.json"))

	data, err := os.ReadFile(configPath)
	if err != nil {
		return cfg
	}

	var fc *FileConfig

	if isLegacyConfig(data) {
		var lc legacyFileConfig
		if err := json.Unmarshal(data, &lc); err != nil {
			return cfg
		}
		fc = convertLegacyConfig(&lc)
		slog.Debug("loaded legacy flat config, consider migrating to nested format")
	} else {
		fc = &FileConfig{}
		if err := json.Unmarshal(data, fc); err != nil {
			return cfg
		}
	}

	// Apply file config (only if env var NOT set)
	applyFileConfig(cfg, fc)

	return cfg
}

// applyFileConfig merges FileConfig values into RuntimeConfig.
// Only applies values where the corresponding env var is NOT set.
func applyFileConfig(cfg *RuntimeConfig, fc *FileConfig) {
	// Server
	if fc.Server.Port != "" && !envMigrateIsSet("PINCHTAB_PORT", "BRIDGE_PORT") {
		cfg.Port = fc.Server.Port
	}
	if fc.Server.Bind != "" && !envMigrateIsSet("PINCHTAB_BIND", "BRIDGE_BIND") {
		cfg.Bind = fc.Server.Bind
	}
	if fc.Server.Token != "" && !envMigrateIsSet("PINCHTAB_TOKEN", "BRIDGE_TOKEN") {
		cfg.Token = fc.Server.Token
	}
	if fc.Server.StateDir != "" && !envMigrateIsSet("PINCHTAB_STATE_DIR", "BRIDGE_STATE_DIR") {
		cfg.StateDir = fc.Server.StateDir
	}
	if fc.Server.CdpURL != "" && os.Getenv("CDP_URL") == "" {
		cfg.CdpURL = fc.Server.CdpURL
	}
	if fc.Server.InstancePortStart != nil && os.Getenv("PINCHTAB_INSTANCE_PORT_START") == "" {
		cfg.InstancePortStart = *fc.Server.InstancePortStart
	}
	if fc.Server.InstancePortEnd != nil && os.Getenv("PINCHTAB_INSTANCE_PORT_END") == "" {
		cfg.InstancePortEnd = *fc.Server.InstancePortEnd
	}

	// Security
	if fc.Security.AllowEvaluate != nil && !envMigrateIsSet("PINCHTAB_ALLOW_EVALUATE", "BRIDGE_ALLOW_EVALUATE") {
		cfg.AllowEvaluate = *fc.Security.AllowEvaluate
	}
	if fc.Security.AllowMacro != nil && !envMigrateIsSet("PINCHTAB_ALLOW_MACRO", "BRIDGE_ALLOW_MACRO") {
		cfg.AllowMacro = *fc.Security.AllowMacro
	}
	if fc.Security.AllowScreencast != nil && !envMigrateIsSet("PINCHTAB_ALLOW_SCREENCAST", "BRIDGE_ALLOW_SCREENCAST") {
		cfg.AllowScreencast = *fc.Security.AllowScreencast
	}
	if fc.Security.AllowDownload != nil && !envMigrateIsSet("PINCHTAB_ALLOW_DOWNLOAD", "BRIDGE_ALLOW_DOWNLOAD") {
		cfg.AllowDownload = *fc.Security.AllowDownload
	}
	if fc.Security.AllowUpload != nil && !envMigrateIsSet("PINCHTAB_ALLOW_UPLOAD", "BRIDGE_ALLOW_UPLOAD") {
		cfg.AllowUpload = *fc.Security.AllowUpload
	}

	// Chrome
	if fc.Chrome.Headless != nil && !envMigrateIsSet("PINCHTAB_HEADLESS", "BRIDGE_HEADLESS") {
		cfg.Headless = *fc.Chrome.Headless
	}
	if fc.Chrome.NoRestore != nil && !envMigrateIsSet("PINCHTAB_NO_RESTORE", "BRIDGE_NO_RESTORE") {
		cfg.NoRestore = *fc.Chrome.NoRestore
	}
	if fc.Chrome.ProfileDir != "" && !envMigrateIsSet("PINCHTAB_PROFILE_DIR", "BRIDGE_PROFILE") {
		cfg.ProfileDir = fc.Chrome.ProfileDir
	}
	if fc.Chrome.ChromeVersion != "" && !envMigrateIsSet("PINCHTAB_CHROME_VERSION", "BRIDGE_CHROME_VERSION") {
		cfg.ChromeVersion = fc.Chrome.ChromeVersion
	}
	if fc.Chrome.Timezone != "" && !envMigrateIsSet("PINCHTAB_TIMEZONE", "BRIDGE_TIMEZONE") {
		cfg.Timezone = fc.Chrome.Timezone
	}
	if fc.Chrome.BlockImages != nil && !envMigrateIsSet("PINCHTAB_BLOCK_IMAGES", "BRIDGE_BLOCK_IMAGES") {
		cfg.BlockImages = *fc.Chrome.BlockImages
	}
	if fc.Chrome.BlockMedia != nil && !envMigrateIsSet("PINCHTAB_BLOCK_MEDIA", "BRIDGE_BLOCK_MEDIA") {
		cfg.BlockMedia = *fc.Chrome.BlockMedia
	}
	if fc.Chrome.BlockAds != nil && !envMigrateIsSet("PINCHTAB_BLOCK_ADS", "BRIDGE_BLOCK_ADS") {
		cfg.BlockAds = *fc.Chrome.BlockAds
	}
	if fc.Chrome.MaxTabs != nil && !envMigrateIsSet("PINCHTAB_MAX_TABS", "BRIDGE_MAX_TABS") {
		cfg.MaxTabs = *fc.Chrome.MaxTabs
	}
	if fc.Chrome.MaxParallelTabs != nil && os.Getenv("PINCHTAB_MAX_PARALLEL_TABS") == "" {
		cfg.MaxParallelTabs = *fc.Chrome.MaxParallelTabs
	}
	if fc.Chrome.ChromeBinary != "" && os.Getenv("CHROME_BIN") == "" && os.Getenv("CHROME_BINARY") == "" {
		cfg.ChromeBinary = fc.Chrome.ChromeBinary
	}
	if fc.Chrome.ChromeExtraFlags != "" && os.Getenv("CHROME_FLAGS") == "" {
		cfg.ChromeExtraFlags = fc.Chrome.ChromeExtraFlags
	}
	if len(fc.Chrome.ExtensionPaths) > 0 && os.Getenv("CHROME_EXTENSION_PATHS") == "" {
		cfg.ExtensionPaths = fc.Chrome.ExtensionPaths
	}
	if fc.Chrome.UserAgent != "" && !envMigrateIsSet("PINCHTAB_USER_AGENT", "BRIDGE_USER_AGENT") {
		cfg.UserAgent = fc.Chrome.UserAgent
	}
	if fc.Chrome.NoAnimations != nil && !envMigrateIsSet("PINCHTAB_NO_ANIMATIONS", "BRIDGE_NO_ANIMATIONS") {
		cfg.NoAnimations = *fc.Chrome.NoAnimations
	}
	if fc.Chrome.StealthLevel != "" && !envMigrateIsSet("PINCHTAB_STEALTH", "BRIDGE_STEALTH") {
		cfg.StealthLevel = fc.Chrome.StealthLevel
	}
	if fc.Chrome.TabEvictionPolicy != "" && os.Getenv("PINCHTAB_TAB_EVICTION_POLICY") == "" {
		cfg.TabEvictionPolicy = fc.Chrome.TabEvictionPolicy
	}

	// Orchestrator
	if fc.Orchestrator.Strategy != "" && os.Getenv("PINCHTAB_STRATEGY") == "" {
		cfg.Strategy = fc.Orchestrator.Strategy
	}
	if fc.Orchestrator.AllocationPolicy != "" && os.Getenv("PINCHTAB_ALLOCATION_POLICY") == "" {
		cfg.AllocationPolicy = fc.Orchestrator.AllocationPolicy
	}

	// Timeouts
	if fc.Timeouts.ActionSec > 0 && !envMigrateIsSet("PINCHTAB_TIMEOUT", "BRIDGE_TIMEOUT") {
		cfg.ActionTimeout = time.Duration(fc.Timeouts.ActionSec) * time.Second
	}
	if fc.Timeouts.NavigateSec > 0 && !envMigrateIsSet("PINCHTAB_NAV_TIMEOUT", "BRIDGE_NAV_TIMEOUT") {
		cfg.NavigateTimeout = time.Duration(fc.Timeouts.NavigateSec) * time.Second
	}
	if fc.Timeouts.ShutdownSec > 0 && os.Getenv("PINCHTAB_SHUTDOWN_TIMEOUT") == "" {
		cfg.ShutdownTimeout = time.Duration(fc.Timeouts.ShutdownSec) * time.Second
	}
	if fc.Timeouts.WaitNavMs > 0 && os.Getenv("PINCHTAB_WAIT_NAV_DELAY") == "" {
		cfg.WaitNavDelay = time.Duration(fc.Timeouts.WaitNavMs) * time.Millisecond
	}
}

// DefaultFileConfig returns a FileConfig with sensible defaults (nested format).
func DefaultFileConfig() FileConfig {
	h := true
	start := 9868
	end := 9968
	maxTabs := 20
	return FileConfig{
		Server: ServerConfig{
			Port:              "9867",
			Bind:              "127.0.0.1",
			StateDir:          userConfigDir(),
			InstancePortStart: &start,
			InstancePortEnd:   &end,
		},
		Chrome: ChromeConfig{
			Headless:          &h,
			ProfileDir:        filepath.Join(userConfigDir(), "chrome-profile"),
			MaxTabs:           &maxTabs,
			StealthLevel:      "light",
			TabEvictionPolicy: "reject",
		},
		Orchestrator: OrchestratorConfig{
			Strategy:         "simple",
			AllocationPolicy: "fcfs",
		},
		Timeouts: TimeoutsConfig{
			ActionSec:   30,
			NavigateSec: 60,
			ShutdownSec: 10,
			WaitNavMs:   1000,
		},
	}
}

// HandleConfigCommand handles `pinchtab config <subcommand>`.
func HandleConfigCommand(cfg *RuntimeConfig) {
	if len(os.Args) < 3 {
		printConfigUsage()
		return
	}

	switch os.Args[2] {
	case "init":
		handleConfigInit()
	case "show":
		handleConfigShow(cfg)
	case "path":
		handleConfigPath()
	case "validate":
		handleConfigValidate()
	case "set":
		handleConfigSet()
	case "patch":
		handleConfigPatch()
	default:
		fmt.Printf("Unknown command: %s\n", os.Args[2])
		printConfigUsage()
		os.Exit(1)
	}
}

func printConfigUsage() {
	fmt.Println("Usage: pinchtab config <command>")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  init              Create default config file")
	fmt.Println("  show              Show current configuration")
	fmt.Println("  path              Show config file path")
	fmt.Println("  validate          Validate config file")
	fmt.Println("  set <path> <val>  Set a config value (e.g., server.port 8080)")
	fmt.Println("  patch <json>      Merge JSON into config")
}

func handleConfigInit() {
	configPath := filepath.Join(userConfigDir(), "config.json")

	if _, err := os.Stat(configPath); err == nil {
		fmt.Printf("Config file already exists at %s\n", configPath)
		fmt.Print("Overwrite? (y/N): ")
		var response string
		_, _ = fmt.Scanln(&response)
		if response != "y" && response != "Y" {
			return
		}
	}

	if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
		fmt.Printf("Error creating directory: %v\n", err)
		os.Exit(1)
	}

	fc := DefaultFileConfig()
	data, _ := json.MarshalIndent(fc, "", "  ")
	if err := os.WriteFile(configPath, data, 0644); err != nil {
		fmt.Printf("Error writing config: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Config file created at %s\n", configPath)
	fmt.Println()
	fmt.Println("Example with auth token:")
	fmt.Println(`{
  "server": {
    "port": "9867",
    "token": "your-secret-token"
  },
  "chrome": {
    "headless": true,
    "maxTabs": 20
  }
}`)
}

func handleConfigShow(cfg *RuntimeConfig) {
	fmt.Println("Current configuration (env > file > defaults):")
	fmt.Println()
	fmt.Println("Server:")
	fmt.Printf("  Port:           %s\n", cfg.Port)
	fmt.Printf("  Bind:           %s\n", cfg.Bind)
	fmt.Printf("  Token:          %s\n", MaskToken(cfg.Token))
	fmt.Printf("  State Dir:      %s\n", cfg.StateDir)
	fmt.Printf("  CDP URL:        %s\n", valueOrNone(cfg.CdpURL))
	fmt.Printf("  Instance Ports: %d-%d\n", cfg.InstancePortStart, cfg.InstancePortEnd)
	fmt.Println()
	fmt.Println("Security:")
	fmt.Printf("  Evaluate:       %v\n", cfg.AllowEvaluate)
	fmt.Printf("  Macro:          %v\n", cfg.AllowMacro)
	fmt.Printf("  Screencast:     %v\n", cfg.AllowScreencast)
	fmt.Printf("  Download:       %v\n", cfg.AllowDownload)
	fmt.Printf("  Upload:         %v\n", cfg.AllowUpload)
	fmt.Println()
	fmt.Println("Chrome:")
	fmt.Printf("  Headless:       %v\n", cfg.Headless)
	fmt.Printf("  No Restore:     %v\n", cfg.NoRestore)
	fmt.Printf("  Profile Dir:    %s\n", cfg.ProfileDir)
	fmt.Printf("  Max Tabs:       %d\n", cfg.MaxTabs)
	fmt.Printf("  Stealth:        %s\n", cfg.StealthLevel)
	fmt.Printf("  Tab Eviction:   %s\n", cfg.TabEvictionPolicy)
	fmt.Printf("  Extensions:     %v\n", cfg.ExtensionPaths)
	fmt.Println()
	fmt.Println("Orchestrator:")
	fmt.Printf("  Strategy:       %s\n", cfg.Strategy)
	fmt.Printf("  Allocation:     %s\n", cfg.AllocationPolicy)
	fmt.Println()
	fmt.Println("Timeouts:")
	fmt.Printf("  Action:         %v\n", cfg.ActionTimeout)
	fmt.Printf("  Navigate:       %v\n", cfg.NavigateTimeout)
	fmt.Printf("  Shutdown:       %v\n", cfg.ShutdownTimeout)
}

func handleConfigPath() {
	configPath := envOrMigrate("PINCHTAB_CONFIG", "BRIDGE_CONFIG", filepath.Join(userConfigDir(), "config.json"))
	fmt.Println(configPath)
}

func handleConfigSet() {
	if len(os.Args) < 5 {
		fmt.Println("Usage: pinchtab config set <path> <value>")
		fmt.Println()
		fmt.Println("Examples:")
		fmt.Println("  pinchtab config set server.port 8080")
		fmt.Println("  pinchtab config set chrome.headless false")
		fmt.Println("  pinchtab config set orchestrator.strategy explicit")
		fmt.Println()
		fmt.Println("Sections: server, chrome, security, orchestrator, timeouts")
		os.Exit(1)
	}

	path := os.Args[3]
	value := os.Args[4]

	fc, configPath, err := LoadFileConfig()
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		os.Exit(1)
	}

	if err := SetConfigValue(fc, path, value); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	// Validate before saving
	if errs := ValidateFileConfig(fc); len(errs) > 0 {
		fmt.Printf("Warning: new value causes validation error(s):\n")
		for _, e := range errs {
			fmt.Printf("  - %v\n", e)
		}
		fmt.Print("Save anyway? (y/N): ")
		var response string
		_, _ = fmt.Scanln(&response)
		if response != "y" && response != "Y" {
			fmt.Println("Aborted.")
			return
		}
	}

	if err := SaveFileConfig(fc, configPath); err != nil {
		fmt.Printf("Error saving config: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✓ Set %s = %s\n", path, value)
}

func handleConfigPatch() {
	if len(os.Args) < 4 {
		fmt.Println("Usage: pinchtab config patch '<json>'")
		fmt.Println()
		fmt.Println("Examples:")
		fmt.Println(`  pinchtab config patch '{"server": {"port": "8080"}}'`)
		fmt.Println(`  pinchtab config patch '{"chrome": {"headless": false, "maxTabs": 50}}'`)
		os.Exit(1)
	}

	jsonPatch := os.Args[3]

	fc, configPath, err := LoadFileConfig()
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		os.Exit(1)
	}

	if err := PatchConfigJSON(fc, jsonPatch); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	// Validate before saving
	if errs := ValidateFileConfig(fc); len(errs) > 0 {
		fmt.Printf("Warning: patch causes validation error(s):\n")
		for _, e := range errs {
			fmt.Printf("  - %v\n", e)
		}
		fmt.Print("Save anyway? (y/N): ")
		var response string
		_, _ = fmt.Scanln(&response)
		if response != "y" && response != "Y" {
			fmt.Println("Aborted.")
			return
		}
	}

	if err := SaveFileConfig(fc, configPath); err != nil {
		fmt.Printf("Error saving config: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("✓ Config patched successfully")
}

func handleConfigValidate() {
	configPath := envOrMigrate("PINCHTAB_CONFIG", "BRIDGE_CONFIG", filepath.Join(userConfigDir(), "config.json"))

	data, err := os.ReadFile(configPath)
	if err != nil {
		fmt.Printf("Error reading config file: %v\n", err)
		os.Exit(1)
	}

	var fc *FileConfig

	if isLegacyConfig(data) {
		var lc legacyFileConfig
		if err := json.Unmarshal(data, &lc); err != nil {
			fmt.Printf("Error parsing config file: %v\n", err)
			os.Exit(1)
		}
		fc = convertLegacyConfig(&lc)
		fmt.Println("Note: Config file uses legacy flat format. Consider migrating to nested format.")
		fmt.Println()
	} else {
		fc = &FileConfig{}
		if err := json.Unmarshal(data, fc); err != nil {
			fmt.Printf("Error parsing config file: %v\n", err)
			os.Exit(1)
		}
	}

	errs := ValidateFileConfig(fc)
	if len(errs) == 0 {
		fmt.Printf("✓ Config file is valid: %s\n", configPath)
		return
	}

	fmt.Printf("✗ Config file has %d error(s):\n", len(errs))
	for _, e := range errs {
		fmt.Printf("  - %v\n", e)
	}
	os.Exit(1)
}

func valueOrNone(s string) string {
	if s == "" {
		return "(none)"
	}
	return s
}

// MaskToken masks a token for display (shows first/last 4 chars).
func MaskToken(t string) string {
	if t == "" {
		return "(none)"
	}
	if len(t) <= 8 {
		return "***"
	}
	return t[:4] + "..." + t[len(t)-4:]
}
