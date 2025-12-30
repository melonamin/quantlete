import Foundation
import os.log

enum ServerStatus: Equatable {
    case stopped
    case starting
    case running
    case error

    var description: String {
        switch self {
        case .stopped:
            return "Server Stopped"
        case .starting:
            return "Server Starting..."
        case .running:
            return "Server Running"
        case .error:
            return "Server Error"
        }
    }
}

@MainActor
class ServerManager: ObservableObject {
    @Published private(set) var status: ServerStatus = .stopped
    @Published private(set) var errorMessage: String?

    let port: Int = 8081

    private var process: Process?
    private var healthCheckTimer: Timer?
    private var restartAttempts = 0
    private let maxRestartAttempts = 3
    private let restartBackoffSeconds: [Double] = [1, 2, 5]

    private let logger = Logger(subsystem: "app.stata", category: "ServerManager")

    private var hasOpenedDashboard = false

    init() {
        NotificationCenter.default.addObserver(
            forName: NSApplication.willTerminateNotification,
            object: nil,
            queue: .main
        ) { [weak self] _ in
            self?.stop()
        }

        start()
    }

    deinit {
        stop()
    }

    func start() {
        guard status == .stopped || status == .error else {
            logger.debug("Server already running or starting")
            return
        }

        status = .starting
        errorMessage = nil

        if isPortInUse(port: port) {
            status = .error
            errorMessage = "Port \(port) is already in use"
            logger.error("Port \(self.port) is already in use")
            return
        }

        guard let binaryPath = locateBinary() else {
            status = .error
            errorMessage = "Server binary not found"
            logger.error("Server binary not found in app bundle")
            return
        }

        ensureDataDirectory()

        let process = Process()
        process.executableURL = URL(fileURLWithPath: binaryPath)
        process.arguments = ["serve", "--port", "\(port)"]
        process.environment = buildEnvironment()

        let outputPipe = Pipe()
        let errorPipe = Pipe()
        process.standardOutput = outputPipe
        process.standardError = errorPipe

        setupOutputHandlers(output: outputPipe, error: errorPipe)

        process.terminationHandler = { [weak self] process in
            Task { @MainActor in
                self?.handleTermination(exitCode: process.terminationStatus)
            }
        }

        do {
            try process.run()
            self.process = process
            logger.info("Server process started with PID \(process.processIdentifier)")
            startHealthChecks()
        } catch {
            status = .error
            errorMessage = error.localizedDescription
            logger.error("Failed to start server: \(error.localizedDescription)")
        }
    }

    func stop() {
        healthCheckTimer?.invalidate()
        healthCheckTimer = nil

        guard let process = process, process.isRunning else {
            status = .stopped
            return
        }

        logger.info("Stopping server (SIGTERM)")
        process.terminate()

        DispatchQueue.global().asyncAfter(deadline: .now() + 5) { [weak self] in
            if process.isRunning {
                self?.logger.warning("Server did not terminate gracefully, sending SIGKILL")
                process.interrupt()
            }
        }

        process.waitUntilExit()
        self.process = nil
        status = .stopped
        restartAttempts = 0
        logger.info("Server stopped")
    }

    func restart() {
        stop()
        DispatchQueue.main.asyncAfter(deadline: .now() + 0.5) { [weak self] in
            self?.start()
        }
    }

    private func locateBinary() -> String? {
        if let bundlePath = Bundle.main.path(forResource: "stata", ofType: nil) {
            return bundlePath
        }

        let resourcesPath = Bundle.main.bundlePath + "/Contents/Resources/stata"
        if FileManager.default.fileExists(atPath: resourcesPath) {
            return resourcesPath
        }

        return nil
    }

    private func buildEnvironment() -> [String: String] {
        var env = ProcessInfo.processInfo.environment
        env["STATA_STORAGE_DATA_DIR"] = dataDirectory()
        env["STATA_SERVER_PORT"] = "\(port)"
        return env
    }

    private func dataDirectory() -> String {
        let appSupport = FileManager.default.urls(
            for: .applicationSupportDirectory,
            in: .userDomainMask
        ).first!

        return appSupport.appendingPathComponent("Stata").path
    }

    private func ensureDataDirectory() {
        let path = dataDirectory()
        if !FileManager.default.fileExists(atPath: path) {
            do {
                try FileManager.default.createDirectory(
                    atPath: path,
                    withIntermediateDirectories: true,
                    attributes: [.posixPermissions: 0o750]
                )
                logger.info("Created data directory at \(path)")
            } catch {
                logger.error("Failed to create data directory: \(error.localizedDescription)")
            }
        }
    }

    private func setupOutputHandlers(output: Pipe, error: Pipe) {
        output.fileHandleForReading.readabilityHandler = { [weak self] handle in
            let data = handle.availableData
            if let str = String(data: data, encoding: .utf8), !str.isEmpty {
                self?.logger.info("Server: \(str.trimmingCharacters(in: .whitespacesAndNewlines))")
            }
        }

        error.fileHandleForReading.readabilityHandler = { [weak self] handle in
            let data = handle.availableData
            if let str = String(data: data, encoding: .utf8), !str.isEmpty {
                self?.logger.error("Server error: \(str.trimmingCharacters(in: .whitespacesAndNewlines))")
            }
        }
    }

    private func handleTermination(exitCode: Int32) {
        logger.info("Server terminated with exit code \(exitCode)")
        process = nil
        healthCheckTimer?.invalidate()
        healthCheckTimer = nil

        if exitCode != 0 && restartAttempts < maxRestartAttempts {
            let backoff = restartBackoffSeconds[min(restartAttempts, restartBackoffSeconds.count - 1)]
            restartAttempts += 1
            logger.info("Attempting restart \(self.restartAttempts)/\(self.maxRestartAttempts) in \(backoff)s")

            status = .error
            errorMessage = "Crashed, restarting..."

            DispatchQueue.main.asyncAfter(deadline: .now() + backoff) { [weak self] in
                self?.start()
            }
        } else if exitCode != 0 {
            status = .error
            errorMessage = "Server crashed (exit code \(exitCode))"
        } else {
            status = .stopped
        }
    }

    private func startHealthChecks() {
        healthCheckTimer = Timer.scheduledTimer(withTimeInterval: 2.0, repeats: true) { [weak self] _ in
            Task { @MainActor in
                await self?.checkHealth()
            }
        }
    }

    private func checkHealth() async {
        let url = URL(string: "http://localhost:\(port)/api/v1/health")!
        var request = URLRequest(url: url)
        request.timeoutInterval = 2.0

        do {
            let (data, response) = try await URLSession.shared.data(for: request)

            guard let httpResponse = response as? HTTPURLResponse,
                  httpResponse.statusCode == 200 else {
                return
            }

            if let json = try? JSONSerialization.jsonObject(with: data) as? [String: Any],
               let status = json["status"] as? String,
               status == "ok" {
                if self.status != .running {
                    self.status = .running
                    self.errorMessage = nil
                    self.restartAttempts = 0
                    logger.info("Server is healthy")

                    if !hasOpenedDashboard && !UserDefaults.standard.bool(forKey: "hasLaunchedBefore") {
                        hasOpenedDashboard = true
                        UserDefaults.standard.set(true, forKey: "hasLaunchedBefore")
                        openDashboard()
                    }
                }
            }
        } catch {
            if self.status == .running {
                logger.warning("Health check failed: \(error.localizedDescription)")
            }
        }
    }

    private func openDashboard() {
        let url = URL(string: "http://localhost:\(port)")!
        NSWorkspace.shared.open(url)
        logger.info("Opened dashboard in browser")
    }

    private func isPortInUse(port: Int) -> Bool {
        let socketFD = socket(AF_INET, SOCK_STREAM, 0)
        guard socketFD >= 0 else { return false }
        defer { close(socketFD) }

        var addr = sockaddr_in()
        addr.sin_family = sa_family_t(AF_INET)
        addr.sin_port = in_port_t(port).bigEndian
        addr.sin_addr.s_addr = INADDR_ANY

        let result = withUnsafePointer(to: &addr) { ptr in
            ptr.withMemoryRebound(to: sockaddr.self, capacity: 1) { sockaddrPtr in
                bind(socketFD, sockaddrPtr, socklen_t(MemoryLayout<sockaddr_in>.size))
            }
        }

        return result != 0
    }
}
