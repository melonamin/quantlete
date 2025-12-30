import SwiftUI
import ServiceManagement

@main
struct StataApp: App {
    @StateObject private var serverManager = ServerManager()
    @AppStorage("hasLaunchedBefore") private var hasLaunchedBefore = false
    @State private var isLoginItemEnabled = SMAppService.mainApp.status == .enabled

    var body: some Scene {
        MenuBarExtra {
            menuContent
        } label: {
            Label("Stata", systemImage: iconName)
        }
    }

    private var iconName: String {
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
        statusSection
        Divider()
        dashboardSection
        Divider()
        serverControlSection
        Divider()
        loginItemSection
        Divider()
        appSection
    }

    @ViewBuilder
    private var statusSection: some View {
        HStack {
            Circle()
                .fill(statusColor)
                .frame(width: 8, height: 8)
            Text(serverManager.status.description)
        }
        .padding(.horizontal, 4)

        if let errorMessage = serverManager.errorMessage {
            Text(errorMessage)
                .font(.caption)
                .foregroundColor(.secondary)
        }
    }

    private var statusColor: Color {
        switch serverManager.status {
        case .running:
            return .green
        case .starting:
            return .yellow
        case .stopped:
            return .gray
        case .error:
            return .red
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
    private var serverControlSection: some View {
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

    @ViewBuilder
    private var loginItemSection: some View {
        Toggle("Start at Login", isOn: $isLoginItemEnabled)
            .onChange(of: isLoginItemEnabled) { _, newValue in
                setLoginItem(enabled: newValue)
            }
    }

    @ViewBuilder
    private var appSection: some View {
        Button("About Stata") {
            showAbout()
        }
        Button("Quit") {
            serverManager.stop()
            NSApplication.shared.terminate(nil)
        }
        .keyboardShortcut("q", modifiers: .command)
    }

    private func openDashboard() {
        let url = URL(string: "http://localhost:\(serverManager.port)")!
        NSWorkspace.shared.open(url)
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
                .applicationName: "Stata",
                .applicationVersion: Bundle.main.infoDictionary?["CFBundleShortVersionString"] as? String ?? "1.0",
                .credits: NSAttributedString(string: "A self-hosted analytics dashboard for Strava activities.")
            ]
        )
        NSApplication.shared.activate(ignoringOtherApps: true)
    }
}
