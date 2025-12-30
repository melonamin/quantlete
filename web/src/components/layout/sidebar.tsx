import { Link, useRouterState } from '@tanstack/react-router'
import { cn } from '@/lib/utils'
import { useSidebarStore, useSyncModalStore } from '@/stores'
import { useAppSettings, useAuthStatus, useImportProgress } from '@/lib/api'
import { isWasmMode } from '@/lib/mode'
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
  ChevronLeft,
  ChevronRight,
  Menu,
  X,
  BarChart3,
  Download,
  Loader2,
  Clock,
  Tag,
  User,
  HeartPulse,
  Zap,
} from 'lucide-react'
import { useEffect, useCallback } from 'react'

export const SIDEBAR_COLLAPSED_WIDTH = 56
export const SIDEBAR_EXPANDED_WIDTH = 220

interface NavItem {
  name: string
  href: string
  icon: React.ComponentType<{ className?: string }>
  shortcut?: string
}

export function Sidebar() {
  const routerState = useRouterState()
  const currentPath = routerState.location.pathname
  const { data: auth } = useAuthStatus()
  const { data: settings } = useAppSettings({ enabled: !!auth?.authenticated })
  const { data: importProgress } = useImportProgress()
  const { collapsed, toggleCollapsed, mobileOpen, setMobileOpen } = useSidebarStore()
  const openSyncModal = useSyncModalStore((s) => s.openModal)

  const isSyncing = importProgress?.status === 'running'
  const isWaitingForRateLimit = importProgress?.waiting_for_rate_limit

  const showEddington =
    settings?.eddington_definitions?.some((d) => d.show_in_nav !== false) ?? true

  const navigation: NavItem[] = [
    { name: 'Dashboard', href: '/', icon: LayoutDashboard, shortcut: 'g d' },
    { name: 'Activities', href: '/activities', icon: Activity, shortcut: 'g a' },
    { name: 'Heatmap', href: '/heatmap', icon: Map, shortcut: 'g h' },
    { name: 'Calendar', href: '/calendar', icon: Calendar, shortcut: 'g c' },
    { name: 'Monthly Stats', href: '/monthly-stats', icon: BarChart3 },
    { name: 'Segments', href: '/segments', icon: Trophy, shortcut: 'g s' },
    { name: 'Gear', href: '/gear', icon: Bike, shortcut: 'g g' },
    { name: 'Photos', href: '/photos', icon: Camera },
    { name: 'Challenges', href: '/challenges', icon: Award },
    ...(showEddington ? [{ name: 'Eddington', href: '/eddington', icon: TrendingUp }] : []),
    { name: 'Best Efforts', href: '/best-efforts', icon: Timer },
    { name: 'Training Load', href: '/training-load', icon: HeartPulse },
    { name: 'Power', href: '/power', icon: Zap },
    { name: 'Rewind', href: '/rewind', icon: History },
    { name: 'Badges', href: '/badges', icon: Tag },
    { name: 'Export', href: '/export', icon: Download },
  ]

  const handleKeyDown = useCallback(
    (e: KeyboardEvent) => {
      if (e.key === '[' || e.key === ']') {
        e.preventDefault()
        toggleCollapsed()
      }
      if ((e.metaKey || e.ctrlKey) && e.key === 'b') {
        e.preventDefault()
        toggleCollapsed()
      }
      if (e.key === 'Escape' && mobileOpen) {
        setMobileOpen(false)
      }
    },
    [toggleCollapsed, mobileOpen, setMobileOpen]
  )

  useEffect(() => {
    document.addEventListener('keydown', handleKeyDown)
    return () => document.removeEventListener('keydown', handleKeyDown)
  }, [handleKeyDown])

  useEffect(() => {
    setMobileOpen(false)
  }, [currentPath, setMobileOpen])

  const isActive = (href: string) =>
    currentPath === href || (href !== '/' && currentPath.startsWith(href))

  return (
    <>
      {/* Mobile overlay */}
      {mobileOpen && (
        <div
          className="fixed inset-0 z-40 bg-background/80 backdrop-blur-sm md:hidden"
          onClick={() => setMobileOpen(false)}
        />
      )}

      {/* Sidebar */}
      <aside
        className={cn(
          'fixed left-0 top-0 z-50 flex h-screen flex-col border-r border-sidebar-border bg-sidebar transition-all duration-200 ease-in-out',
          collapsed ? 'w-[--sidebar-width-collapsed]' : 'w-[--sidebar-width-expanded]',
          'max-md:-translate-x-full max-md:data-[mobile-open=true]:translate-x-0',
          'max-md:w-[--sidebar-width-expanded]'
        )}
        data-mobile-open={mobileOpen}
        role="navigation"
        aria-label="Main navigation"
      >
        {/* Logo/Brand */}
        <div className="flex h-14 items-center border-b border-sidebar-border px-3">
          <Link to="/" className="flex items-center gap-2 overflow-hidden">
            <div className="flex h-8 w-8 shrink-0 items-center justify-center rounded-sm bg-strava font-bold text-white text-sm">
              S
            </div>
            <span
              className={cn(
                'font-semibold text-lg whitespace-nowrap transition-opacity duration-200',
                collapsed ? 'opacity-0 w-0' : 'opacity-100'
              )}
            >
              Quantlete
            </span>
          </Link>
        </div>

        {/* Navigation */}
        <nav className="flex-1 overflow-y-auto py-2">
          <ul className="space-y-0.5 px-2">
            {navigation.map((item) => (
              <li key={item.name}>
                <Link
                  to={item.href}
                  className={cn(
                    'group flex items-center gap-3 rounded-sm px-2 py-2 text-sm transition-colors',
                    isActive(item.href)
                      ? 'bg-sidebar-accent text-terminal-green'
                      : 'text-sidebar-foreground/70 hover:bg-sidebar-accent/50 hover:text-sidebar-foreground'
                  )}
                  title={collapsed ? item.name : undefined}
                >
                  <item.icon
                    className={cn(
                      'h-4 w-4 shrink-0',
                      isActive(item.href) ? 'text-terminal-green' : ''
                    )}
                  />
                  <span
                    className={cn(
                      'truncate transition-opacity duration-200',
                      collapsed ? 'opacity-0 w-0 overflow-hidden' : 'opacity-100'
                    )}
                  >
                    {item.name}
                  </span>
                  {!collapsed && item.shortcut && (
                    <span className="ml-auto text-[10px] text-sidebar-foreground/30">
                      {item.shortcut}
                    </span>
                  )}
                </Link>
              </li>
            ))}
          </ul>
        </nav>

        {/* Bottom section */}
        <div className="border-t border-sidebar-border px-2 py-2">
          {/* Sync indicator */}
          {isSyncing && (
            <Link
              to="/settings"
              onClick={(e) => {
                // In WASM mode, open the sync modal instead of navigating
                if (isWasmMode()) {
                  e.preventDefault()
                  openSyncModal()
                }
              }}
              className={cn(
                'flex items-center gap-3 rounded-sm px-2 py-2 text-sm transition-colors',
                isWaitingForRateLimit
                  ? 'text-amber-600 dark:text-amber-500 hover:bg-amber-100/50 dark:hover:bg-amber-950/30'
                  : 'text-primary hover:bg-sidebar-accent/50'
              )}
              title={collapsed ? (isWaitingForRateLimit ? 'Waiting for rate limit' : 'Sync in progress') : undefined}
            >
              {isWaitingForRateLimit ? (
                <Clock className="h-4 w-4 shrink-0" />
              ) : (
                <Loader2 className="h-4 w-4 shrink-0 animate-spin" />
              )}
              <span
                className={cn(
                  'truncate transition-opacity duration-200',
                  collapsed ? 'opacity-0 w-0 overflow-hidden' : 'opacity-100'
                )}
              >
                {isWaitingForRateLimit ? 'Waiting...' : 'Syncing...'}
              </span>
            </Link>
          )}

          <Link
            to="/athlete"
            className={cn(
              'group flex items-center gap-3 rounded-sm px-2 py-2 text-sm transition-colors',
              currentPath === '/athlete'
                ? 'bg-sidebar-accent text-terminal-green'
                : 'text-sidebar-foreground/70 hover:bg-sidebar-accent/50 hover:text-sidebar-foreground'
            )}
            title={collapsed ? 'Athlete' : undefined}
          >
            <User className="h-4 w-4 shrink-0" />
            <span
              className={cn(
                'truncate transition-opacity duration-200',
                collapsed ? 'opacity-0 w-0 overflow-hidden' : 'opacity-100'
              )}
            >
              Athlete
            </span>
          </Link>

          <Link
            to="/settings"
            className={cn(
              'group flex items-center gap-3 rounded-sm px-2 py-2 text-sm transition-colors',
              currentPath === '/settings'
                ? 'bg-sidebar-accent text-terminal-green'
                : 'text-sidebar-foreground/70 hover:bg-sidebar-accent/50 hover:text-sidebar-foreground'
            )}
            title={collapsed ? 'Settings' : undefined}
          >
            <Settings className="h-4 w-4 shrink-0" />
            <span
              className={cn(
                'truncate transition-opacity duration-200',
                collapsed ? 'opacity-0 w-0 overflow-hidden' : 'opacity-100'
              )}
            >
              Settings
            </span>
          </Link>

          {/* Collapse toggle - desktop only */}
          <button
            onClick={toggleCollapsed}
            className="hidden md:flex mt-1 w-full items-center gap-3 rounded-sm px-2 py-2 text-sm text-sidebar-foreground/50 transition-colors hover:bg-sidebar-accent/50 hover:text-sidebar-foreground"
            title={collapsed ? 'Expand sidebar ([ or ])' : 'Collapse sidebar ([ or ])'}
          >
            {collapsed ? (
              <ChevronRight className="h-4 w-4 shrink-0" />
            ) : (
              <>
                <ChevronLeft className="h-4 w-4 shrink-0" />
                <span className="text-xs">Collapse</span>
                <span className="ml-auto text-[10px]">[ ]</span>
              </>
            )}
          </button>
        </div>
      </aside>

      {/* Mobile menu button */}
      <button
        onClick={() => setMobileOpen(!mobileOpen)}
        className="fixed left-3 top-3 z-50 flex h-10 w-10 items-center justify-center rounded-sm border border-border bg-card md:hidden"
        aria-label={mobileOpen ? 'Close menu' : 'Open menu'}
      >
        {mobileOpen ? <X className="h-5 w-5" /> : <Menu className="h-5 w-5" />}
      </button>
    </>
  )
}
