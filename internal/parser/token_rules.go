package parser

import (
	"strconv"
	"strings"

	"github.com/nschaetti/cashwarrior/internal/domain"
)

func classifyFlag(raw string) (Token, bool) {
	if raw == "-h" {
		return Token{Raw: raw, Key: "help", Kind: TokenFlag}, true
	}
	if !strings.HasPrefix(raw, "--") || len(raw) <= 2 {
		return Token{}, false
	}
	trimmed := strings.TrimPrefix(raw, "--")
	parts := strings.SplitN(trimmed, "=", 2)
	if len(parts) == 1 {
		return Token{Raw: raw, Key: parts[0], Kind: TokenFlag}, true
	}
	return Token{Raw: raw, Key: parts[0], Value: parts[1], Kind: TokenFlag}, true
}

// classifyNegativeTag classifies tokens like "-@groceries".
func classifyNegativeTag(raw string) (Token, bool) {
	if len(raw) > 2 && strings.HasPrefix(raw, "-@") {
		return Token{
			Raw:  raw,
			Kind: TokenTagNegative,
		}, true
	}
	return Token{}, false
}

// classifyTag classifies tokens like "@groceries".
func classifyTag(raw string) (Token, bool) {
	if len(raw) > 1 && strings.HasPrefix(raw, "@") {
		return Token{
			Raw:  raw,
			Kind: TokenTag,
		}, true
	}
	return Token{}, false
}

// classifyAmount classifies signed numeric tokens like "+12.50" or "-3".
func classifyAmount(raw string) (Token, bool) {
	if strings.HasPrefix(raw, "+") || strings.HasPrefix(raw, "-") {
		// Try to parse as a float.
		amount, err := strconv.ParseFloat(raw, 32)
		if err == nil {
			return Token{
				Raw:    raw,
				Amount: float32(amount),
				Kind:   TokenAmount,
			}, true
		}
	}
	return Token{}, false
}

// classifyAttribute classifies tokens like "key:value" and "key:".
func classifyAttribute(raw string) (Token, bool) {
	if strings.Contains(raw, ":") {
		if strings.HasSuffix(raw, ":") {
			return Token{
				Raw:  raw,
				Key:  strings.SplitN(raw, ":", 2)[0],
				Kind: TokenAttributeClear,
			}, true
		}
		return Token{
			Raw:   raw,
			Key:   strings.SplitN(raw, ":", 2)[0],
			Value: strings.SplitN(raw, ":", 2)[1],
			Kind:  TokenAttribute,
		}, true
	}
	return Token{}, false
}

// classifyID classifies transaction IDs in the public format.
func classifyID(raw string) (Token, bool) {
	transactionId, err := domain.ParseTransactionID(raw)
	if err != nil {
		return Token{}, false
	}
	return Token{
		Raw:     raw,
		TransID: transactionId,
		Kind:    TokenID,
	}, true
}

// classifyPeriod classifies known period keywords (for example "today").
func classifyPeriod(raw string) (Token, bool) {
	period, err := domain.ParsePeriod(raw)
	if err != nil {
		return Token{}, false
	}
	return Token{
		Raw:    raw,
		Kind:   TokenPeriod,
		Period: period,
	}, true
}

// classifyText is the fallback classifier for plain text tokens.
func classifyText(raw string) (Token, bool) {
	return Token{
		Raw:  raw,
		Kind: TokenText,
	}, true
}
