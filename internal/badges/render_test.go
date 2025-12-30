package badges

import (
	"strings"
	"testing"
)

func TestRenderSVG(t *testing.T) {
	tests := []struct {
		name     string
		data     BadgeData
		theme    Theme
		size     Size
		wantErr  bool
		contains []string
	}{
		{
			name: "basic badge",
			data: BadgeData{
				Title: "Total Distance",
				Value: "1,234 km",
			},
			theme: DefaultTheme(),
			size:  DefaultSize(),
			contains: []string{
				"<svg",
				"TOTAL DISTANCE", // Title is uppercased
				"1,234 km",
				"</svg>",
			},
		},
		{
			name: "badge with subtitle",
			data: BadgeData{
				Title:    "Eddington",
				Value:    "E42",
				Subtitle: "42 days with 42+ km",
			},
			theme: DefaultTheme(),
			size:  DefaultSize(),
			contains: []string{
				"EDDINGTON",
				"E42",
				"42 days with 42&#43; km", // + is HTML-escaped
			},
		},
		{
			name: "strava theme",
			data: BadgeData{
				Title: "Distance",
				Value: "500 mi",
			},
			theme: GetTheme("strava"),
			size:  DefaultSize(),
			contains: []string{
				"#fc4c02", // Strava orange accent
			},
		},
		{
			name: "compact size",
			data: BadgeData{
				Title: "Time",
				Value: "100h",
			},
			theme: DefaultTheme(),
			size:  GetSize("compact"),
			contains: []string{
				`width="200"`,
				`height="80"`,
			},
		},
	}

	// Test transparent background separately
	t.Run("transparent background", func(t *testing.T) {
		svg, err := RenderSVG(BadgeData{
			Title: "Test",
			Value: "123",
		}, DefaultTheme(), DefaultSize(), GetBackground("transparent"))
		if err != nil {
			t.Errorf("RenderSVG() error = %v", err)
			return
		}
		// Should NOT contain background rect (fill="#1a1a1f")
		if strings.Contains(svg, `fill="#1a1a1f" rx="8"`) {
			t.Error("transparent background should not render background rect")
		}
	})

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svg, err := RenderSVG(tt.data, tt.theme, tt.size, DefaultBackground())
			if (err != nil) != tt.wantErr {
				t.Errorf("RenderSVG() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			for _, want := range tt.contains {
				if !strings.Contains(svg, want) {
					t.Errorf("RenderSVG() output missing %q", want)
				}
			}
		})
	}
}

func TestGetTheme(t *testing.T) {
	tests := []struct {
		id       string
		wantName string
	}{
		{"terminal", "Terminal"},
		{"strava", "Strava"},
		{"amber", "Amber"},
		{"cyan", "Cyan"},
		{"minimal", "Minimal"},
		{"unknown", "Terminal"}, // Falls back to default
		{"", "Terminal"},        // Falls back to default
	}

	for _, tt := range tests {
		t.Run(tt.id, func(t *testing.T) {
			theme := GetTheme(tt.id)
			if theme.Name != tt.wantName {
				t.Errorf("GetTheme(%q) name = %q, want %q", tt.id, theme.Name, tt.wantName)
			}
		})
	}
}

func TestGetBackground(t *testing.T) {
	tests := []struct {
		id        string
		wantID    string
		wantColor string
	}{
		{"dark", "dark", "#1a1a1f"},
		{"light", "light", "#ffffff"},
		{"transparent", "transparent", ""},
		{"unknown", "dark", "#1a1a1f"}, // Falls back to default
		{"", "dark", "#1a1a1f"},        // Falls back to default
	}

	for _, tt := range tests {
		t.Run(tt.id, func(t *testing.T) {
			bg := GetBackground(tt.id)
			if bg.ID != tt.wantID {
				t.Errorf("GetBackground(%q) ID = %q, want %q", tt.id, bg.ID, tt.wantID)
			}
			if bg.Color != tt.wantColor {
				t.Errorf("GetBackground(%q) Color = %q, want %q", tt.id, bg.Color, tt.wantColor)
			}
		})
	}
}

func TestGetSize(t *testing.T) {
	tests := []struct {
		id        string
		wantWidth int
	}{
		{"compact", 200},
		{"standard", 400},
		{"wide", 500},
		{"unknown", 400}, // Falls back to default
		{"", 400},        // Falls back to default
	}

	for _, tt := range tests {
		t.Run(tt.id, func(t *testing.T) {
			size := GetSize(tt.id)
			if size.Width != tt.wantWidth {
				t.Errorf("GetSize(%q) width = %d, want %d", tt.id, size.Width, tt.wantWidth)
			}
		})
	}
}

func TestFormatDistance(t *testing.T) {
	tests := []struct {
		meters float64
		unit   UnitSystem
		want   string
	}{
		{1000, UnitMetric, "1.0 km"},
		{1500, UnitMetric, "1.5 km"},
		{12345.67, UnitMetric, "12.3 km"},
		{1609.34, UnitImperial, "1.0 mi"},
		{16093.4, UnitImperial, "10.0 mi"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := FormatDistance(tt.meters, tt.unit)
			if got != tt.want {
				t.Errorf("FormatDistance(%v, %v) = %q, want %q", tt.meters, tt.unit, got, tt.want)
			}
		})
	}
}

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		seconds int
		want    string
	}{
		{3600, "1h 0m"},
		{3661, "1h 1m"},
		{7200, "2h 0m"},
		{86400, "24h 0m"},
		{90061, "25h 1m"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := FormatDuration(tt.seconds)
			if got != tt.want {
				t.Errorf("FormatDuration(%d) = %q, want %q", tt.seconds, got, tt.want)
			}
		})
	}
}

func TestFormatElevation(t *testing.T) {
	tests := []struct {
		meters float64
		unit   UnitSystem
		want   string
	}{
		{1000, UnitMetric, "1,000 m"},
		{12345, UnitMetric, "12,345 m"},
		{304.8, UnitImperial, "1,000 ft"},
		{3048, UnitImperial, "10,000 ft"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := FormatElevation(tt.meters, tt.unit)
			if got != tt.want {
				t.Errorf("FormatElevation(%v, %v) = %q, want %q", tt.meters, tt.unit, got, tt.want)
			}
		})
	}
}

func TestFormatNumber(t *testing.T) {
	tests := []struct {
		n    int
		want string
	}{
		{0, "0"},
		{999, "999"},
		{1000, "1,000"},
		{1234567, "1,234,567"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := FormatNumber(tt.n)
			if got != tt.want {
				t.Errorf("FormatNumber(%d) = %q, want %q", tt.n, got, tt.want)
			}
		})
	}
}
