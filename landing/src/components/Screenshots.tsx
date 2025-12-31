const screenshots = [
  {
    title: 'Dashboard Preview',
    description: 'Customizable widget-based dashboard',
  },
  {
    title: 'Activity Maps',
    description: 'Interactive route visualization',
  },
  {
    title: 'Training Analytics',
    description: 'Deep performance insights',
  },
]

function TerminalWindow({
  title,
  description,
}: {
  title: string
  description: string
}) {
  return (
    <div className="terminal-window group hover:border-terminal-green/30 transition-colors duration-300">
      {/* Terminal header */}
      <div className="terminal-header">
        <div className="terminal-dot terminal-dot-red" />
        <div className="terminal-dot terminal-dot-yellow" />
        <div className="terminal-dot terminal-dot-green" />
        <span className="ml-3 text-xs text-muted-foreground">{title}</span>
      </div>

      {/* Content placeholder */}
      <div className="aspect-video flex flex-col items-center justify-center p-8 bg-gradient-to-br from-card to-muted/20">
        {/* Placeholder content with terminal aesthetic */}
        <div className="w-full max-w-xs space-y-3 opacity-50">
          <div className="h-2 bg-terminal-green/20 rounded-sm w-3/4" />
          <div className="h-2 bg-terminal-green/15 rounded-sm w-full" />
          <div className="h-2 bg-terminal-green/10 rounded-sm w-5/6" />
          <div className="h-2 bg-terminal-green/15 rounded-sm w-2/3" />
        </div>

        <div className="mt-8 text-center">
          <p className="text-sm text-muted-foreground mb-1">Screenshot placeholder</p>
          <p className="text-xs text-terminal-dim">{description}</p>
        </div>
      </div>
    </div>
  )
}

export function Screenshots() {
  return (
    <section className="px-6 py-24 border-t border-border-dim">
      <div className="max-w-6xl mx-auto">
        {/* Section header */}
        <div className="text-center mb-16">
          <div className="inline-flex items-center gap-2 px-3 py-1.5 rounded-md bg-card border border-border-dim text-xs text-muted-foreground mb-6">
            <span className="text-terminal-green">{'>'}</span>
            <span>cat ./screenshots/*</span>
          </div>
          <h2 className="text-3xl md:text-4xl font-bold mb-4">
            See it in <span className="text-terminal-green">action</span>
          </h2>
          <p className="text-muted-foreground max-w-xl mx-auto">
            A glimpse of what awaits you inside Quantlete
          </p>
        </div>

        {/* Screenshots grid */}
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
          {screenshots.map((screenshot) => (
            <TerminalWindow
              key={screenshot.title}
              title={screenshot.title}
              description={screenshot.description}
            />
          ))}
        </div>
      </div>
    </section>
  )
}
