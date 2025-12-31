import { Apple, Globe, Server, ExternalLink } from 'lucide-react'

const installOptions = [
  {
    id: 'macos',
    icon: Apple,
    title: 'macOS Menu Bar',
    description: 'Native menu bar app with embedded server. One-click access to your dashboard.',
    features: ['Auto-starts with system', 'Menu bar controls', 'Local SQLite database'],
    cta: 'Download',
    ctaLink: 'https://github.com/melonamin/quantlete/releases',
    highlight: true,
  },
  {
    id: 'web',
    icon: Globe,
    title: 'Web App',
    description: 'Runs entirely in your browser. No server required. Your data stays local.',
    features: ['Zero installation', 'WASM-powered', 'OPFS storage'],
    cta: 'Open Web App',
    ctaLink: 'https://app.quantlete.fit',
    highlight: false,
  },
  {
    id: 'selfhost',
    icon: Server,
    title: 'Self-Hosted',
    description: 'Docker or standalone binary. Full control over your deployment.',
    features: ['Docker support', 'Single binary', 'Your infrastructure'],
    cta: 'Documentation',
    ctaLink: 'https://github.com/melonamin/quantlete#installation',
    highlight: false,
  },
]

export function Installation() {
  return (
    <section id="install" className="px-6 py-24 border-t border-border-dim">
      <div className="max-w-6xl mx-auto">
        {/* Section header */}
        <div className="text-center mb-16">
          <div className="inline-flex items-center gap-2 px-3 py-1.5 rounded-md bg-card border border-border-dim text-xs text-muted-foreground mb-6">
            <span className="text-terminal-green">install</span>
            <span className="text-terminal-dim">--method</span>
            <span className="text-terminal-cyan">[macos|web|docker]</span>
          </div>
          <h2 className="text-3xl md:text-4xl font-bold mb-4">
            Choose your <span className="text-terminal-green">deployment</span>
          </h2>
          <p className="text-muted-foreground max-w-xl mx-auto">
            Three ways to run Quantlete. Pick what works best for you.
          </p>
        </div>

        {/* Install options grid */}
        <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
          {installOptions.map((option) => (
            <div
              key={option.id}
              className={`
                relative flex flex-col p-6 rounded-md border transition-all duration-300
                ${
                  option.highlight
                    ? 'border-terminal-green/50 bg-terminal-green/5 hover:border-terminal-green hover:bg-terminal-green/10'
                    : 'border-border-dim bg-card/50 hover:border-border hover:bg-card'
                }
              `}
            >
              {/* Recommended badge */}
              {option.highlight && (
                <div className="absolute -top-3 left-1/2 -translate-x-1/2 px-3 py-1 rounded-full bg-terminal-green text-primary-foreground text-xs font-medium">
                  Recommended
                </div>
              )}

              {/* Icon */}
              <div
                className={`
                  w-12 h-12 rounded-md flex items-center justify-center mb-4
                  ${option.highlight ? 'bg-terminal-green/20 text-terminal-green' : 'bg-muted text-muted-foreground'}
                `}
              >
                <option.icon className="w-6 h-6" />
              </div>

              {/* Content */}
              <h3 className="text-xl font-semibold mb-2">{option.title}</h3>
              <p className="text-sm text-muted-foreground mb-4 flex-grow">
                {option.description}
              </p>

              {/* Features list */}
              <ul className="space-y-2 mb-6">
                {option.features.map((feature) => (
                  <li
                    key={feature}
                    className="flex items-center gap-2 text-xs text-muted-foreground"
                  >
                    <span className="text-terminal-green">{'>'}</span>
                    <span>{feature}</span>
                  </li>
                ))}
              </ul>

              {/* CTA */}
              <a
                href={option.ctaLink}
                className={`
                  btn w-full justify-center
                  ${option.highlight ? 'btn-primary' : 'btn-outline'}
                `}
              >
                <span>{option.cta}</span>
                <ExternalLink className="w-3.5 h-3.5" />
              </a>
            </div>
          ))}
        </div>

        {/* Requirements note */}
        <div className="mt-12 text-center text-xs text-terminal-dim">
          <span className="text-terminal-green">Note:</span> macOS app requires macOS 13.0+
          <span className="mx-2">·</span>
          Web app requires a modern browser with OPFS support
          <span className="mx-2">·</span>
          Self-hosted works on Linux, macOS, and Windows
        </div>
      </div>
    </section>
  )
}
