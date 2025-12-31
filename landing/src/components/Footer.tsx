import { Github } from 'lucide-react'

export function Footer() {
  const currentYear = new Date().getFullYear()

  return (
    <footer className="px-6 py-12 border-t border-border-dim">
      <div className="max-w-6xl mx-auto">
        <div className="flex flex-col md:flex-row items-center justify-between gap-6">
          {/* Logo / Brand */}
          <div className="flex items-center gap-3">
            <div className="w-8 h-8 rounded-md bg-card border border-border flex items-center justify-center">
              <span className="text-terminal-green font-bold text-sm">Q</span>
            </div>
            <span className="text-sm text-muted-foreground">
              Quantlete
            </span>
          </div>

          {/* Links */}
          <div className="flex items-center gap-6 text-sm text-muted-foreground">
            <a
              href="https://github.com/melonamin/quantlete"
              className="hover:text-foreground transition-colors inline-flex items-center gap-2"
            >
              <Github className="w-4 h-4" />
              <span>GitHub</span>
            </a>
            <a
              href="https://github.com/melonamin/quantlete/releases"
              className="hover:text-foreground transition-colors"
            >
              Releases
            </a>
            <a
              href="https://github.com/melonamin/quantlete#installation"
              className="hover:text-foreground transition-colors"
            >
              Docs
            </a>
          </div>

          {/* Copyright */}
          <div className="text-xs text-terminal-dim">
            <span>© {currentYear} Quantlete</span>
            <span className="mx-2">·</span>
            <span>MIT License</span>
          </div>
        </div>

        {/* Terminal decoration */}
        <div className="mt-8 text-center">
          <div className="inline-flex items-center gap-2 text-xs text-terminal-dim">
            <span className="text-terminal-green">{'>'}</span>
            <span>Built with</span>
            <span className="text-terminal-green">{'<3'}</span>
            <span>for athletes who love data</span>
          </div>
        </div>
      </div>
    </footer>
  )
}
