# Configuration

Complete reference for PinchTab configuration. Supports environment variables, config files (JSON), and CLI commands.

## Configuration Priority

Values are loaded in this order (highest priority first):

1. **Environment variables** — always win
2. **Config file** — `~/.config/pinchtab/config.json` (or `~/.pinchtab/config.json` legacy)
3. **Built-in defaults** — used if nothing else is set

## Config File

### Location

Default location varies by OS:
- **macOS:** `~/Library/Application Support/pinchtab/config.json`
- **Linux:** `~/.config/pinchtab/config.json` (or `$XDG_CONFIG_HOME/pinchtab/config.json`)
- **Windows:** `%APPDATA%\pinchtab\config.json`

For backward compatibility, `~/.pinchtab/config.json` is used if it exists and the new location doesn't.

Override with: `PINCHTAB_CONFIG=/path/to/config.json`

### Format (Nested JSON)

```json
{
  "server": {
    "port": "9867",
    "bind": "127.0.0.1",
    "token": "your-secret-token",
    "stateDir": "/path/to/state",
    "instancePortStart": 9868,
    "instancePortEnd": 9968
  },
  "chrome": {
    "headless": true,
    "maxTabs": 20,
    "profileDir": "/path/to/chrome-profile",
    "stealthLevel": "light",
    "tabEvictionPolicy": "reject",
    "blockAds": false,
    "blockImages": false,
    "blockMedia": false,
    "noRestore": false,
    "noAnimations": false
  },
  "security": {
    "allowEvaluate": false,
    "allowMacro": false,
    "allowScreencast": false,
    "allowDownload": false,
    "allowUpload": false
  },
  "orchestrator": {
    "strategy": "simple",
    "allocationPolicy": "fcfs"
  },
  "timeouts": {
    "actionSec": 30,
    "navigateSec": 60,
    "shutdownSec": 10,
    "waitNavMs": 1000
  }
}
```

### Legacy Flat Format (Deprecated)

Older flat format is still supported for backward compatibility:

```json
{
  "port": "9867",
  "headless": true,
  "maxTabs": 20,
  "allowEvaluate": false,
  "timeoutSec": 30,
  "navigateSec": 60
}
```

Run `pinchtab config init` to generate a new config with the nested format.

## Environment Variables

Environment variables always take precedence over config file values.

### Server

| Variable | Default | Description |
|----------|---------|-------------|
| `PINCHTAB_PORT` | `9867` | HTTP server port |
| `PINCHTAB_BIND` | `127.0.0.1` | Bind address |
| `PINCHTAB_TOKEN` | (none) | API authentication token |
| `PINCHTAB_STATE_DIR` | (OS config dir) | State/data directory |
| `PINCHTAB_CONFIG` | (OS config dir)/config.json | Config file path |
| `PINCHTAB_INSTANCE_PORT_START` | `9868` | First port for browser instances |
| `PINCHTAB_INSTANCE_PORT_END` | `9968` | Last port for browser instances |
| `CDP_URL` | (none) | External Chrome DevTools Protocol URL |

### Security

| Variable | Default | Description |
|----------|---------|-------------|
| `PINCHTAB_ALLOW_EVALUATE` | `false` | Enable `/evaluate` endpoint |
| `PINCHTAB_ALLOW_MACRO` | `false` | Enable macro recording/playback |
| `PINCHTAB_ALLOW_SCREENCAST` | `false` | Enable screencast streaming |
| `PINCHTAB_ALLOW_DOWNLOAD` | `false` | Enable file downloads |
| `PINCHTAB_ALLOW_UPLOAD` | `false` | Enable file uploads |

### Chrome/Browser

| Variable | Default | Description |
|----------|---------|-------------|
| `PINCHTAB_HEADLESS` | `true` | Run Chrome headless |
| `PINCHTAB_PROFILE_DIR` | (state dir)/chrome-profile | Chrome profile directory |
| `PINCHTAB_MAX_TABS` | `20` | Maximum tabs per instance |
| `PINCHTAB_MAX_PARALLEL_TABS` | `0` (auto) | Max parallel tab operations |
| `PINCHTAB_STEALTH` | `light` | Stealth level: `light`, `medium`, `full` |
| `PINCHTAB_TAB_EVICTION_POLICY` | `reject` | Tab limit behavior: `reject`, `close_oldest`, `close_lru` |
| `PINCHTAB_NO_RESTORE` | `false` | Don't restore previous session |
| `PINCHTAB_NO_ANIMATIONS` | `false` | Disable CSS animations |
| `PINCHTAB_BLOCK_ADS` | `false` | Block ad domains |
| `PINCHTAB_BLOCK_IMAGES` | `false` | Block image loading |
| `PINCHTAB_BLOCK_MEDIA` | `false` | Block video/audio |
| `PINCHTAB_CHROME_VERSION` | `144.0.7559.133` | Chrome version for UA/fingerprint |
| `PINCHTAB_USER_AGENT` | (auto) | Custom user agent |
| `PINCHTAB_TIMEZONE` | (system) | Browser timezone |
| `CHROME_BIN` | (auto) | Chrome binary path |
| `CHROME_FLAGS` | (none) | Extra Chrome flags |
| `CHROME_EXTENSION_PATHS` | (none) | Comma-separated extension paths |

### Orchestrator (Dashboard Mode)

| Variable | Default | Description |
|----------|---------|-------------|
| `PINCHTAB_STRATEGY` | `simple` | Strategy: `simple`, `explicit`, `simple-autorestart` |
| `PINCHTAB_ALLOCATION_POLICY` | `fcfs` | Instance allocation: `fcfs`, `round_robin`, `random` |

### Legacy Variables (Deprecated)

The following `BRIDGE_*` variables still work but emit warnings. Use the `PINCHTAB_*` equivalents:

| Legacy | New |
|--------|-----|
| `BRIDGE_PORT` | `PINCHTAB_PORT` |
| `BRIDGE_BIND` | `PINCHTAB_BIND` |
| `BRIDGE_TOKEN` | `PINCHTAB_TOKEN` |
| `BRIDGE_HEADLESS` | `PINCHTAB_HEADLESS` |
| `BRIDGE_PROFILE` | `PINCHTAB_PROFILE_DIR` |
| `BRIDGE_MAX_TABS` | `PINCHTAB_MAX_TABS` |
| `BRIDGE_STEALTH` | `PINCHTAB_STEALTH` |
| `BRIDGE_ALLOW_EVALUATE` | `PINCHTAB_ALLOW_EVALUATE` |

## CLI Commands

### pinchtab config init

Create a default config file:

```bash
pinchtab config init
```

Creates `config.json` in the default location with sensible defaults.

### pinchtab config show

Show current effective configuration:

```bash
pinchtab config show
```

Shows all settings with their current values (from env vars, config file, or defaults).

### pinchtab config path

Show config file path:

```bash
pinchtab config path
```

### pinchtab config validate

Validate config file:

```bash
pinchtab config validate
```

Checks for:
- Valid port numbers (1-65535)
- Valid enum values (strategy, stealthLevel, tabEvictionPolicy, etc.)
- Valid timeout values (non-negative)
- Instance port range (start <= end)

## Examples

### Basic Setup (Defaults)

```bash
pinchtab
```

Runs on `localhost:9867`, headless, no authentication.

### With Authentication

```bash
PINCHTAB_TOKEN=my-secret-token pinchtab
```

Or in config file:
```json
{
  "server": {
    "token": "my-secret-token"
  }
}
```

### Network Accessible

```bash
PINCHTAB_BIND=0.0.0.0 PINCHTAB_TOKEN=secret pinchtab
```

**⚠️ Always use a token when binding to 0.0.0.0**

### Headed Mode for Debugging

```bash
PINCHTAB_HEADLESS=false pinchtab
```

Or in config file:
```json
{
  "chrome": {
    "headless": false
  }
}
```

### Maximum Stealth

```bash
PINCHTAB_STEALTH=full pinchtab
```

Higher stealth = more bot detection bypass, but slower.

### Enable Dangerous Endpoints

```bash
PINCHTAB_ALLOW_EVALUATE=true \
PINCHTAB_ALLOW_MACRO=true \
PINCHTAB_TOKEN=secret \
pinchtab
```

Or in config file:
```json
{
  "server": {
    "token": "secret"
  },
  "security": {
    "allowEvaluate": true,
    "allowMacro": true
  }
}
```

### Custom Ports

```bash
PINCHTAB_PORT=8080 \
PINCHTAB_INSTANCE_PORT_START=8100 \
PINCHTAB_INSTANCE_PORT_END=8200 \
pinchtab dashboard
```

### Tab Eviction Policy

When max tabs is reached:
- `reject` — Return error (default, safest)
- `close_oldest` — Close oldest tab by creation time
- `close_lru` — Close least recently used tab

```json
{
  "chrome": {
    "maxTabs": 10,
    "tabEvictionPolicy": "close_lru"
  }
}
```

## Validation

All enum fields are validated on load:

| Field | Valid Values |
|-------|--------------|
| `chrome.stealthLevel` | `light`, `medium`, `full` |
| `chrome.tabEvictionPolicy` | `reject`, `close_oldest`, `close_lru` |
| `orchestrator.strategy` | `simple`, `explicit`, `simple-autorestart` |
| `orchestrator.allocationPolicy` | `fcfs`, `round_robin`, `random` |

Run `pinchtab config validate` to check your config file.

## Related Documentation

- [API Reference](endpoints.md) — HTTP endpoints
- [CLI Reference](cli-quick-reference.md) — Command line usage
- [Instance API](instance-api.md) — Multi-instance management
