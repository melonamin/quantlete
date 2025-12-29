package challenges

import (
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"
)

// ParsedProfileChallenge represents a challenge scraped from a Strava public profile.
type ParsedProfileChallenge struct {
	Name        string
	Slug        string
	BadgeURL    string
	CompletedOn time.Time
	ChallengeID string
	Teaser      string
}

var (
	// Matches trophy list items on public profile page
	trophyItemRe = regexp.MustCompile(`<li class="Trophies_listItem[^"]*">(?P<content>[\s\S]*?)</li>`)
	// Extracts challenge name from h4 tag
	challengeNameRe = regexp.MustCompile(`<h4[^>]*>(?P<name>.*?)</h4>`)
	// Extracts teaser from title attribute
	teaserRe = regexp.MustCompile(`<a[^>]+title="(?P<teaser>[^"]*)"[^>]*>`)
	// Extracts logo URL from img src
	logoURLRe = regexp.MustCompile(`<img[^>]+src="(?P<url>[^"]+)"`)
	// Extracts challenge slug from href
	challengeURLRe = regexp.MustCompile(`<a[^>]+href="/challenges/(?P<slug>[^"]+)"`)
	// Extracts challenge ID from badge URL pattern
	challengeIDRe = regexp.MustCompile(`/challenges/(?P<id>[^/]+)/`)
	// Extracts completion time from time tag
	timeTagRe = regexp.MustCompile(`<time[^>]*>(?P<time>.*?)</time>`)
)

// FetchPublicProfile fetches and parses challenges from a Strava public profile page.
func FetchPublicProfile(athleteID string) ([]ParsedProfileChallenge, error) {
	url := fmt.Sprintf("https://www.strava.com/athletes/%s", athleteID)
	resp, err := http.Get(url) //nolint:gosec // URL is constructed from trusted input
	if err != nil {
		return nil, fmt.Errorf("failed to fetch public profile: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("public profile returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, 10<<20))
	if err != nil {
		return nil, fmt.Errorf("failed to read profile body: %w", err)
	}

	return ParsePublicProfileHTML(string(body))
}

// ParsePublicProfileHTML parses challenges from Strava public profile HTML.
func ParsePublicProfileHTML(html string) ([]ParsedProfileChallenge, error) {
	matches := trophyItemRe.FindAllStringSubmatch(html, -1)
	if len(matches) == 0 {
		return nil, nil
	}

	var challenges []ParsedProfileChallenge
	seen := make(map[string]struct{})

	for _, match := range matches {
		if len(match) < 2 {
			continue
		}
		content := match[1]

		// Extract challenge name
		var name string
		if m := challengeNameRe.FindStringSubmatch(content); len(m) >= 2 {
			name = strings.TrimSpace(m[1])
		}
		if name == "" {
			continue
		}

		// Extract slug from URL
		var slug string
		if m := challengeURLRe.FindStringSubmatch(content); len(m) >= 2 {
			slug = strings.TrimSpace(m[1])
		}
		if slug == "" {
			continue
		}

		// Deduplicate
		if _, ok := seen[slug]; ok {
			continue
		}
		seen[slug] = struct{}{}

		// Extract badge URL
		var badgeURL string
		if m := logoURLRe.FindStringSubmatch(content); len(m) >= 2 {
			badgeURL = strings.TrimSpace(m[1])
		}

		// Extract challenge ID from badge URL
		var challengeID string
		if m := challengeIDRe.FindStringSubmatch(badgeURL); len(m) >= 2 {
			challengeID = m[1]
		}

		// Extract teaser
		var teaser string
		if m := teaserRe.FindStringSubmatch(content); len(m) >= 2 {
			teaser = strings.TrimSpace(m[1])
		}

		// Extract completion time (e.g., "Jan 2024")
		var completedOn time.Time
		if m := timeTagRe.FindStringSubmatch(content); len(m) >= 2 {
			timeStr := strings.TrimSpace(m[1])
			if timeStr != "" {
				// Try parsing "Jan 2024" format
				if t, err := time.Parse("Jan 2006", timeStr); err == nil {
					completedOn = t
				}
			}
		}

		if completedOn.IsZero() {
			// Default to current time if parsing fails
			completedOn = time.Now()
		}

		challenges = append(challenges, ParsedProfileChallenge{
			Name:        name,
			Slug:        slug,
			BadgeURL:    badgeURL,
			CompletedOn: completedOn,
			ChallengeID: challengeID,
			Teaser:      teaser,
		})
	}

	return challenges, nil
}
