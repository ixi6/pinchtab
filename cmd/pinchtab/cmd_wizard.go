package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/pinchtab/pinchtab/internal/cli"
	"github.com/pinchtab/pinchtab/internal/config"
)

// runSecurityWizard runs the interactive security setup wizard.
// isNew indicates a fresh install (full wizard) vs upgrade (migration notice).
// Returns true if the user completed setup, false if they cancelled.
func runSecurityWizard(cfg *config.FileConfig, configPath string, isNew bool) bool {
	interactive := isInteractiveTerminal()

	if !interactive {
		return runNonInteractiveSetup(cfg, configPath, isNew)
	}

	if isNew {
		return runFullWizard(cfg, configPath)
	}
	return runUpgradeNotice(cfg, configPath)
}

// runNonInteractiveSetup prints a summary and applies defaults silently.
func runNonInteractiveSetup(cfg *config.FileConfig, configPath string, isNew bool) bool {
	if isNew {
		fmt.Println()
		fmt.Println(cli.StyleStdout(cli.HeadingStyle, "🔒 Security defaults applied"))
		fmt.Printf("   Dashboard:  http://%s:%s\n", orDefault(cfg.Server.Bind, "127.0.0.1"), orDefault(cfg.Server.Port, "9867"))
		if cfg.Server.Token != "" {
			fmt.Printf("   Token:      %s\n", cfg.Server.Token[:12]+"...")
			fmt.Printf("   Dashboard:  http://%s:%s?token=%s\n",
				orDefault(cfg.Server.Bind, "127.0.0.1"),
				orDefault(cfg.Server.Port, "9867"),
				cfg.Server.Token)
		}
		fmt.Println("   IDPI:       enabled (localhost only)")
		fmt.Println()
		fmt.Println("   Run " + cli.StyleStdout(cli.CommandStyle, "pinchtab security") + " to review all settings.")
		fmt.Println()
	} else {
		fmt.Println()
		fmt.Println(cli.StyleStdout(cli.HeadingStyle, "🔒 Config updated to v"+config.CurrentConfigVersion))
		fmt.Println("   Run " + cli.StyleStdout(cli.CommandStyle, "pinchtab security") + " to review changes.")
		fmt.Println()
	}

	cfg.ConfigVersion = config.CurrentConfigVersion
	_ = config.SaveFileConfig(cfg, configPath)
	return true
}

// runFullWizard runs the interactive first-run wizard.
func runFullWizard(cfg *config.FileConfig, configPath string) bool {
	fmt.Println()
	fmt.Println(cli.StyleStdout(cli.HeadingStyle, "🔒 Security Setup"))
	fmt.Println()
	fmt.Println("   We care about your security — PinchTab ships with the strongest defaults.")
	fmt.Println()
	printSeparator()

	// Step 1: IDPI
	fmt.Println()
	fmt.Println(cli.StyleStdout(cli.HeadingStyle, "   1. IDPI (Injection Detection)"))
	fmt.Println()
	domains := cfg.Security.IDPI.AllowedDomains
	if len(domains) == 0 {
		domains = []string{"127.0.0.1", "localhost", "::1"}
	}
	fmt.Printf("   Allowed domains: %s\n", cli.StyleStdout(cli.ValueStyle, strings.Join(domains, ", ")))
	fmt.Println("   Only locally-served sites can be automated without warnings.")
	fmt.Println()

	picked, err := promptSelect("IDPI allowed domains", []menuOption{
		{label: "Confirm defaults", value: "confirm"},
		{label: "Open in browser to edit", value: "browser"},
	})
	if err != nil {
		return false
	}
	if picked == "browser" {
		fmt.Println()
		fmt.Println("   " + cli.StyleStdout(cli.MutedStyle, "Complete setup in your browser, then restart the server."))
		fmt.Printf("   %s\n", dashboardURL(cfg, "/setup"))
		fmt.Println()
		// Save what we have so far
		cfg.ConfigVersion = config.CurrentConfigVersion
		_ = config.SaveFileConfig(cfg, configPath)
		return true
	}

	// Step 2: Dashboard access
	printSeparator()
	fmt.Println()
	fmt.Println(cli.StyleStdout(cli.HeadingStyle, "   2. Dashboard Access"))
	fmt.Println()
	if cfg.Server.Token != "" {
		fmt.Printf("   Token: %s\n", cli.StyleStdout(cli.ValueStyle, cfg.Server.Token))
		fmt.Println()
		url := dashboardURL(cfg, "")
		fmt.Printf("   Dashboard: %s\n", cli.StyleStdout(cli.CommandStyle, url))
	} else {
		fmt.Println("   " + cli.StyleStdout(cli.WarningStyle, "No token set — dashboard is unprotected!"))
	}
	fmt.Println()

	picked, err = promptSelect("Dashboard token", []menuOption{
		{label: "Keep token", value: "keep"},
		{label: "Copy to clipboard", value: "copy"},
		{label: "Generate new token", value: "new"},
	})
	if err != nil {
		return false
	}

	switch picked {
	case "copy":
		copyToClipboard(cfg.Server.Token)
	case "new":
		token, err := config.GenerateAuthToken()
		if err == nil {
			cfg.Server.Token = token
			fmt.Printf("   New token: %s\n", cli.StyleStdout(cli.ValueStyle, token))
		}
	}

	// Step 3: Security summary
	printSeparator()
	fmt.Println()
	fmt.Println(cli.StyleStdout(cli.HeadingStyle, "   3. Security Settings"))
	fmt.Println()
	printSecuritySetting("evaluate", boolPtrValue(cfg.Security.AllowEvaluate))
	printSecuritySetting("download", boolPtrValue(cfg.Security.AllowDownload))
	printSecuritySetting("upload", boolPtrValue(cfg.Security.AllowUpload))
	printSecuritySetting("macros", boolPtrValue(cfg.Security.AllowMacro))
	printSecuritySetting("screencast", boolPtrValue(cfg.Security.AllowScreencast))
	fmt.Println()
	fmt.Println("   Review or change with " + cli.StyleStdout(cli.CommandStyle, "pinchtab security"))
	fmt.Println()

	printSeparator()
	fmt.Println()

	// Save
	cfg.ConfigVersion = config.CurrentConfigVersion
	if err := config.SaveFileConfig(cfg, configPath); err != nil {
		fmt.Fprintln(os.Stderr, cli.StyleStderr(cli.ErrorStyle, fmt.Sprintf("failed to save config: %v", err)))
		return false
	}

	fmt.Println(cli.StyleStdout(cli.SuccessStyle, "   ✓ Setup complete!"))
	fmt.Println()
	return true
}

// runUpgradeNotice shows a brief notice for config upgrades.
func runUpgradeNotice(cfg *config.FileConfig, configPath string) bool {
	fmt.Println()
	fmt.Println(cli.StyleStdout(cli.HeadingStyle, "🔒 Security update (v"+config.CurrentConfigVersion+")"))
	fmt.Println()

	oldVersion := cfg.ConfigVersion
	if oldVersion == "" {
		oldVersion = "pre-0.8.0"
	}
	fmt.Printf("   Config upgraded: %s → %s\n", oldVersion, config.CurrentConfigVersion)

	if cfg.Server.Token != "" {
		fmt.Printf("   Dashboard token: %s\n", cli.StyleStdout(cli.ValueStyle, cfg.Server.Token[:min(12, len(cfg.Server.Token))]+"..."))
	}

	fmt.Println()
	fmt.Println("   Run " + cli.StyleStdout(cli.CommandStyle, "pinchtab security") + " to review all settings.")
	fmt.Println()

	cfg.ConfigVersion = config.CurrentConfigVersion
	_ = config.SaveFileConfig(cfg, configPath)
	return true
}

// ─── Helpers ─────────────────────────────────────────────────────

func printSeparator() {
	fmt.Println("   " + cli.StyleStdout(cli.MutedStyle, strings.Repeat("━", 50)))
}

func printSecuritySetting(name string, enabled bool) {
	status := cli.StyleStdout(cli.SuccessStyle, "disabled")
	if enabled {
		status = cli.StyleStdout(cli.WarningStyle, "enabled")
	}
	fmt.Printf("   • %-12s %s\n", name+":", status)
}

func boolPtrValue(p *bool) bool {
	if p == nil {
		return false
	}
	return *p
}

func dashboardURL(cfg *config.FileConfig, path string) string {
	host := orDefault(cfg.Server.Bind, "127.0.0.1")
	port := orDefault(cfg.Server.Port, "9867")
	url := fmt.Sprintf("http://%s:%s%s", host, port, path)
	if cfg.Server.Token != "" {
		url += "?token=" + cfg.Server.Token
	}
	return url
}

func orDefault(val, fallback string) string {
	if val == "" {
		return fallback
	}
	return val
}

// copyToClipboard is defined in cmd_config.go
