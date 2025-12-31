import {
  Activity,
  Map,
  BarChart3,
  Layers,
  Wrench,
  Trophy,
  Zap,
  Calendar,
  Globe,
  Timer,
} from 'lucide-react'

const features = [
  {
    icon: Activity,
    title: 'Import from Strava',
    description: 'Seamlessly sync your activities via the Strava API',
  },
  {
    icon: Layers,
    title: 'Customizable Dashboard',
    description: 'Arrange widgets to focus on your key metrics',
  },
  {
    icon: Map,
    title: 'Activity Heatmaps',
    description: 'Visualize all your routes on an interactive map',
  },
  {
    icon: BarChart3,
    title: 'Detailed Charts',
    description: 'Analyze trends with powerful visualizations',
  },
  {
    icon: Trophy,
    title: 'Segment Analysis',
    description: 'Track your performance on favorite segments',
  },
  {
    icon: Wrench,
    title: 'Gear Tracking',
    description: 'Monitor usage and maintenance schedules',
  },
  {
    icon: Zap,
    title: 'Eddington Numbers',
    description: 'Track your cycling consistency metric',
  },
  {
    icon: Timer,
    title: 'Best Efforts',
    description: 'Personal records across all distances',
  },
  {
    icon: Calendar,
    title: 'Year in Review',
    description: 'Beautiful annual summaries of your training',
  },
  {
    icon: Globe,
    title: 'Browser-Only Mode',
    description: 'WASM-powered, no server required',
  },
]

export function Features() {
  return (
    <section className="px-6 py-24 border-t border-border-dim">
      <div className="max-w-6xl mx-auto">
        {/* Section header */}
        <div className="text-center mb-16">
          <div className="inline-flex items-center gap-2 px-3 py-1.5 rounded-md bg-card border border-border-dim text-xs text-muted-foreground mb-6">
            <span className="text-terminal-green">def</span>
            <span className="text-terminal-cyan">features</span>
            <span className="text-terminal-dim">()</span>
          </div>
          <h2 className="text-3xl md:text-4xl font-bold mb-4">
            Everything you need to
            <span className="text-terminal-green"> analyze</span> your training
          </h2>
          <p className="text-muted-foreground max-w-xl mx-auto">
            Comprehensive analytics tools built for athletes who take their training seriously
          </p>
        </div>

        {/* Features grid */}
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
          {features.map((feature, index) => (
            <div
              key={feature.title}
              className="group p-5 rounded-md border border-border-dim bg-card/50 hover:bg-card hover:border-border transition-all duration-300"
              style={{ animationDelay: `${index * 0.05}s` }}
            >
              <div className="flex items-start gap-4">
                <div className="p-2 rounded-md bg-muted/50 text-terminal-green group-hover:bg-terminal-green/10 transition-colors">
                  <feature.icon className="w-5 h-5" />
                </div>
                <div>
                  <h3 className="font-medium text-foreground mb-1">
                    {feature.title}
                  </h3>
                  <p className="text-sm text-muted-foreground">
                    {feature.description}
                  </p>
                </div>
              </div>
            </div>
          ))}
        </div>
      </div>
    </section>
  )
}
