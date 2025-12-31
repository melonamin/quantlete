import SwiftUI
import WebKit

/// A SwiftUI view that wraps WKWebView for displaying the dashboard
struct DashboardWebView: NSViewRepresentable {
    let url: URL

    func makeNSView(context: Context) -> WKWebView {
        let configuration = WKWebViewConfiguration()

        // Enable developer tools in debug builds
        #if DEBUG
        configuration.preferences.setValue(true, forKey: "developerExtrasEnabled")
        #endif

        let webView = WKWebView(frame: .zero, configuration: configuration)
        webView.navigationDelegate = context.coordinator
        webView.allowsBackForwardNavigationGestures = true

        return webView
    }

    func updateNSView(_ webView: WKWebView, context: Context) {
        // Only load if URL changed or first load
        if webView.url != url {
            let request = URLRequest(url: url)
            webView.load(request)
        }
    }

    func makeCoordinator() -> Coordinator {
        Coordinator()
    }

    class Coordinator: NSObject, WKNavigationDelegate {
        /// Handle navigation policy - keep internal links in webview, open external in browser
        func webView(
            _ webView: WKWebView,
            decidePolicyFor navigationAction: WKNavigationAction,
            decisionHandler: @escaping (WKNavigationActionPolicy) -> Void
        ) {
            guard let url = navigationAction.request.url else {
                decisionHandler(.allow)
                return
            }

            // Allow localhost navigation (our dashboard)
            if url.host == "localhost" || url.host == "127.0.0.1" {
                decisionHandler(.allow)
                return
            }

            // Open external links in system browser
            if navigationAction.navigationType == .linkActivated {
                NSWorkspace.shared.open(url)
                decisionHandler(.cancel)
                return
            }

            // Allow other requests (e.g., OAuth redirects that come back)
            decisionHandler(.allow)
        }

        /// Handle new window requests (target="_blank") - open in browser
        func webView(
            _ webView: WKWebView,
            createWebViewWith configuration: WKWebViewConfiguration,
            for navigationAction: WKNavigationAction,
            windowFeatures: WKWindowFeatures
        ) -> WKWebView? {
            if let url = navigationAction.request.url {
                NSWorkspace.shared.open(url)
            }
            return nil
        }
    }
}

/// The dashboard window content view
struct DashboardWindowContent: View {
    let port: Int

    var body: some View {
        DashboardWebView(url: URL(string: "http://localhost:\(port)")!)
            .frame(minWidth: 800, minHeight: 600)
    }
}
