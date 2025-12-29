package challenges

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
)

// BadgeDownloader downloads and stores challenge badge images locally.
type BadgeDownloader struct {
	dataDir string
}

// NewBadgeDownloader creates a new badge downloader that stores images in the given data directory.
func NewBadgeDownloader(dataDir string) *BadgeDownloader {
	return &BadgeDownloader{dataDir: dataDir}
}

// Download fetches a badge image from the given URL and stores it locally.
// Returns the local path relative to the data directory, or empty string on failure.
func (d *BadgeDownloader) Download(badgeURL string) (string, error) {
	if badgeURL == "" {
		return "", nil
	}

	// Create challenges directory if it doesn't exist
	badgesDir := filepath.Join(d.dataDir, "challenges")
	if err := os.MkdirAll(badgesDir, 0o750); err != nil {
		return "", fmt.Errorf("failed to create badges directory: %w", err)
	}

	// Determine file extension from URL
	ext := ".png"
	if strings.Contains(strings.ToLower(badgeURL), ".jpg") || strings.Contains(strings.ToLower(badgeURL), ".jpeg") {
		ext = ".jpg"
	} else if strings.Contains(strings.ToLower(badgeURL), ".webp") {
		ext = ".webp"
	}

	// Generate unique filename
	filename := uuid.NewString() + ext
	localPath := filepath.Join("challenges", filename)
	fullPath := filepath.Join(d.dataDir, localPath)

	// Download the image
	resp, err := http.Get(badgeURL) //nolint:gosec // URL comes from Strava
	if err != nil {
		return "", fmt.Errorf("failed to download badge: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("badge download returned status %d", resp.StatusCode)
	}

	// Create the file
	f, err := os.Create(fullPath) //nolint:gosec // fullPath is constructed from trusted data directory
	if err != nil {
		return "", fmt.Errorf("failed to create badge file: %w", err)
	}
	defer func() { _ = f.Close() }()

	// Copy the image data (limit to 5MB)
	if _, err := io.Copy(f, io.LimitReader(resp.Body, 5<<20)); err != nil {
		_ = os.Remove(fullPath)
		return "", fmt.Errorf("failed to write badge file: %w", err)
	}

	return localPath, nil
}

// BadgeExists checks if a badge file exists at the given local path.
func (d *BadgeDownloader) BadgeExists(localPath string) bool {
	if localPath == "" {
		return false
	}
	fullPath := filepath.Join(d.dataDir, localPath)
	_, err := os.Stat(fullPath)
	return err == nil
}
