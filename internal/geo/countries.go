package geo

import (
	"strings"
	"sync"
	"unicode"

	"golang.org/x/text/language"
	"golang.org/x/text/language/display"
)

var (
	countryOnce sync.Once
	nameToISO2  map[string]string
)

func normalizeCountryName(s string) string {
	s = strings.TrimSpace(strings.ToLower(s))
	var b strings.Builder
	b.Grow(len(s))
	space := false
	for _, r := range s {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			if space && b.Len() > 0 {
				b.WriteByte(' ')
			}
			space = false
			b.WriteRune(r)
			continue
		}
		space = true
	}
	return strings.TrimSpace(b.String())
}

func initCountryMap() {
	nameToISO2 = make(map[string]string, 512)
	namer := display.English.Regions()
	for _, region := range language.Supported.Regions() {
		iso2 := region.String()
		if len(iso2) != 2 {
			continue
		}
		name := namer.Name(region)
		if name == "" {
			continue
		}
		nameToISO2[normalizeCountryName(name)] = iso2
	}

	// Common aliases / variations.
	aliases := map[string]string{
		"usa":                      "US",
		"united states":            "US",
		"united states of america": "US",
		"uk":                       "GB",
		"united kingdom":           "GB",
		"great britain":            "GB",
		"russia":                   "RU",
		"iran":                     "IR",
		"vietnam":                  "VN",
		"south korea":              "KR",
		"north korea":              "KP",
		"bolivia":                  "BO",
		"venezuela":                "VE",
		"tanzania":                 "TZ",
		"laos":                     "LA",
		"moldova":                  "MD",
		"syria":                    "SY",
	}
	for k, v := range aliases {
		nameToISO2[normalizeCountryName(k)] = v
	}
}

// ISO2FromCountry attempts to map an English country name to an ISO-3166-1 alpha-2 code.
// If input already looks like a 2-letter code, it is normalized and returned.
func ISO2FromCountry(country string) (string, bool) {
	c := strings.TrimSpace(country)
	if len(c) == 2 {
		return strings.ToUpper(c), true
	}
	countryOnce.Do(initCountryMap)
	iso2, ok := nameToISO2[normalizeCountryName(c)]
	return iso2, ok
}
