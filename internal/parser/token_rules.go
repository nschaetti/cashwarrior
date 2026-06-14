package parser

import (
	"strings"
)

// classifyFlag classifies tokens like "-h" or "--key=value" as TokenFlag
// -h, --... -> TokenFlag
func classifyFlag(raw string) bool {
	// -h help
	if raw == "-h" {
		return true
	}

	// alone --, not a flag
	if !strings.HasPrefix(raw, "--") || len(raw) <= 2 {
		return false
	}

	trimmed := strings.TrimPrefix(raw, "--")

	// --key=value
	parts := strings.SplitN(trimmed, "=", 2)
	// key
	if len(parts) == 1 {
		return true
	}

	// key=value
	return true
}

// classifyNegativeTag classifies tokens like "-@groceries".
func classifyNegativeTag(raw string) bool {
	if len(raw) > 2 && strings.HasPrefix(raw, "-@") {
		return true
	}
	return false
}

// classifyTag classifies tokens like "@groceries".
func classifyTag(raw string) bool {
	if len(raw) > 1 && strings.HasPrefix(raw, "@") {
		return true
	}
	return false
}

// classifyAttribute classifies tokens like "key:value" and "key:".
func classifyAttribute(raw string) bool {
	if strings.Contains(raw, ":") {
		if strings.HasSuffix(raw, ":") {
			return true
		}
		return true
	}
	return false
}

// classifyText is the fallback classifier for plain text tokens.
func classifyText(raw string) bool {
	return true
}
