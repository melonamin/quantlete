import AppKit
import SwiftUI

/// Manages the dashboard window lifecycle
@MainActor
final class DashboardWindowController {
    /// Shared instance for single-window pattern
    static let shared = DashboardWindowController()

    private var window: NSWindow?

    private init() {}

    private var currentPort: Int = 8081

    /// Show the dashboard window, creating it if needed
    func showWindow(port: Int = 8081) {
        // Show in Dock and Cmd+Tab when dashboard is open
        NSApp.setActivationPolicy(.regular)

        // If port changed, close existing window to recreate with new URL
        if let existingWindow = window, port == currentPort {
            // Bring existing window to front
            existingWindow.makeKeyAndOrderFront(nil)
            NSApp.activate(ignoringOtherApps: true)
            return
        } else if window != nil && port != currentPort {
            // Port changed, close old window
            window?.close()
            window = nil
        }

        currentPort = port

        // Create new window
        let contentView = DashboardWindowContent(port: port)
        let hostingView = NSHostingView(rootView: contentView)

        let newWindow = NSWindow(
            contentRect: NSRect(x: 0, y: 0, width: 1200, height: 800),
            styleMask: [.titled, .closable, .miniaturizable, .resizable],
            backing: .buffered,
            defer: false
        )

        newWindow.contentView = hostingView
        newWindow.title = "Quantlete"
        newWindow.minSize = NSSize(width: 800, height: 600)
        newWindow.isReleasedWhenClosed = false
        newWindow.delegate = WindowDelegate.shared

        // Restore saved frame or center
        if let savedFrame = UserDefaults.standard.string(forKey: "DashboardWindowFrame"),
           let frame = NSRectFromString(savedFrame) as NSRect?,
           frame.width > 0
        {
            newWindow.setFrame(frame, display: true)
        } else {
            newWindow.center()
        }

        self.window = newWindow
        newWindow.makeKeyAndOrderFront(nil)
        NSApp.activate(ignoringOtherApps: true)
    }

    /// Close the dashboard window
    func closeWindow() {
        window?.close()
    }

    /// Check if window is currently visible
    var isWindowVisible: Bool {
        window?.isVisible ?? false
    }

    /// Save window frame for restoration
    func saveWindowFrame() {
        guard let window = window else { return }
        let frameString = NSStringFromRect(window.frame)
        UserDefaults.standard.set(frameString, forKey: "DashboardWindowFrame")
    }
}

/// Window delegate to handle close events
private class WindowDelegate: NSObject, NSWindowDelegate {
    static let shared = WindowDelegate()

    func windowWillClose(_ notification: Notification) {
        // Save frame before closing
        Task { @MainActor in
            DashboardWindowController.shared.saveWindowFrame()
            // Hide from Dock when dashboard is closed (back to menu bar only)
            NSApp.setActivationPolicy(.accessory)
        }
    }

    func windowDidResize(_ notification: Notification) {
        // Save frame on resize
        Task { @MainActor in
            DashboardWindowController.shared.saveWindowFrame()
        }
    }

    func windowDidMove(_ notification: Notification) {
        // Save frame on move
        Task { @MainActor in
            DashboardWindowController.shared.saveWindowFrame()
        }
    }
}
