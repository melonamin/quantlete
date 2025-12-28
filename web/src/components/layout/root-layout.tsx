import type { ReactNode } from 'react'
import { Link, useRouterState } from '@tanstack/react-router'
import { cn } from '@/lib/utils'
import { useAppSettings, useAuthStatus } from '@/lib/api'
import {
  LayoutDashboard,
  Activity,
  Map,
  Calendar,
  Trophy,
  Bike,
  TrendingUp,
  Timer,
  Camera,
  Award,
  Settings,
  History,
} from 'lucide-react'

interface RootLayoutProps {
  children: ReactNode
}

export function RootLayout({ children }: RootLayoutProps) {
  const routerState = useRouterState()
  const currentPath = routerState.location.pathname
  const { data: auth } = useAuthStatus()
  const isAuthenticated = auth?.authenticated
  const { data: settings } = useAppSettings({ enabled: !!isAuthenticated })

  const showEddington =
    settings?.eddington_definitions && settings.eddington_definitions.length
      ? settings.eddington_definitions.some((d) => d.show_in_nav !== false)
      : true

  const navigation = [
    { name: 'Dashboard', href: '/', icon: LayoutDashboard },
    { name: 'Activities', href: '/activities', icon: Activity },
    { name: 'Heatmap', href: '/heatmap', icon: Map },
    { name: 'Calendar', href: '/calendar', icon: Calendar },
    { name: 'Segments', href: '/segments', icon: Trophy },
    { name: 'Gear', href: '/gear', icon: Bike },
    { name: 'Photos', href: '/photos', icon: Camera },
    { name: 'Challenges', href: '/challenges', icon: Award },
    ...(showEddington ? [{ name: 'Eddington', href: '/eddington', icon: TrendingUp }] : []),
    { name: 'Best Efforts', href: '/best-efforts', icon: Timer },
    { name: 'Rewind', href: '/rewind', icon: History },
  ]

  return (
    <div className="min-h-screen bg-background">
      {/* Header */}
      <header className="sticky top-0 z-50 border-b border-border bg-background/95 backdrop-blur supports-[backdrop-filter]:bg-background/60">
        <div className="flex h-14 items-center px-4 lg:px-6">
          <Link to="/" className="flex items-center gap-2">
            <div className="h-8 w-8 rounded-lg bg-strava flex items-center justify-center">
              <span className="text-white font-bold text-sm">S</span>
            </div>
            <span className="font-semibold text-lg">Stata</span>
          </Link>

          {/* Desktop navigation */}
          <nav className="ml-8 hidden md:flex items-center gap-1">
            {navigation.map((item) => {
              const isActive =
                currentPath === item.href ||
                (item.href !== '/' && currentPath.startsWith(item.href))
              return (
                <Link
                  key={item.name}
                  to={item.href}
                  className={cn(
                    'flex items-center gap-2 px-3 py-2 text-sm font-medium rounded-md transition-colors',
                    isActive
                      ? 'bg-accent text-accent-foreground'
                      : 'text-muted-foreground hover:text-foreground hover:bg-accent/50'
                  )}
                >
                  <item.icon className="h-4 w-4" />
                  {item.name}
                </Link>
              )
            })}
          </nav>

          <div className="ml-auto flex items-center gap-2">
            <Link
              to="/settings"
              className={cn(
                'flex items-center gap-2 px-3 py-2 text-sm font-medium rounded-md transition-colors',
                currentPath === '/settings'
                  ? 'bg-accent text-accent-foreground'
                  : 'text-muted-foreground hover:text-foreground hover:bg-accent/50'
              )}
            >
              <Settings className="h-4 w-4" />
              <span className="hidden sm:inline">Settings</span>
            </Link>
          </div>
        </div>
      </header>

      {/* Main content */}
      <main className="flex-1">{children}</main>
    </div>
  )
}
