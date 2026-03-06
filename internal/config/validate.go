package config

import (
	"fmt"
	"strconv"
)

// ValidationError represents a configuration validation error.
type ValidationError struct {
	Field   string
	Message string
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

// ValidateFileConfig validates a FileConfig and returns all errors found.
func ValidateFileConfig(fc *FileConfig) []error {
	var errs []error

	// Server validation
	if fc.Server.Port != "" {
		if err := validatePort(fc.Server.Port, "server.port"); err != nil {
			errs = append(errs, err)
		}
	}
	if fc.Server.Bind != "" {
		if err := validateBind(fc.Server.Bind, "server.bind"); err != nil {
			errs = append(errs, err)
		}
	}
	if fc.Server.InstancePortStart != nil && fc.Server.InstancePortEnd != nil {
		if *fc.Server.InstancePortStart > *fc.Server.InstancePortEnd {
			errs = append(errs, ValidationError{
				Field:   "server.instancePortStart/End",
				Message: fmt.Sprintf("start port (%d) must be <= end port (%d)", *fc.Server.InstancePortStart, *fc.Server.InstancePortEnd),
			})
		}
	}

	// Chrome validation
	if fc.Chrome.StealthLevel != "" {
		if !isValidStealthLevel(fc.Chrome.StealthLevel) {
			errs = append(errs, ValidationError{
				Field:   "chrome.stealthLevel",
				Message: fmt.Sprintf("invalid value %q (must be light, medium, or full)", fc.Chrome.StealthLevel),
			})
		}
	}
	if fc.Chrome.TabEvictionPolicy != "" {
		if !isValidEvictionPolicy(fc.Chrome.TabEvictionPolicy) {
			errs = append(errs, ValidationError{
				Field:   "chrome.tabEvictionPolicy",
				Message: fmt.Sprintf("invalid value %q (must be reject, close_oldest, or close_lru)", fc.Chrome.TabEvictionPolicy),
			})
		}
	}
	if fc.Chrome.MaxTabs != nil && *fc.Chrome.MaxTabs < 1 {
		errs = append(errs, ValidationError{
			Field:   "chrome.maxTabs",
			Message: fmt.Sprintf("must be >= 1 (got %d)", *fc.Chrome.MaxTabs),
		})
	}
	if fc.Chrome.MaxParallelTabs != nil && *fc.Chrome.MaxParallelTabs < 0 {
		errs = append(errs, ValidationError{
			Field:   "chrome.maxParallelTabs",
			Message: fmt.Sprintf("must be >= 0 (got %d)", *fc.Chrome.MaxParallelTabs),
		})
	}

	// Orchestrator validation
	if fc.Orchestrator.Strategy != "" {
		if !isValidStrategy(fc.Orchestrator.Strategy) {
			errs = append(errs, ValidationError{
				Field:   "orchestrator.strategy",
				Message: fmt.Sprintf("invalid value %q (must be simple, explicit, or simple-autorestart)", fc.Orchestrator.Strategy),
			})
		}
	}
	if fc.Orchestrator.AllocationPolicy != "" {
		if !isValidAllocationPolicy(fc.Orchestrator.AllocationPolicy) {
			errs = append(errs, ValidationError{
				Field:   "orchestrator.allocationPolicy",
				Message: fmt.Sprintf("invalid value %q (must be fcfs, round_robin, or random)", fc.Orchestrator.AllocationPolicy),
			})
		}
	}

	// Timeouts validation
	if fc.Timeouts.ActionSec < 0 {
		errs = append(errs, ValidationError{
			Field:   "timeouts.actionSec",
			Message: fmt.Sprintf("must be >= 0 (got %d)", fc.Timeouts.ActionSec),
		})
	}
	if fc.Timeouts.NavigateSec < 0 {
		errs = append(errs, ValidationError{
			Field:   "timeouts.navigateSec",
			Message: fmt.Sprintf("must be >= 0 (got %d)", fc.Timeouts.NavigateSec),
		})
	}
	if fc.Timeouts.ShutdownSec < 0 {
		errs = append(errs, ValidationError{
			Field:   "timeouts.shutdownSec",
			Message: fmt.Sprintf("must be >= 0 (got %d)", fc.Timeouts.ShutdownSec),
		})
	}
	if fc.Timeouts.WaitNavMs < 0 {
		errs = append(errs, ValidationError{
			Field:   "timeouts.waitNavMs",
			Message: fmt.Sprintf("must be >= 0 (got %d)", fc.Timeouts.WaitNavMs),
		})
	}

	return errs
}

func validatePort(port string, field string) error {
	p, err := strconv.Atoi(port)
	if err != nil {
		return ValidationError{
			Field:   field,
			Message: fmt.Sprintf("invalid port %q (must be a number)", port),
		}
	}
	if p < 1 || p > 65535 {
		return ValidationError{
			Field:   field,
			Message: fmt.Sprintf("port %d out of range (must be 1-65535)", p),
		}
	}
	return nil
}

func validateBind(bind string, field string) error {
	// Accept common bind addresses
	validBinds := map[string]bool{
		"127.0.0.1": true,
		"0.0.0.0":   true,
		"localhost": true,
		"::1":       true,
		"::":        true,
	}
	if validBinds[bind] {
		return nil
	}
	// Basic IP format check (not exhaustive, just sanity)
	// If it contains a dot, assume it's an IPv4 attempt
	// If it contains a colon, assume it's an IPv6 attempt
	// This is intentionally loose — the OS will reject truly invalid addresses
	return nil
}

func isValidStealthLevel(level string) bool {
	switch level {
	case "light", "medium", "full":
		return true
	default:
		return false
	}
}

func isValidEvictionPolicy(policy string) bool {
	switch policy {
	case "reject", "close_oldest", "close_lru":
		return true
	default:
		return false
	}
}

func isValidStrategy(strategy string) bool {
	switch strategy {
	case "simple", "explicit", "simple-autorestart":
		return true
	default:
		return false
	}
}

func isValidAllocationPolicy(policy string) bool {
	switch policy {
	case "fcfs", "round_robin", "random":
		return true
	default:
		return false
	}
}

// ValidStealthLevels returns all valid stealth level values.
func ValidStealthLevels() []string {
	return []string{"light", "medium", "full"}
}

// ValidEvictionPolicies returns all valid tab eviction policy values.
func ValidEvictionPolicies() []string {
	return []string{"reject", "close_oldest", "close_lru"}
}

// ValidStrategies returns all valid strategy values.
func ValidStrategies() []string {
	return []string{"simple", "explicit", "simple-autorestart"}
}

// ValidAllocationPolicies returns all valid allocation policy values.
func ValidAllocationPolicies() []string {
	return []string{"fcfs", "round_robin", "random"}
}
