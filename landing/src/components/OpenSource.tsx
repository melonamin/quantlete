import { Github, Heart, Scale } from 'lucide-react'

export function OpenSource() {
  return (
    <section className="px-6 py-24 border-t border-border-dim">
      <div className="max-w-4xl mx-auto">
        {/* Terminal-style open source callout */}
        <div className="terminal-window">
          <div className="terminal-header">
            <div className="terminal-dot terminal-dot-red" />
            <div className="terminal-dot terminal-dot-yellow" />
            <div className="terminal-dot terminal-dot-green" />
            <span className="ml-3 text-xs text-muted-foreground">open-source.md</span>
          </div>

          <div className="p-8 md:p-12 text-center">
            {/* Icon */}
            <div className="inline-flex items-center justify-center w-16 h-16 rounded-full bg-terminal-green/10 text-terminal-green mb-6">
              <Heart className="w-8 h-8" />
            </div>

            {/* Heading */}
            <h2 className="text-3xl md:text-4xl font-bold mb-4">
              100% <span className="text-terminal-green">Open Source</span>
            </h2>

            {/* Description */}
            <p className="text-muted-foreground max-w-xl mx-auto mb-8 leading-relaxed">
              Quantlete is free and open source software. Your data belongs to you.
              Inspect the code, contribute features, or self-host with full confidence.
            </p>

            {/* License badge */}
            <div className="flex items-center justify-center gap-4 mb-8">
              <div className="inline-flex items-center gap-2 px-3 py-1.5 rounded-md bg-muted border border-border-dim text-sm">
                <Scale className="w-4 h-4 text-terminal-green" />
                <span className="text-muted-foreground">MIT License</span>
              </div>
            </div>

            {/* GitHub CTA */}
            <a
              href="https://github.com/melonamin/quantlete"
              className="btn btn-primary inline-flex"
            >
              <Github className="w-4 h-4" />
              <span>View on GitHub</span>
            </a>

            {/* Terminal decoration */}
            <div className="mt-8 pt-6 border-t border-border-dim">
              <pre className="text-xs text-terminal-dim text-left inline-block">
                <code>
                  <span className="text-terminal-green">$</span> git clone https://github.com/melonamin/quantlete.git{'\n'}
                  <span className="text-terminal-green">$</span> cd quantlete{'\n'}
                  <span className="text-terminal-green">$</span> just dev
                </code>
              </pre>
            </div>
          </div>
        </div>

        {/* Stats row */}
        <div className="mt-8 grid grid-cols-3 gap-4 text-center">
          <div className="p-4 rounded-md border border-border-dim bg-card/30">
            <div className="text-2xl font-bold text-terminal-green mb-1">Go</div>
            <div className="text-xs text-muted-foreground">Backend</div>
          </div>
          <div className="p-4 rounded-md border border-border-dim bg-card/30">
            <div className="text-2xl font-bold text-terminal-cyan mb-1">React</div>
            <div className="text-xs text-muted-foreground">Frontend</div>
          </div>
          <div className="p-4 rounded-md border border-border-dim bg-card/30">
            <div className="text-2xl font-bold text-terminal-amber mb-1">SQLite</div>
            <div className="text-xs text-muted-foreground">Database</div>
          </div>
        </div>
      </div>
    </section>
  )
}
