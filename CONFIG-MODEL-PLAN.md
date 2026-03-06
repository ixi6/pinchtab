# Config Model Evolution Plan

**Branch:** `feat/config-model`
**Base:** `origin/main` (cd5f42a)

## Current State

The config system already has solid foundations:
- `RuntimeConfig` struct (flat, typed, ~40 fields)
- `FileConfig` struct (flat, ~15 fields — incomplete)
- `Load()` with proper precedence: env vars > config file > defaults
- Migration helpers for `BRIDGE_*` → `PINCHTAB_*`
- CLI commands: `init`, `show`
- Integration tests: `tests/integration/config_test.go`

## Gap Analysis (vs PR #91 model)

| Feature | Current | Target |
|---------|---------|--------|
| FileConfig structure | Flat JSON | Nested sections (Server/Chrome/Orchestrator/Timeouts) |
| FileConfig fields | 15 fields | All RuntimeConfig fields mirrored |
| CLI commands | init, show | + set, patch, validate |
| Validation | None | ValidateConfig() with enum checks |
| Editor helpers | None | SetConfigValue(), PatchConfigJSON() |
| YAML support | None | JSON + YAML |
| Documentation | Outdated | Full tables, JSON/YAML examples |

## Implementation Plan

### Phase 1: Expand FileConfig (Breaking: None)

**File:** `internal/config/config.go`

1. Add nested `FileConfig` structure:
```go
type FileConfig struct {
    Server      ServerConfig      `json:"server"`
    Chrome      ChromeConfig      `json:"chrome"`
    Orchestrator OrchestratorConfig `json:"orchestrator"`
    Timeouts    TimeoutsConfig    `json:"timeouts"`
}

type ServerConfig struct {
    Port      string `json:"port,omitempty"`
    Bind      string `json:"bind,omitempty"`
    Token     string `json:"token,omitempty"`
    StateDir  string `json:"stateDir,omitempty"`
    // ... all server-related fields
}

type ChromeConfig struct {
    Headless        *bool  `json:"headless,omitempty"`
    MaxTabs         *int   `json:"maxTabs,omitempty"`
    ProfileDir      string `json:"profileDir,omitempty"`
    StealthLevel    string `json:"stealthLevel,omitempty"`
    TabEvictionPolicy string `json:"tabEvictionPolicy,omitempty"`
    // ... all chrome-related fields
}

type OrchestratorConfig struct {
    Strategy         string `json:"strategy,omitempty"`
    AllocationPolicy string `json:"allocationPolicy,omitempty"`
}

type TimeoutsConfig struct {
    ActionSec   int `json:"actionSec,omitempty"`
    NavigateSec int `json:"navigateSec,omitempty"`
    ShutdownSec int `json:"shutdownSec,omitempty"`
}
```

2. Update `Load()` selective merge to handle nested structure
3. Maintain backward compatibility with flat config files (detect & migrate)

**Tests:**
- Add `TestLoadNestedConfig` 
- Add `TestLoadFlatConfigBackwardCompat`
- Update existing tests as needed

### Phase 2: Validation (Breaking: None)

**File:** `internal/config/validate.go` (new)

```go
func ValidateConfig(fc *FileConfig) []error {
    var errs []error
    
    // Enum validation
    if fc.Chrome.StealthLevel != "" {
        if !isValidStealthLevel(fc.Chrome.StealthLevel) {
            errs = append(errs, fmt.Errorf("invalid stealthLevel: %s (must be light|medium|full)", fc.Chrome.StealthLevel))
        }
    }
    
    if fc.Chrome.TabEvictionPolicy != "" {
        if !isValidEvictionPolicy(fc.Chrome.TabEvictionPolicy) {
            errs = append(errs, fmt.Errorf("invalid tabEvictionPolicy: %s (must be reject|close_oldest|close_lru)", fc.Chrome.TabEvictionPolicy))
        }
    }
    
    if fc.Orchestrator.Strategy != "" {
        if !isValidStrategy(fc.Orchestrator.Strategy) {
            errs = append(errs, fmt.Errorf("invalid strategy: %s (must be simple|explicit|simple-autorestart)", fc.Orchestrator.Strategy))
        }
    }
    
    // Range validation
    if fc.Server.Port != "" {
        port, err := strconv.Atoi(fc.Server.Port)
        if err != nil || port < 1 || port > 65535 {
            errs = append(errs, fmt.Errorf("invalid port: %s (must be 1-65535)", fc.Server.Port))
        }
    }
    
    return errs
}
```

**Tests:**
- `TestValidateConfig_ValidValues`
- `TestValidateConfig_InvalidEnum`
- `TestValidateConfig_InvalidPort`

### Phase 3: Editor Helpers (Breaking: None)

**File:** `internal/config/editor.go` (new)

```go
// SetConfigValue sets a dotted path in FileConfig (e.g., "server.port", "chrome.headless")
func SetConfigValue(fc *FileConfig, path string, value string) error

// PatchConfigJSON merges a JSON patch into FileConfig
func PatchConfigJSON(fc *FileConfig, jsonPatch string) error

// SaveConfig writes FileConfig to disk (JSON or YAML based on extension)
func SaveConfig(fc *FileConfig, path string) error
```

**Tests:**
- `TestSetConfigValue_ServerPort`
- `TestSetConfigValue_ChromeHeadless`
- `TestSetConfigValue_InvalidPath`
- `TestPatchConfigJSON`
- `TestSaveConfig_JSON`
- `TestSaveConfig_YAML`

### Phase 4: CLI Commands (Breaking: None)

**File:** `internal/config/config.go` (extend `HandleConfigCommand`)

Add subcommands:
- `pinchtab config set server.port 8080` — set single value
- `pinchtab config patch '{"chrome":{"headless":false}}'` — JSON patch
- `pinchtab config validate` — check config file for errors
- `pinchtab config show --format yaml` — output in YAML

**Tests:**
- Integration tests for each command
- Unit tests for parsing/formatting

### Phase 5: YAML Support (Breaking: None)

**File:** `internal/config/config.go`

1. Detect file extension (.json vs .yaml/.yml)
2. Use `gopkg.in/yaml.v3` for YAML (or keep it pure with custom marshaler)
3. Update `SaveConfig()` to support both formats

**Decision:** Use `gopkg.in/yaml.v3` — standard, well-tested, minimal overhead.

### Phase 6: Documentation Update

**File:** `docs/references/configuration.md`

1. Add nested JSON example
2. Add YAML example
3. Update env var tables with all new fields
4. Add "Config File vs Env Vars" section
5. Document CLI commands (set, patch, validate)

## Test Impact Analysis

| Test File | Changes Needed |
|-----------|----------------|
| `internal/config/config_test.go` | Add nested config tests, backward compat tests |
| `tests/integration/config_test.go` | Add CLI command tests (set, patch, validate) |
| New: `internal/config/validate_test.go` | Validation unit tests |
| New: `internal/config/editor_test.go` | Editor helper unit tests |

## Backward Compatibility

1. **Flat config files** — `Load()` will detect flat vs nested and handle both
2. **Env vars** — Continue to take precedence over file config
3. **`BRIDGE_*` vars** — Migration warnings still work
4. **Existing CLI** — `init` and `show` unchanged

## Migration Path

Users with existing flat config files:
1. `pinchtab config show` — shows current effective config
2. `pinchtab config init --migrate` — converts flat to nested (optional, not required)
3. Flat configs continue to work indefinitely

## Implementation Order

1. ✅ Phase 1 (Nested FileConfig) — foundation for everything else
2. ✅ Phase 2 (Validation) — catches errors early
3. ✅ Phase 3 (Editor Helpers) — SetConfigValue, PatchConfigJSON, Load/SaveFileConfig
4. ✅ Phase 4 (CLI Commands) — set, patch commands with validation
5. Phase 5 (YAML) — nice-to-have, can defer
6. ✅ Phase 6 (Docs) — updated configuration.md

## Estimated LOC

| Phase | New Lines | Modified Lines |
|-------|-----------|----------------|
| 1. Nested FileConfig | ~150 | ~100 |
| 2. Validation | ~100 | ~20 |
| 3. Editor Helpers | ~150 | ~10 |
| 4. CLI Commands | ~100 | ~50 |
| 5. YAML Support | ~50 | ~30 |
| 6. Documentation | ~200 | ~100 |
| **Total** | **~750** | **~310** |

## Open Questions

1. **YAML dependency** — Add `gopkg.in/yaml.v3` or keep pure stdlib?
   - Recommendation: Add it — it's the standard and battle-tested

2. **Flat config auto-migration** — Should `pinchtab config init` auto-migrate existing flat configs?
   - Recommendation: Yes, with confirmation prompt

3. **Config file location** — Keep `~/.pinchtab/config.json` or move to XDG-compliant `~/.config/pinchtab/config.json`?
   - Recommendation: Already handled by `userConfigDir()` — defaults to XDG with fallback to legacy
