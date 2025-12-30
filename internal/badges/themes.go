package badges

// Theme defines the color scheme for a badge.
type Theme struct {
	ID         string
	Name       string
	Background string
	Foreground string
	Accent     string
	Muted      string
}

// Size defines the dimensions of a badge.
type Size struct {
	ID           string
	Name         string
	Width        int
	Height       int
	TitleSize    int
	ValueSize    int
	SubtitleSize int
}

// UnitSystem represents the measurement system for formatting.
type UnitSystem int

const (
	UnitMetric UnitSystem = iota
	UnitImperial
)

// themes contains all available badge themes.
var themes = map[string]Theme{
	"terminal": {
		ID:         "terminal",
		Name:       "Terminal",
		Background: "#1a1a1f",
		Foreground: "#e5e5e5",
		Accent:     "#22c55e",
		Muted:      "#a1a1aa",
	},
	"strava": {
		ID:         "strava",
		Name:       "Strava",
		Background: "#1a1a1f",
		Foreground: "#e5e5e5",
		Accent:     "#fc4c02",
		Muted:      "#a1a1aa",
	},
	"amber": {
		ID:         "amber",
		Name:       "Amber",
		Background: "#1a1a1f",
		Foreground: "#e5e5e5",
		Accent:     "#f59e0b",
		Muted:      "#a1a1aa",
	},
	"cyan": {
		ID:         "cyan",
		Name:       "Cyan",
		Background: "#1a1a1f",
		Foreground: "#e5e5e5",
		Accent:     "#06b6d4",
		Muted:      "#a1a1aa",
	},
	"minimal": {
		ID:         "minimal",
		Name:       "Minimal",
		Background: "#ffffff",
		Foreground: "#1a1a1f",
		Accent:     "#1a1a1f",
		Muted:      "#71717a",
	},
}

// sizes contains all available badge sizes.
var sizes = map[string]Size{
	"compact": {
		ID:           "compact",
		Name:         "Compact",
		Width:        200,
		Height:       80,
		TitleSize:    10,
		ValueSize:    24,
		SubtitleSize: 10,
	},
	"standard": {
		ID:           "standard",
		Name:         "Standard",
		Width:        400,
		Height:       120,
		TitleSize:    12,
		ValueSize:    36,
		SubtitleSize: 12,
	},
	"wide": {
		ID:           "wide",
		Name:         "Wide",
		Width:        500,
		Height:       100,
		TitleSize:    12,
		ValueSize:    32,
		SubtitleSize: 12,
	},
}

// Background defines the background style.
type Background struct {
	ID    string
	Color string // empty string = transparent
}

// backgrounds contains all available badge backgrounds.
var backgrounds = map[string]Background{
	"dark": {
		ID:    "dark",
		Color: "#1a1a1f",
	},
	"light": {
		ID:    "light",
		Color: "#ffffff",
	},
	"transparent": {
		ID:    "transparent",
		Color: "", // empty = no background rect
	},
}

// GetTheme returns a theme by ID, defaulting to "terminal".
func GetTheme(id string) Theme {
	if t, ok := themes[id]; ok {
		return t
	}
	return themes["terminal"]
}

// GetSize returns a size by ID, defaulting to "standard".
func GetSize(id string) Size {
	if s, ok := sizes[id]; ok {
		return s
	}
	return sizes["standard"]
}

// GetBackground returns a background by ID, defaulting to "dark".
func GetBackground(id string) Background {
	if b, ok := backgrounds[id]; ok {
		return b
	}
	return backgrounds["dark"]
}

// DefaultTheme returns the default theme.
func DefaultTheme() Theme {
	return themes["terminal"]
}

// DefaultSize returns the default size.
func DefaultSize() Size {
	return sizes["standard"]
}

// DefaultBackground returns the default background.
func DefaultBackground() Background {
	return backgrounds["dark"]
}
