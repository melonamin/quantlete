import { Apple, Globe, Server } from 'lucide-react'

export function Hero() {
  return (
    <section className="relative min-h-screen flex items-center justify-center px-6 py-24">
      {/* Background grid pattern */}
      <div
        className="absolute inset-0 opacity-[0.02]"
        style={{
          backgroundImage: `
            linear-gradient(oklch(0.72 0.19 145 / 0.5) 1px, transparent 1px),
            linear-gradient(90deg, oklch(0.72 0.19 145 / 0.5) 1px, transparent 1px)
          `,
          backgroundSize: '50px 50px',
        }}
      />

      <div className="relative z-10 max-w-4xl mx-auto text-center">
        {/* Terminal prompt */}
        <div className="animate-fade-in mb-8 inline-flex items-center gap-2 px-4 py-2 rounded-md bg-card border border-border-dim text-sm text-muted-foreground">
          <span className="text-terminal-green">$</span>
          <span>quantlete</span>
          <span className="text-terminal-dim">--version</span>
          <span className="text-terminal-amber">v1.0.0</span>
        </div>

        {/* Main headline */}
        <h1 className="animate-fade-in animate-delay-1 text-5xl md:text-7xl font-bold tracking-tight mb-6">
          <span className="text-terminal-green glow-green">Statistics</span>
          <br />
          <span className="text-foreground">for Strava</span>
          <span className="cursor-blink text-terminal-green">_</span>
        </h1>

        {/* Subtitle */}
        <p className="animate-fade-in animate-delay-2 text-lg md:text-xl text-muted-foreground max-w-2xl mx-auto mb-12 leading-relaxed">
          A self-hosted analytics dashboard for Strava athletes.
          <br className="hidden md:block" />
          Get comprehensive statistics, visualizations, and insights into your training data.
        </p>

        {/* CTA buttons */}
        <div className="animate-fade-in animate-delay-3 flex flex-col sm:flex-row items-center justify-center gap-4">
          <a
            href="https://github.com/melonamin/quantlete/releases"
            className="btn btn-primary group"
          >
            <Apple className="w-4 h-4" />
            <span>Download for macOS</span>
          </a>
          <a
            href="https://app.quantlete.fit"
            className="btn btn-outline group"
          >
            <Globe className="w-4 h-4" />
            <span>Open Web App</span>
          </a>
          <a
            href="https://github.com/melonamin/quantlete#installation"
            className="btn btn-outline group"
          >
            <Server className="w-4 h-4" />
            <span>Self-Host</span>
          </a>
        </div>

        {/* Terminal decoration */}
        <div className="animate-fade-in animate-delay-4 mt-16 text-xs text-terminal-dim font-mono">
          <span className="text-terminal-green">{'>'}</span> 100% open source
          <span className="mx-3 text-border">|</span>
          <span className="text-terminal-green">{'>'}</span> Your data stays with you
          <span className="mx-3 text-border">|</span>
          <span className="text-terminal-green">{'>'}</span> No cloud required
        </div>
      </div>

      {/* Scroll indicator */}
      <div className="absolute bottom-8 left-1/2 -translate-x-1/2 animate-fade-in animate-delay-5">
        <div className="w-6 h-10 border border-border rounded-full flex items-start justify-center p-2">
          <div className="w-1 h-2 bg-muted-foreground rounded-full animate-bounce" />
        </div>
      </div>
    </section>
  )
}
