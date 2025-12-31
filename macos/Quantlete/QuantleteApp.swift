import SwiftUI
import ServiceManagement

@main
struct QuantleteApp: App {
    @StateObject private var serverManager = ServerManager()
    @AppStorage("hasLaunchedBefore") private var hasLaunchedBefore = false
    @State private var isLoginItemEnabled = SMAppService.mainApp.status == .enabled

    var body: some Scene {
        MenuBarExtra {
            menuContent
        } label: {
            Label("Quantlete", systemImage: iconName)
        }
    }

    private var iconName: String {
        // Show sync icon when syncing
        if let progress = serverManager.importProgress,
           progress.status == "running" {
            return "arrow.triangle.2.circlepath"
        }

        switch serverManager.status {
        case .running:
            return "chart.line.uptrend.xyaxis"
        case .starting:
            return "chart.line.uptrend.xyaxis"
        case .stopped:
            return "chart.line.uptrend.xyaxis"
        case .error:
            return "exclamationmark.triangle"
        }
    }

    @ViewBuilder
    private var menuContent: some View {
        dashboardSection
        Divider()
        syncStatusSection
        Divider()
        loginItemSection
        Divider()
        advancedSection
        Divider()
        appSection
    }

    private var statusColor: Color {
        switch serverManager.status {
        case .running:
            return .green
        case .starting:
            return .yellow
        case .stopped, .error:
            return .red
        }
    }

    @ViewBuilder
    private var syncStatusSection: some View {
        if let progress = serverManager.importProgress, progress.status == "running" {
            // Sync in progress
            Text(phaseDescription(progress.phase))

            // Progress info
            let (current, total) = currentProgress(progress)
            if total > 0 {
                ProgressView(value: Double(current), total: Double(total))
                    .progressViewStyle(.linear)
                    .frame(width: 150)
                Text("\(current) / \(total)")
                    .font(.caption)
                    .foregroundColor(.secondary)
            }

            // Rate limit
            if progress.rateLimitLimit15Min > 0 {
                Text("API: \(progress.rateLimitUsed15Min)/\(progress.rateLimitLimit15Min)")
                    .font(.caption)
                    .foregroundColor(.secondary)
            }

            // ETA
            if let eta = progress.estimatedETA, !eta.isEmpty {
                Text("ETA: \(eta)")
                    .font(.caption)
                    .foregroundColor(.secondary)
            }

            // Rate limit warning
            if progress.waitingForRateLimit {
                Text("Waiting for rate limit...")
                    .font(.caption)
                    .foregroundColor(.orange)
            }

            Divider()

            Button("Cancel Sync") {
                serverManager.cancelSync()
            }
        } else if let progress = serverManager.importProgress, progress.status == "failed" {
            // Sync failed
            Text(progress.error ?? "Sync failed")
                .font(.caption)
                .foregroundColor(.red)

            Button {
                serverManager.startSync()
            } label: {
                HStack {
                    Circle()
                        .fill(statusColor)
                        .frame(width: 8, height: 8)
                    Text("Retry Sync")
                }
            }
            .disabled(serverManager.status != .running)
        } else {
            // Idle or completed - show Sync Now with status dot
            if let errorMessage = serverManager.errorMessage {
                Text(errorMessage)
                    .font(.caption)
                    .foregroundColor(.red)
            }

            Button {
                serverManager.startSync()
            } label: {
                HStack {
                    Circle()
                        .fill(statusColor)
                        .frame(width: 8, height: 8)
                    Text("Sync Now")
                }
            }
            .disabled(serverManager.status != .running)
        }
    }

    private func phaseDescription(_ phase: String) -> String {
        switch phase {
        case "activities":
            return "Syncing activities..."
        case "gear":
            return "Syncing gear..."
        case "streams":
            return "Fetching GPS/HR data..."
        case "activity_details":
            return "Fetching activity details..."
        case "segment_details":
            return "Fetching segments..."
        case "photos":
            return "Fetching photos..."
        case "completed":
            return "Completed"
        default:
            return "Syncing..."
        }
    }

    private func currentProgress(_ progress: ImportProgress) -> (current: Int, total: Int) {
        switch progress.phase {
        case "activities":
            return (progress.activitiesDone, progress.activitiesTotal)
        case "streams":
            return (progress.streamsDone, progress.streamsTotal)
        case "activity_details":
            return (progress.detailsDone, progress.detailsTotal)
        case "segment_details":
            return (progress.segmentsDone, progress.segmentsTotal)
        case "photos":
            return (progress.photosDone, progress.photosTotal)
        default:
            return (0, 0)
        }
    }

    @ViewBuilder
    private var dashboardSection: some View {
        Button("Open Dashboard") {
            openDashboard()
        }
        .keyboardShortcut("o", modifiers: .command)
    }

    @ViewBuilder
    private var advancedSection: some View {
        Menu("Advanced") {
            Button("Open in Browser") {
                openInBrowser()
            }

            Divider()

            switch serverManager.status {
            case .stopped, .error:
                Button("Start Server") {
                    serverManager.start()
                }
            case .running:
                Button("Stop Server") {
                    serverManager.stop()
                }
                Button("Restart Server") {
                    serverManager.restart()
                }
            case .starting:
                Button("Starting...") {}
                    .disabled(true)
            }
        }
    }

    private func openInBrowser() {
        let url = URL(string: "http://localhost:\(serverManager.port)")!
        NSWorkspace.shared.open(url)
    }

    @ViewBuilder
    private var loginItemSection: some View {
        Toggle("Start at Login", isOn: $isLoginItemEnabled)
            .onChange(of: isLoginItemEnabled) { _, newValue in
                setLoginItem(enabled: newValue)
            }
    }

    @ViewBuilder
    private var appSection: some View {
        Button("About Quantlete") {
            showAbout()
        }
        Button("Quit") {
            serverManager.stop()
            NSApplication.shared.terminate(nil)
        }
        .keyboardShortcut("q", modifiers: .command)
    }

    private func openDashboard() {
        DashboardWindowController.shared.showWindow(port: serverManager.port)
    }

    private func setLoginItem(enabled: Bool) {
        do {
            if enabled {
                try SMAppService.mainApp.register()
            } else {
                try SMAppService.mainApp.unregister()
            }
        } catch {
            isLoginItemEnabled = SMAppService.mainApp.status == .enabled
        }
    }

    private func showAbout() {
        NSApplication.shared.orderFrontStandardAboutPanel(
            options: [
                .applicationName: "Quantlete",
                .applicationVersion: Bundle.main.infoDictionary?["CFBundleShortVersionString"] as? String ?? "1.0",
                .credits: NSAttributedString(string: "A self-hosted analytics dashboard for Strava activities.")
            ]
        )
        NSApplication.shared.activate(ignoringOtherApps: true)
    }
}
