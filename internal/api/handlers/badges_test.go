package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestParseParams(t *testing.T) {
	tests := []struct {
		name       string
		queryTheme string
		querySize  string
		queryBg    string
		queryUnit  string
		wantUnit   string
		wantBg     string
	}{
		{
			name:       "defaults",
			queryTheme: "",
			querySize:  "",
			queryBg:    "",
			queryUnit:  "",
			wantUnit:   "metric",
			wantBg:     "dark",
		},
		{
			name:       "imperial unit",
			queryTheme: "terminal",
			querySize:  "standard",
			queryBg:    "",
			queryUnit:  "imperial",
			wantUnit:   "imperial",
			wantBg:     "dark",
		},
		{
			name:       "metric explicit",
			queryTheme: "strava",
			querySize:  "compact",
			queryBg:    "",
			queryUnit:  "metric",
			wantUnit:   "metric",
			wantBg:     "dark",
		},
		{
			name:       "light background",
			queryTheme: "",
			querySize:  "",
			queryBg:    "light",
			queryUnit:  "",
			wantUnit:   "metric",
			wantBg:     "light",
		},
		{
			name:       "transparent background",
			queryTheme: "",
			querySize:  "",
			queryBg:    "transparent",
			queryUnit:  "",
			wantUnit:   "metric",
			wantBg:     "transparent",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/badges/distance.svg", nil)
			q := req.URL.Query()
			if tt.queryTheme != "" {
				q.Set("theme", tt.queryTheme)
			}
			if tt.querySize != "" {
				q.Set("size", tt.querySize)
			}
			if tt.queryBg != "" {
				q.Set("bg", tt.queryBg)
			}
			if tt.queryUnit != "" {
				q.Set("unit", tt.queryUnit)
			}
			req.URL.RawQuery = q.Encode()

			_, _, bg, unit := parseParams(req)

			gotUnit := "metric"
			if unit == 1 { // UnitImperial = 1
				gotUnit = "imperial"
			}

			if gotUnit != tt.wantUnit {
				t.Errorf("parseParams() unit = %q, want %q", gotUnit, tt.wantUnit)
			}
			if bg.ID != tt.wantBg {
				t.Errorf("parseParams() bg = %q, want %q", bg.ID, tt.wantBg)
			}
		})
	}
}

func TestWriteSVG(t *testing.T) {
	rr := httptest.NewRecorder()

	testSVG := `<svg xmlns="http://www.w3.org/2000/svg"><text>Test</text></svg>`
	writeSVG(rr, testSVG)

	if rr.Code != http.StatusOK {
		t.Errorf("writeSVG() status = %d, want %d", rr.Code, http.StatusOK)
	}

	contentType := rr.Header().Get("Content-Type")
	if contentType != "image/svg+xml" {
		t.Errorf("writeSVG() Content-Type = %q, want %q", contentType, "image/svg+xml")
	}

	cacheControl := rr.Header().Get("Cache-Control")
	if cacheControl != "public, max-age=3600" {
		t.Errorf("writeSVG() Cache-Control = %q, want %q", cacheControl, "public, max-age=3600")
	}

	etag := rr.Header().Get("ETag")
	if etag == "" {
		t.Error("writeSVG() should set ETag header")
	}

	if !strings.HasPrefix(etag, `"`) || !strings.HasSuffix(etag, `"`) {
		t.Errorf("writeSVG() ETag should be quoted, got %q", etag)
	}

	body := rr.Body.String()
	if body != testSVG {
		t.Errorf("writeSVG() body = %q, want %q", body, testSVG)
	}
}

func TestWriteBadgeError(t *testing.T) {
	rr := httptest.NewRecorder()

	writeBadgeError(rr, "Test Error")

	if rr.Code != http.StatusOK {
		t.Errorf("writeBadgeError() status = %d, want %d", rr.Code, http.StatusOK)
	}

	contentType := rr.Header().Get("Content-Type")
	if contentType != "image/svg+xml" {
		t.Errorf("writeBadgeError() Content-Type = %q, want %q", contentType, "image/svg+xml")
	}

	cacheControl := rr.Header().Get("Cache-Control")
	if cacheControl != "no-cache" {
		t.Errorf("writeBadgeError() Cache-Control = %q, want %q", cacheControl, "no-cache")
	}

	body := rr.Body.String()
	if !strings.Contains(body, "<svg") {
		t.Error("writeBadgeError() should return SVG content")
	}
	if !strings.Contains(body, "Error") {
		t.Error("writeBadgeError() should contain 'Error' title")
	}
	if !strings.Contains(body, "Test Error") {
		t.Error("writeBadgeError() should contain error message")
	}
}

func TestFormatMonthName(t *testing.T) {
	tests := []struct {
		monthStr string
		want     string
	}{
		{"2024-01", "January 2024"},
		{"2024-02", "February 2024"},
		{"2024-03", "March 2024"},
		{"2024-04", "April 2024"},
		{"2024-05", "May 2024"},
		{"2024-06", "June 2024"},
		{"2024-07", "July 2024"},
		{"2024-08", "August 2024"},
		{"2024-09", "September 2024"},
		{"2024-10", "October 2024"},
		{"2024-11", "November 2024"},
		{"2024-12", "December 2024"},
		{"2023-06", "June 2023"},
		// Edge cases
		{"invalid", "invalid"},
		{"2024", "2024"},
		{"2024-13", "2024-13"},   // Invalid month
		{"2024-00", "2024-00"},   // Invalid month
		{"2024-XX", "2024-XX"},   // Non-numeric month
	}

	for _, tt := range tests {
		t.Run(tt.monthStr, func(t *testing.T) {
			got := formatMonthName(tt.monthStr)
			if got != tt.want {
				t.Errorf("formatMonthName(%q) = %q, want %q", tt.monthStr, got, tt.want)
			}
		})
	}
}

func TestBadgesHandler_GetDistanceBadge_NoAthlete(t *testing.T) {
	// Test that badge returns error when no athlete is authenticated
	// and public badges are not enabled
	handler := NewBadgesHandler(nil, nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/badges/distance.svg", nil)
	rr := httptest.NewRecorder()

	handler.GetDistanceBadge(rr, req)

	// Should return 200 with error SVG (badges always return 200 so they display)
	if rr.Code != http.StatusOK {
		t.Errorf("GetDistanceBadge() status = %d, want %d", rr.Code, http.StatusOK)
	}

	body := rr.Body.String()
	if !strings.Contains(body, "<svg") {
		t.Error("GetDistanceBadge() should return SVG content")
	}
	if !strings.Contains(body, "Not Available") {
		t.Error("GetDistanceBadge() should show 'Not Available' when no athlete")
	}
}

func TestBadgesHandler_GetTimeBadge_NoAthlete(t *testing.T) {
	handler := NewBadgesHandler(nil, nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/badges/time.svg", nil)
	rr := httptest.NewRecorder()

	handler.GetTimeBadge(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("GetTimeBadge() status = %d, want %d", rr.Code, http.StatusOK)
	}

	body := rr.Body.String()
	if !strings.Contains(body, "Not Available") {
		t.Error("GetTimeBadge() should show 'Not Available' when no athlete")
	}
}

func TestBadgesHandler_GetElevationBadge_NoAthlete(t *testing.T) {
	handler := NewBadgesHandler(nil, nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/badges/elevation.svg", nil)
	rr := httptest.NewRecorder()

	handler.GetElevationBadge(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("GetElevationBadge() status = %d, want %d", rr.Code, http.StatusOK)
	}

	body := rr.Body.String()
	if !strings.Contains(body, "Not Available") {
		t.Error("GetElevationBadge() should show 'Not Available' when no athlete")
	}
}

func TestBadgesHandler_GetActivitiesBadge_NoAthlete(t *testing.T) {
	handler := NewBadgesHandler(nil, nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/badges/activities.svg", nil)
	rr := httptest.NewRecorder()

	handler.GetActivitiesBadge(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("GetActivitiesBadge() status = %d, want %d", rr.Code, http.StatusOK)
	}

	body := rr.Body.String()
	if !strings.Contains(body, "Not Available") {
		t.Error("GetActivitiesBadge() should show 'Not Available' when no athlete")
	}
}

func TestBadgesHandler_GetEddingtonBadge_NoAthlete(t *testing.T) {
	handler := NewBadgesHandler(nil, nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/badges/eddington.svg", nil)
	rr := httptest.NewRecorder()

	handler.GetEddingtonBadge(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("GetEddingtonBadge() status = %d, want %d", rr.Code, http.StatusOK)
	}

	body := rr.Body.String()
	if !strings.Contains(body, "Not Available") {
		t.Error("GetEddingtonBadge() should show 'Not Available' when no athlete")
	}
}

func TestBadgesHandler_GetYearBadge_NoAthlete(t *testing.T) {
	handler := NewBadgesHandler(nil, nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/badges/year-2024.svg", nil)
	rr := httptest.NewRecorder()

	// Note: chi.URLParam won't work in this test without a router
	// but GetYearBadge should still handle the no-athlete case first
	handler.GetYearBadge(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("GetYearBadge() status = %d, want %d", rr.Code, http.StatusOK)
	}

	body := rr.Body.String()
	if !strings.Contains(body, "Not Available") {
		t.Error("GetYearBadge() should show 'Not Available' when no athlete")
	}
}

func TestBadgesHandler_GetMonthBadge_NoAthlete(t *testing.T) {
	handler := NewBadgesHandler(nil, nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/badges/month-2024-12.svg", nil)
	rr := httptest.NewRecorder()

	handler.GetMonthBadge(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("GetMonthBadge() status = %d, want %d", rr.Code, http.StatusOK)
	}

	body := rr.Body.String()
	if !strings.Contains(body, "Not Available") {
		t.Error("GetMonthBadge() should show 'Not Available' when no athlete")
	}
}
