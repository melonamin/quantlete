package badges

import (
	"bytes"
	"html/template"
	"strings"
)

// BadgeData contains the data to render a badge.
type BadgeData struct {
	Title    string
	Value    string
	Subtitle string
}

// svgTemplate is the template for rendering badges.
// Uses html/template for proper escaping.
var svgTemplate = template.Must(template.New("badge").Parse(`<svg width="{{.Width}}" height="{{.Height}}" viewBox="0 0 {{.Width}} {{.Height}}" xmlns="http://www.w3.org/2000/svg">
  {{if .Background}}<rect width="{{.Width}}" height="{{.Height}}" fill="{{.Background}}" rx="8" ry="8"/>{{end}}
  <rect x="1" y="1" width="{{.Width2}}" height="{{.Height2}}" fill="none" stroke="{{.Muted}}" stroke-width="1" stroke-opacity="0.3" rx="7" ry="7"/>
  <text x="20" y="{{.TitleY}}" fill="{{.Muted}}" font-size="{{.TitleSize}}" font-family="'JetBrains Mono', monospace" font-weight="500" style="letter-spacing: 0.05em">{{.Title}}</text>
  <text x="20" y="{{.ValueY}}" fill="{{.Accent}}" font-size="{{.ValueSize}}" font-family="'JetBrains Mono', monospace" font-weight="700">{{.Value}}</text>
  {{if .Subtitle}}<text x="20" y="{{.SubtitleY}}" fill="{{.Muted}}" font-size="{{.SubtitleSize}}" font-family="'JetBrains Mono', monospace" font-weight="400">{{.Subtitle}}</text>{{end}}
</svg>`))

// templateData combines theme, size, and badge data for template rendering.
type templateData struct {
	// Dimensions
	Width   int
	Height  int
	Width2  int // Width - 2 for inner border
	Height2 int // Height - 2 for inner border

	// Colors
	Background string
	Foreground string
	Accent     string
	Muted      string

	// Font sizes
	TitleSize    int
	ValueSize    int
	SubtitleSize int

	// Vertical positions
	TitleY    int
	ValueY    int
	SubtitleY int

	// Content
	Title    string
	Value    string
	Subtitle string
}

// RenderSVG renders a badge as SVG.
func RenderSVG(data BadgeData, theme Theme, size Size, bg Background) (string, error) {
	padding := 20

	// Calculate vertical positions
	titleY := padding + size.TitleSize
	valueY := size.Height/2 + size.ValueSize/3
	if data.Subtitle == "" {
		valueY += 4
	}
	subtitleY := valueY + size.SubtitleSize + 8

	td := templateData{
		Width:        size.Width,
		Height:       size.Height,
		Width2:       size.Width - 2,
		Height2:      size.Height - 2,
		Background:   bg.Color, // Use background color (empty = transparent)
		Foreground:   theme.Foreground,
		Accent:       theme.Accent,
		Muted:        theme.Muted,
		TitleSize:    size.TitleSize,
		ValueSize:    size.ValueSize,
		SubtitleSize: size.SubtitleSize,
		TitleY:       titleY,
		ValueY:       valueY,
		SubtitleY:    subtitleY,
		Title:        strings.ToUpper(data.Title),
		Value:        data.Value,
		Subtitle:     data.Subtitle,
	}

	var buf bytes.Buffer
	if err := svgTemplate.Execute(&buf, td); err != nil {
		return "", err
	}

	return buf.String(), nil
}
