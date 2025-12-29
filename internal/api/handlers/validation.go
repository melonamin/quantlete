package handlers

import (
	"fmt"
	"net"
	"net/url"
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

// allowedImportHosts restricts which hosts can be fetched for challenge import.
// This prevents SSRF attacks by only allowing trusted external hosts.
var allowedImportHosts = map[string]bool{
	"www.strava.com": true,
	"strava.com":     true,
}

// ValidateImportURL validates a URL for challenge import to prevent SSRF attacks.
// Only HTTPS URLs to trusted hosts (Strava) are allowed.
func ValidateImportURL(rawURL string) error {
	rawURL = strings.TrimSpace(rawURL)
	if rawURL == "" {
		return fmt.Errorf("URL is required")
	}

	parsed, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Errorf("invalid URL format")
	}

	// Only allow HTTPS
	if parsed.Scheme != "https" {
		return fmt.Errorf("only HTTPS URLs are allowed")
	}

	// Normalize hostname
	host := strings.ToLower(parsed.Hostname())
	if host == "" {
		return fmt.Errorf("URL must include a hostname")
	}

	// Check against allowlist
	if !allowedImportHosts[host] {
		return fmt.Errorf("URL must be from strava.com")
	}

	// Validate the host doesn't resolve to internal IPs (defense in depth)
	if err := validateExternalHost(host); err != nil {
		return err
	}

	return nil
}

// validateExternalHost ensures a hostname doesn't resolve to internal/private IPs.
func validateExternalHost(host string) error {
	ips, err := net.LookupIP(host)
	if err != nil {
		return fmt.Errorf("failed to resolve hostname")
	}

	for _, ip := range ips {
		if isPrivateIP(ip) {
			return fmt.Errorf("URL resolves to internal address")
		}
	}

	return nil
}

// isPrivateIP checks if an IP address is private, loopback, or otherwise internal.
func isPrivateIP(ip net.IP) bool {
	// Check for loopback (127.x.x.x)
	if ip.IsLoopback() {
		return true
	}

	// Check for link-local (169.254.x.x for IPv4, fe80::/10 for IPv6)
	if ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() {
		return true
	}

	// Check for private ranges
	if ip.IsPrivate() {
		return true
	}

	// Check for unspecified (0.0.0.0 or ::)
	if ip.IsUnspecified() {
		return true
	}

	// Block IPv6 unique local addresses (fc00::/7)
	if ip.To4() == nil && len(ip) == net.IPv6len {
		if ip[0] == 0xfc || ip[0] == 0xfd {
			return true
		}
	}

	// Block cloud metadata IP (169.254.169.254)
	if ip.Equal(net.ParseIP("169.254.169.254")) {
		return true
	}

	return false
}
