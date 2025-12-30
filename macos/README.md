# Stata macOS App

A lightweight menu bar application that manages the Stata server on macOS.

## Requirements

- macOS 13.0 (Ventura) or later
- Xcode 15.0 or later
- Go 1.21 or later (for building the server binary)

## Building

### Quick Build (for testing)

```bash
# From the project root
just build-macos
```

### Manual Build Steps

1. **Build the Go server binary:**

```bash
# For Apple Silicon
GOOS=darwin GOARCH=arm64 go build -o bin/stata-darwin-arm64 ./cmd/stata

# For Intel
GOOS=darwin GOARCH=amd64 go build -o bin/stata-darwin-amd64 ./cmd/stata

# Create universal binary (optional)
lipo -create -output bin/stata bin/stata-darwin-arm64 bin/stata-darwin-amd64
```

2. **Build the Swift app:**

```bash
cd macos
xcodebuild -project Stata.xcodeproj -scheme Stata -configuration Release build
```

3. **Copy server binary into app bundle:**

```bash
cp bin/stata ~/Library/Developer/Xcode/DerivedData/Stata-*/Build/Products/Release/Stata.app/Contents/Resources/
```

## Development

Open the project in Xcode:

```bash
open macos/Stata.xcodeproj
```

For development, you can run the Go server separately and the Swift app will detect it:

```bash
# Terminal 1: Run Go server
just dev-api

# Terminal 2: Run Swift app from Xcode
# Press Cmd+R in Xcode
```

## Code Signing & Distribution

### Development (unsigned)

For local testing, the app works without code signing. macOS may show a warning on first launch - right-click and choose "Open" to bypass.

### Release (signed & notarized)

1. **Set your Team ID in Xcode:**
   - Open `Stata.xcodeproj`
   - Select the Stata target
   - Go to Signing & Capabilities
   - Select your Development Team

2. **Build for release:**

```bash
xcodebuild -project Stata.xcodeproj \
  -scheme Stata \
  -configuration Release \
  -archivePath build/Stata.xcarchive \
  archive

xcodebuild -exportArchive \
  -archivePath build/Stata.xcarchive \
  -exportPath build/Stata \
  -exportOptionsPlist ExportOptions.plist
```

3. **Notarize:**

```bash
xcrun notarytool submit build/Stata/Stata.app \
  --apple-id YOUR_APPLE_ID \
  --team-id YOUR_TEAM_ID \
  --password YOUR_APP_SPECIFIC_PASSWORD \
  --wait

xcrun stapler staple build/Stata/Stata.app
```

4. **Create DMG:**

```bash
hdiutil create -volname Stata -srcfolder build/Stata/Stata.app -ov -format UDZO Stata.dmg
```

## Architecture

```
Stata.app/
├── Contents/
│   ├── MacOS/
│   │   └── Stata          # SwiftUI menu bar app
│   ├── Resources/
│   │   └── stata          # Go server binary
│   ├── Info.plist
│   └── Entitlements
```

The Swift app:
- Locates the embedded `stata` binary in Resources
- Sets `STATA_STORAGE_DATA_DIR` to `~/Library/Application Support/Stata/`
- Spawns the server as a child process
- Monitors health via `http://localhost:8081/api/v1/health`
- Provides menu bar controls (start/stop/restart)
- Supports "Start at Login" via SMAppService

## Data Location

All data is stored in:
```
~/Library/Application Support/Stata/
├── stata.db      # SQLite database
```

## Troubleshooting

### Server won't start

Check if port 8081 is already in use:
```bash
lsof -i :8081
```

### View logs

The app logs to the unified logging system. View in Console.app:
```bash
log stream --predicate 'subsystem == "app.stata"'
```

### Reset all data

```bash
rm -rf ~/Library/Application\ Support/Stata/
```
