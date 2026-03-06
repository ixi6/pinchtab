package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// SetConfigValue sets a dotted path in FileConfig (e.g., "server.port", "chrome.headless").
// Returns the updated FileConfig and any error.
func SetConfigValue(fc *FileConfig, path string, value string) error {
	parts := strings.SplitN(path, ".", 2)
	if len(parts) != 2 {
		return fmt.Errorf("invalid path %q (expected section.field, e.g., server.port)", path)
	}

	section, field := parts[0], parts[1]

	switch section {
	case "server":
		return setServerField(&fc.Server, field, value)
	case "chrome":
		return setChromeField(&fc.Chrome, field, value)
	case "security":
		return setSecurityField(&fc.Security, field, value)
	case "orchestrator":
		return setOrchestratorField(&fc.Orchestrator, field, value)
	case "timeouts":
		return setTimeoutsField(&fc.Timeouts, field, value)
	default:
		return fmt.Errorf("unknown section %q (valid: server, chrome, security, orchestrator, timeouts)", section)
	}
}

func setServerField(s *ServerConfig, field, value string) error {
	switch field {
	case "port":
		s.Port = value
	case "bind":
		s.Bind = value
	case "token":
		s.Token = value
	case "stateDir":
		s.StateDir = value
	case "cdpUrl":
		s.CdpURL = value
	case "instancePortStart":
		n, err := strconv.Atoi(value)
		if err != nil {
			return fmt.Errorf("server.instancePortStart must be a number: %w", err)
		}
		s.InstancePortStart = &n
	case "instancePortEnd":
		n, err := strconv.Atoi(value)
		if err != nil {
			return fmt.Errorf("server.instancePortEnd must be a number: %w", err)
		}
		s.InstancePortEnd = &n
	default:
		return fmt.Errorf("unknown field server.%s", field)
	}
	return nil
}

func setChromeField(c *ChromeConfig, field, value string) error {
	switch field {
	case "headless":
		b, err := parseBool(value)
		if err != nil {
			return fmt.Errorf("chrome.headless: %w", err)
		}
		c.Headless = &b
	case "noRestore":
		b, err := parseBool(value)
		if err != nil {
			return fmt.Errorf("chrome.noRestore: %w", err)
		}
		c.NoRestore = &b
	case "profileDir":
		c.ProfileDir = value
	case "chromeVersion":
		c.ChromeVersion = value
	case "timezone":
		c.Timezone = value
	case "blockImages":
		b, err := parseBool(value)
		if err != nil {
			return fmt.Errorf("chrome.blockImages: %w", err)
		}
		c.BlockImages = &b
	case "blockMedia":
		b, err := parseBool(value)
		if err != nil {
			return fmt.Errorf("chrome.blockMedia: %w", err)
		}
		c.BlockMedia = &b
	case "blockAds":
		b, err := parseBool(value)
		if err != nil {
			return fmt.Errorf("chrome.blockAds: %w", err)
		}
		c.BlockAds = &b
	case "maxTabs":
		n, err := strconv.Atoi(value)
		if err != nil {
			return fmt.Errorf("chrome.maxTabs must be a number: %w", err)
		}
		c.MaxTabs = &n
	case "maxParallelTabs":
		n, err := strconv.Atoi(value)
		if err != nil {
			return fmt.Errorf("chrome.maxParallelTabs must be a number: %w", err)
		}
		c.MaxParallelTabs = &n
	case "chromeBinary":
		c.ChromeBinary = value
	case "chromeExtraFlags":
		c.ChromeExtraFlags = value
	case "userAgent":
		c.UserAgent = value
	case "noAnimations":
		b, err := parseBool(value)
		if err != nil {
			return fmt.Errorf("chrome.noAnimations: %w", err)
		}
		c.NoAnimations = &b
	case "stealthLevel":
		c.StealthLevel = value
	case "tabEvictionPolicy":
		c.TabEvictionPolicy = value
	default:
		return fmt.Errorf("unknown field chrome.%s", field)
	}
	return nil
}

func setSecurityField(s *SecurityConfig, field, value string) error {
	b, err := parseBool(value)
	if err != nil {
		return fmt.Errorf("security.%s: %w", field, err)
	}

	switch field {
	case "allowEvaluate":
		s.AllowEvaluate = &b
	case "allowMacro":
		s.AllowMacro = &b
	case "allowScreencast":
		s.AllowScreencast = &b
	case "allowDownload":
		s.AllowDownload = &b
	case "allowUpload":
		s.AllowUpload = &b
	default:
		return fmt.Errorf("unknown field security.%s", field)
	}
	return nil
}

func setOrchestratorField(o *OrchestratorConfig, field, value string) error {
	switch field {
	case "strategy":
		o.Strategy = value
	case "allocationPolicy":
		o.AllocationPolicy = value
	default:
		return fmt.Errorf("unknown field orchestrator.%s", field)
	}
	return nil
}

func setTimeoutsField(t *TimeoutsConfig, field, value string) error {
	n, err := strconv.Atoi(value)
	if err != nil {
		return fmt.Errorf("timeouts.%s must be a number: %w", field, err)
	}

	switch field {
	case "actionSec":
		t.ActionSec = n
	case "navigateSec":
		t.NavigateSec = n
	case "shutdownSec":
		t.ShutdownSec = n
	case "waitNavMs":
		t.WaitNavMs = n
	default:
		return fmt.Errorf("unknown field timeouts.%s", field)
	}
	return nil
}

func parseBool(s string) (bool, error) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "true", "1", "yes", "on":
		return true, nil
	case "false", "0", "no", "off":
		return false, nil
	default:
		return false, fmt.Errorf("invalid boolean %q (use true/false)", s)
	}
}

// PatchConfigJSON merges a JSON object into FileConfig.
// The patch should be a valid JSON object with the same structure as FileConfig.
func PatchConfigJSON(fc *FileConfig, jsonPatch string) error {
	// First, serialize current config to JSON
	current, err := json.Marshal(fc)
	if err != nil {
		return fmt.Errorf("failed to serialize current config: %w", err)
	}

	// Unmarshal current into a map for merging
	var currentMap map[string]interface{}
	if err := json.Unmarshal(current, &currentMap); err != nil {
		return fmt.Errorf("failed to parse current config: %w", err)
	}

	// Parse the patch
	var patchMap map[string]interface{}
	if err := json.Unmarshal([]byte(jsonPatch), &patchMap); err != nil {
		return fmt.Errorf("invalid JSON patch: %w", err)
	}

	// Deep merge patch into current
	merged := deepMerge(currentMap, patchMap)

	// Serialize merged back to JSON
	mergedJSON, err := json.Marshal(merged)
	if err != nil {
		return fmt.Errorf("failed to serialize merged config: %w", err)
	}

	// Unmarshal back into FileConfig
	if err := json.Unmarshal(mergedJSON, fc); err != nil {
		return fmt.Errorf("failed to parse merged config: %w", err)
	}

	return nil
}

// deepMerge recursively merges src into dst, returning the result.
func deepMerge(dst, src map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})

	// Copy dst
	for k, v := range dst {
		result[k] = v
	}

	// Merge src
	for k, v := range src {
		if srcMap, ok := v.(map[string]interface{}); ok {
			if dstMap, ok := result[k].(map[string]interface{}); ok {
				result[k] = deepMerge(dstMap, srcMap)
				continue
			}
		}
		result[k] = v
	}

	return result
}

// LoadFileConfig loads a FileConfig from the default or specified path.
// Returns the config and the path it was loaded from.
func LoadFileConfig() (*FileConfig, string, error) {
	configPath := envOrMigrate("PINCHTAB_CONFIG", "BRIDGE_CONFIG", filepath.Join(userConfigDir(), "config.json"))

	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			// Return empty config if file doesn't exist
			return &FileConfig{}, configPath, nil
		}
		return nil, configPath, fmt.Errorf("failed to read config file: %w", err)
	}

	var fc *FileConfig

	if isLegacyConfig(data) {
		var lc legacyFileConfig
		if err := json.Unmarshal(data, &lc); err != nil {
			return nil, configPath, fmt.Errorf("failed to parse legacy config: %w", err)
		}
		fc = convertLegacyConfig(&lc)
	} else {
		fc = &FileConfig{}
		if err := json.Unmarshal(data, fc); err != nil {
			return nil, configPath, fmt.Errorf("failed to parse config: %w", err)
		}
	}

	return fc, configPath, nil
}

// SaveFileConfig saves a FileConfig to the specified path.
func SaveFileConfig(fc *FileConfig, path string) error {
	// Create directory if needed
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := json.MarshalIndent(fc, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to serialize config: %w", err)
	}

	// Add trailing newline
	data = append(data, '\n')

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}
