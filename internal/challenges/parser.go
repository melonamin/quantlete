package challenges

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"
)

// GenerateChallengeID creates a deterministic challenge ID from completion date and name.
// Format: challenge-{YYYY-MM}_{sanitized_name}
// This prevents duplicate imports of the same challenge.
func GenerateChallengeID(completionDate time.Time, name string) string {
	// Sanitize name: replace whitespace with underscores, lowercase, truncate to 250 chars
	sanitized := strings.Map(func(r rune) rune {
		if unicode.IsSpace(r) {
			return '_'
		}
		return unicode.ToLower(r)
	}, name)
	if len(sanitized) > 250 {
		sanitized = sanitized[:250]
	}
	return fmt.Sprintf("challenge-%s_%s", completionDate.Format("2006-01"), sanitized)
}

type ParsedChallenge struct {
	Name           string
	Slug           string
	BadgeURL       string
	CompletionDate *time.Time
	Month          string // YYYY-MM
}

var (
	monthHeadingRe  = regexp.MustCompile(`(?i)\b(Jan(?:uary)?|Feb(?:ruary)?|Mar(?:ch)?|Apr(?:il)?|May|Jun(?:e)?|Jul(?:y)?|Aug(?:ust)?|Sep(?:tember)?|Oct(?:ober)?|Nov(?:ember)?|Dec(?:ember)?)\s+(\d{4})\b`)
	challengeHrefRe = regexp.MustCompile(`href=(?:"|')(?:(?:https?://www\.strava\.com)?/challenges/)([A-Za-z0-9_-]+)(?:/[^"']*)?(?:"|')`)
	imgSrcRe        = regexp.MustCompile(`(?i)<img[^>]+src=["']([^"']+)["']`)
	imgAltRe        = regexp.MustCompile(`(?i)<img[^>]+alt=["']([^"']+)["']`)
	datetimeRe      = regexp.MustCompile(`(?i)datetime=(?:"|')(\d{4}-\d{2}-\d{2})(?:[T ][^"']*)?(?:"|')`)
	dateTextRe      = regexp.MustCompile(`(?i)\b(Jan(?:uary)?|Feb(?:ruary)?|Mar(?:ch)?|Apr(?:il)?|May|Jun(?:e)?|Jul(?:y)?|Aug(?:ust)?|Sep(?:tember)?|Oct(?:ober)?|Nov(?:ember)?|Dec(?:ember)?)\s+(\d{1,2}),\s*(\d{4})\b`)
)

func monthNameToNumber(name string) int {
	n := strings.ToLower(strings.TrimSpace(name))
	switch n[:3] {
	case "jan":
		return 1
	case "feb":
		return 2
	case "mar":
		return 3
	case "apr":
		return 4
	case "may":
		return 5
	case "jun":
		return 6
	case "jul":
		return 7
	case "aug":
		return 8
	case "sep":
		return 9
	case "oct":
		return 10
	case "nov":
		return 11
	case "dec":
		return 12
	default:
		return 0
	}
}

type monthMarker struct {
	pos  int
	year int
	mon  int
}

func ParseTrophyCaseHTML(html string) ([]ParsedChallenge, error) {
	s := html

	var markers []monthMarker
	for _, m := range monthHeadingRe.FindAllStringSubmatchIndex(s, -1) {
		if len(m) < 6 {
			continue
		}
		monthName := s[m[2]:m[3]]
		yearStr := s[m[4]:m[5]]
		year, err := strconv.Atoi(yearStr)
		if err != nil {
			continue
		}
		mon := monthNameToNumber(monthName)
		if mon == 0 {
			continue
		}
		markers = append(markers, monthMarker{pos: m[0], year: year, mon: mon})
	}

	monthForPos := func(pos int) string {
		var best *monthMarker
		for i := range markers {
			if markers[i].pos <= pos {
				if best == nil || markers[i].pos > best.pos {
					best = &markers[i]
				}
			}
		}
		if best == nil {
			return ""
		}
		return fmt.Sprintf("%04d-%02d", best.year, best.mon)
	}

	type hit struct {
		pos  int
		slug string
	}
	var hits []hit
	for _, m := range challengeHrefRe.FindAllStringSubmatchIndex(s, -1) {
		if len(m) < 4 {
			continue
		}
		slug := s[m[2]:m[3]]
		hits = append(hits, hit{pos: m[0], slug: slug})
	}

	if len(hits) == 0 {
		return nil, nil
	}

	seen := make(map[string]struct{}, len(hits))
	out := make([]ParsedChallenge, 0, len(hits))

	for _, h := range hits {
		key := h.slug
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}

		start := h.pos
		end := h.pos + 2000
		if end > len(s) {
			end = len(s)
		}
		window := s[start:end]

		var badge string
		if m := imgSrcRe.FindStringSubmatch(window); len(m) >= 2 {
			badge = m[1]
		}

		name := ""
		if m := imgAltRe.FindStringSubmatch(window); len(m) >= 2 {
			name = strings.TrimSpace(m[1])
		}
		if name == "" {
			name = strings.ReplaceAll(h.slug, "-", " ")
		}

		var completion *time.Time
		if m := datetimeRe.FindStringSubmatch(window); len(m) >= 2 {
			if t, err := time.Parse("2006-01-02", m[1]); err == nil {
				completion = &t
			}
		}
		if completion == nil {
			if m := dateTextRe.FindStringSubmatch(window); len(m) >= 4 {
				mon := monthNameToNumber(m[1])
				day, _ := strconv.Atoi(m[2])
				year, _ := strconv.Atoi(m[3])
				if mon > 0 && day > 0 && year > 0 {
					t := time.Date(year, time.Month(mon), day, 0, 0, 0, 0, time.UTC)
					completion = &t
				}
			}
		}

		month := ""
		if completion != nil {
			month = fmt.Sprintf("%04d-%02d", completion.Year(), int(completion.Month()))
		} else {
			month = monthForPos(h.pos)
		}

		out = append(out, ParsedChallenge{
			Name:           name,
			Slug:           h.slug,
			BadgeURL:       badge,
			CompletionDate: completion,
			Month:          month,
		})
	}

	return out, nil
}
