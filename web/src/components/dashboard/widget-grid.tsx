import { useEffect, useMemo, useRef, useState, type ReactNode } from 'react'
import {
  DndContext,
  closestCenter,
  KeyboardSensor,
  PointerSensor,
  useSensor,
  useSensors,
  DragOverlay,
  type DragStartEvent,
  type DragEndEvent,
} from '@dnd-kit/core'
import {
  arrayMove,
  SortableContext,
  sortableKeyboardCoordinates,
  rectSortingStrategy,
} from '@dnd-kit/sortable'
import {
  useDashboardConfig,
  useUpdateDashboardConfig,
  type DashboardWidgetConfig,
  type WidgetWidth,
} from '@/lib/api/dashboard'
import { useDashboardLayoutStore } from '@/stores/dashboard'
import { SortableWidget } from './sortable-widget'
import { WidgetPanel } from './widget-panel'
import { cn } from '@/lib/utils/cn'
import { Settings2, X, GripVertical } from 'lucide-react'

export interface WidgetDefinition {
  id: string
  title: string
  defaultWidth: WidgetWidth
  defaultHidden?: boolean
  render: () => ReactNode
}

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
  const {
    config,
    setConfig,
    editMode,
    setEditMode,
    setWidgetHidden,
    setWidgetWidth,
    reorderWidgets,
  } = useDashboardLayoutStore()

  const [showPanel, setShowPanel] = useState(false)
  const [activeId, setActiveId] = useState<string | null>(null)
  const lastSavedRef = useRef<string>('')
  const hasLoadedRef = useRef(false)
  const saveTimer = useRef<number | null>(null)

  const sensors = useSensors(
    useSensor(PointerSensor, {
      activationConstraint: {
        distance: 8,
      },
    }),
    useSensor(KeyboardSensor, {
      coordinateGetter: sortableKeyboardCoordinates,
    })
  )

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

  // Keyboard shortcut to exit edit mode
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === 'Escape' && editMode) {
        setEditMode(false)
      }
    }
    window.addEventListener('keydown', handleKeyDown)
    return () => window.removeEventListener('keydown', handleKeyDown)
  }, [editMode, setEditMode])

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
  const visibleIds = visible.map((w) => w.id)

  const handleDragStart = (event: DragStartEvent) => {
    setActiveId(event.active.id as string)
  }

  const handleDragEnd = (event: DragEndEvent) => {
    const { active, over } = event
    setActiveId(null)

    if (over && active.id !== over.id) {
      const oldIndex = visibleIds.indexOf(active.id as string)
      const newIndex = visibleIds.indexOf(over.id as string)
      const newOrder = arrayMove(visibleIds, oldIndex, newIndex)
      reorderWidgets(newOrder)
    }
  }

  const activeWidget = activeId ? visible.find((w) => w.id === activeId) : null

  if (error) {
    return (
      <div className="rounded-sm border border-destructive bg-destructive/10 p-4">
        <p className="text-sm text-destructive">Failed to load dashboard layout.</p>
      </div>
    )
  }

  return (
    <div className="relative">
      {/* Edit Mode Status Bar */}
      {editMode && (
        <div className="mb-4 flex items-center justify-between rounded-sm border border-terminal-green/30 bg-terminal-green/5 px-4 py-2">
          <div className="flex items-center gap-3">
            <div className="flex items-center gap-2">
              <div className="h-2 w-2 animate-pulse rounded-full bg-terminal-green" />
              <span className="text-sm font-medium text-terminal-green">EDIT MODE</span>
            </div>
            <span className="text-xs text-muted-foreground">
              Drag widgets to reorder • Click width buttons to resize • Press ESC to exit
            </span>
          </div>
          <div className="flex items-center gap-2">
            {save.isPending && (
              <span className="text-xs text-terminal-amber animate-pulse">SAVING...</span>
            )}
            <button
              onClick={() => setShowPanel(true)}
              className="flex items-center gap-1.5 rounded-sm border border-border bg-card px-2 py-1 text-xs transition-colors hover:border-terminal-green hover:text-terminal-green"
            >
              <Settings2 className="h-3 w-3" />
              WIDGETS
            </button>
            <button
              onClick={() => setEditMode(false)}
              className="flex items-center gap-1.5 rounded-sm border border-border bg-card px-2 py-1 text-xs transition-colors hover:border-destructive hover:text-destructive"
            >
              <X className="h-3 w-3" />
              EXIT
            </button>
          </div>
        </div>
      )}

      {/* Normal Mode Controls */}
      {!editMode && (
        <div className="mb-4 flex items-center justify-end">
          <button
            onClick={() => setEditMode(true)}
            disabled={isLoading}
            className="flex items-center gap-1.5 rounded-sm border border-border bg-card px-3 py-1.5 text-xs transition-all hover:border-terminal-green hover:text-terminal-green disabled:opacity-50"
          >
            <Settings2 className="h-3.5 w-3.5" />
            Customize Dashboard
          </button>
        </div>
      )}

      {/* Widget Grid */}
      <DndContext
        sensors={sensors}
        collisionDetection={closestCenter}
        onDragStart={handleDragStart}
        onDragEnd={handleDragEnd}
      >
        <SortableContext items={visibleIds} strategy={rectSortingStrategy}>
          <div
            className={cn(
              'grid gap-4 md:grid-cols-12',
              editMode && 'rounded-sm border border-dashed border-border/50 p-4'
            )}
          >
            {visible.map((w) => (
              <SortableWidget
                key={w.id}
                id={w.id}
                title={w.title}
                width={w.width}
                editMode={editMode}
                onWidthChange={(width) => setWidgetWidth(w.id, width)}
                onHide={() => setWidgetHidden(w.id, true)}
              >
                {w.render()}
              </SortableWidget>
            ))}
          </div>
        </SortableContext>

        {/* Drag Overlay - shows dragged widget preview */}
        <DragOverlay>
          {activeWidget ? (
            <div
              className={cn(
                'rounded-sm border-2 border-terminal-green bg-card/95 shadow-2xl shadow-terminal-green/20',
                'opacity-90 backdrop-blur',
                activeWidget.width === 4 && 'w-[300px]',
                activeWidget.width === 6 && 'w-[400px]',
                activeWidget.width === 8 && 'w-[500px]',
                activeWidget.width === 12 && 'w-[600px]'
              )}
            >
              <div className="flex items-center gap-2 border-b border-border px-3 py-2">
                <GripVertical className="h-4 w-4 text-terminal-green" />
                <span className="text-sm font-medium">{activeWidget.title}</span>
              </div>
              <div className="p-4 opacity-50">
                <div className="h-24 rounded-sm bg-muted/50" />
              </div>
            </div>
          ) : null}
        </DragOverlay>
      </DndContext>

      {/* Widget Management Panel */}
      <WidgetPanel
        open={showPanel}
        onClose={() => setShowPanel(false)}
        widgets={ordered}
        onToggleWidget={(id, hidden) => setWidgetHidden(id, hidden)}
        onReorder={(newOrder) => reorderWidgets(newOrder)}
      />
    </div>
  )
}
