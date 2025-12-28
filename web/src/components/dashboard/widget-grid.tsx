import { useEffect, useMemo, useRef, useState, type ReactNode } from 'react'
import {
  useDashboardConfig,
  useUpdateDashboardConfig,
  type DashboardWidgetConfig,
  type WidgetWidth,
} from '@/lib/api/dashboard'
import { Button } from '@/components/ui/button'
import { useDashboardLayoutStore } from '@/stores/dashboard'
import { cn } from '@/lib/utils/cn'

export interface WidgetDefinition {
  id: string
  title: string
  defaultWidth: WidgetWidth
  defaultHidden?: boolean
  render: () => ReactNode
}

const WIDTH_OPTIONS: { label: string; value: WidgetWidth }[] = [
  { label: '33%', value: 4 },
  { label: '50%', value: 6 },
  { label: '66%', value: 8 },
  { label: '100%', value: 12 },
]

function normalizeConfig(
  cfg: { version: number; widgets: DashboardWidgetConfig[] },
  defs: WidgetDefinition[]
) {
  const byId = new Map(defs.map((d) => [d.id, d] as const))

  const existing = cfg.widgets.filter((w) => byId.has(w.id))
  const existingIds = new Set(existing.map((w) => w.id))
  const missing = defs
    .filter((d) => !existingIds.has(d.id))
    .map(
      (d) =>
        ({
          id: d.id,
          width: d.defaultWidth,
          hidden: !!d.defaultHidden,
        }) satisfies DashboardWidgetConfig
    )
  return { ...cfg, widgets: [...existing, ...missing] }
}

export function WidgetGrid({ widgets }: { widgets: WidgetDefinition[] }) {
  const { data: serverConfig, isLoading, error } = useDashboardConfig()
  const save = useUpdateDashboardConfig()
  const { config, setConfig, editMode, setEditMode, moveWidget, setWidgetHidden, setWidgetWidth } =
    useDashboardLayoutStore()

  const [showManage, setShowManage] = useState(false)
  const lastSavedRef = useRef<string>('')
  const hasLoadedRef = useRef(false)
  const saveTimer = useRef<number | null>(null)

  const normalizedServerConfig = useMemo(() => {
    if (!serverConfig) return null
    return normalizeConfig(serverConfig, widgets)
  }, [serverConfig, widgets])

  useEffect(() => {
    if (!normalizedServerConfig) return
    if (hasLoadedRef.current) return
    setConfig(normalizedServerConfig)
    hasLoadedRef.current = true
    lastSavedRef.current = JSON.stringify(normalizedServerConfig)
  }, [normalizedServerConfig, setConfig])

  useEffect(() => {
    if (!hasLoadedRef.current) return
    if (!config) return
    const next = JSON.stringify(config)
    if (next === lastSavedRef.current) return

    if (saveTimer.current) window.clearTimeout(saveTimer.current)
    saveTimer.current = window.setTimeout(() => {
      save.mutate(config, {
        onSuccess: (saved) => {
          lastSavedRef.current = JSON.stringify(saved)
          setConfig(saved)
        },
      })
    }, 650)

    return () => {
      if (saveTimer.current) window.clearTimeout(saveTimer.current)
    }
  }, [config, save, setConfig])

  const ordered = useMemo(() => {
    if (!config) return []
    const defsById = new Map(widgets.map((w) => [w.id, w] as const))
    return config.widgets
      .map((w) => ({ ...w, def: defsById.get(w.id) }))
      .filter((w) => w.def)
      .map((w) => ({
        id: w.id,
        width: w.width,
        hidden: w.hidden,
        title: (w.def as WidgetDefinition).title,
        render: (w.def as WidgetDefinition).render,
      }))
  }, [config, widgets])

  const visible = ordered.filter((w) => !w.hidden)

  const onDragStart = (e: React.DragEvent<HTMLDivElement>, id: string) => {
    if (!editMode) return
    e.dataTransfer.effectAllowed = 'move'
    e.dataTransfer.setData('text/plain', id)
  }

  const onDrop = (e: React.DragEvent<HTMLDivElement>, overId: string) => {
    if (!editMode) return
    e.preventDefault()
    const activeId = e.dataTransfer.getData('text/plain')
    if (activeId) moveWidget(activeId, overId)
  }

  if (error) {
    return (
      <div className="rounded-lg border border-destructive bg-destructive/10 p-4">
        <p className="text-sm text-destructive">Failed to load dashboard layout.</p>
      </div>
    )
  }

  return (
    <div>
      <div className="mb-4 flex items-center justify-between">
        <div className="flex items-center gap-2">
          <Button
            variant="outline"
            size="sm"
            onClick={() => setEditMode(!editMode)}
            disabled={isLoading}
          >
            {editMode ? 'Done' : 'Edit layout'}
          </Button>
          <Button
            variant="ghost"
            size="sm"
            onClick={() => setShowManage(true)}
            disabled={isLoading || !config}
          >
            Manage widgets
          </Button>
        </div>
        {save.isPending && <p className="text-xs text-muted-foreground">Saving…</p>}
      </div>

      <div className={cn('grid gap-6 md:grid-cols-12', editMode && 'select-none')}>
        {visible.map((w) => (
          <div
            key={w.id}
            className={cn(
              'relative md:col-span-12',
              w.width === 4 && 'md:col-span-4',
              w.width === 6 && 'md:col-span-6',
              w.width === 8 && 'md:col-span-8',
              w.width === 12 && 'md:col-span-12',
              editMode && 'cursor-move'
            )}
            draggable={editMode}
            onDragStart={(e) => onDragStart(e, w.id)}
            onDragOver={(e) => editMode && e.preventDefault()}
            onDrop={(e) => onDrop(e, w.id)}
          >
            {editMode && (
              <div className="absolute right-2 top-2 z-10 flex items-center gap-2 rounded-md bg-background/80 p-1 backdrop-blur">
                <select
                  className="h-8 rounded-md border border-border bg-background px-2 text-xs"
                  value={w.width}
                  onChange={(e) => setWidgetWidth(w.id, Number(e.target.value) as WidgetWidth)}
                >
                  {WIDTH_OPTIONS.map((o) => (
                    <option key={o.value} value={o.value}>
                      {o.label}
                    </option>
                  ))}
                </select>
                <Button variant="outline" size="sm" onClick={() => setWidgetHidden(w.id, true)}>
                  Hide
                </Button>
              </div>
            )}
            {w.render()}
          </div>
        ))}
      </div>

      {showManage && config && (
        <ManageWidgetsDialog
          onClose={() => setShowManage(false)}
          widgets={ordered}
          setWidgetHidden={setWidgetHidden}
        />
      )}
    </div>
  )
}

function ManageWidgetsDialog({
  onClose,
  widgets,
  setWidgetHidden,
}: {
  onClose: () => void
  widgets: { id: string; title: string; hidden: boolean }[]
  setWidgetHidden: (id: string, hidden: boolean) => void
}) {
  useEffect(() => {
    const onKey = (e: KeyboardEvent) => {
      if (e.key === 'Escape') onClose()
    }
    window.addEventListener('keydown', onKey)
    return () => window.removeEventListener('keydown', onKey)
  }, [onClose])

  return (
    <div className="fixed inset-0 z-50 bg-background/80 backdrop-blur-sm">
      <div className="mx-auto mt-20 w-full max-w-lg rounded-lg border border-border bg-card p-4 shadow-lg">
        <div className="mb-3 flex items-center justify-between">
          <h2 className="text-base font-semibold">Manage widgets</h2>
          <Button variant="ghost" size="sm" onClick={onClose}>
            Close
          </Button>
        </div>
        <div className="space-y-2">
          {widgets.map((w) => (
            <label
              key={w.id}
              className="flex items-center justify-between rounded-md border border-border px-3 py-2"
            >
              <span className="text-sm">{w.title}</span>
              <input
                type="checkbox"
                checked={!w.hidden}
                onChange={(e) => setWidgetHidden(w.id, !e.target.checked)}
              />
            </label>
          ))}
        </div>
      </div>
    </div>
  )
}
