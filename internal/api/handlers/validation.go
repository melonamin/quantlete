package handlers

import (
	"fmt"
	"regexp"
	"strings"
)

// Validation constants for API inputs.
const (
	MaxGearNameLength    = 100
	MaxHashtagLength     = 50
	MaxDescriptionLength = 500
	DefaultPageSize      = 100
	MaxPageSize          = 1000
)

// hashtagRe matches valid hashtag format: alphanumeric with optional underscores/hyphens.
var hashtagRe = regexp.MustCompile(`^#?[A-Za-z0-9][A-Za-z0-9_-]{0,48}$`)

// ValidateGearName validates a gear name.
func ValidateGearName(name string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return fmt.Errorf("name is required")
	}
	if len(name) > MaxGearNameLength {
		return fmt.Errorf("name must be %d characters or less", MaxGearNameLength)
	}
	return nil
}

// ValidateHashtag validates a hashtag format.
func ValidateHashtag(hashtag string) error {
	hashtag = strings.TrimSpace(hashtag)
	if hashtag == "" {
		return fmt.Errorf("hashtag is required")
	}
	if len(hashtag) > MaxHashtagLength {
		return fmt.Errorf("hashtag must be %d characters or less", MaxHashtagLength)
	}
	if !hashtagRe.MatchString(hashtag) {
		return fmt.Errorf("hashtag must be alphanumeric (may include underscores and hyphens)")
	}
	return nil
}

// ValidatePurchaseCurrency validates currency code format.
func ValidatePurchaseCurrency(currency string) error {
	currency = strings.TrimSpace(currency)
	if currency == "" {
		return nil // optional field
	}
	if len(currency) != 3 {
		return fmt.Errorf("currency must be a 3-letter ISO code (e.g., USD, EUR)")
	}
	return nil
}

// ValidateCustomGearRequest validates a custom gear create/update request.
func ValidateCustomGearRequest(name, hashtag, currency string, isCreate bool) error {
	if isCreate || name != "" {
		if err := ValidateGearName(name); err != nil {
			return err
		}
	}
	if isCreate || hashtag != "" {
		if err := ValidateHashtag(hashtag); err != nil {
			return err
		}
	}
	if err := ValidatePurchaseCurrency(currency); err != nil {
		return err
	}
	return nil
}
