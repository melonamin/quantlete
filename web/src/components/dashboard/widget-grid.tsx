import { useEffect, useMemo, useRef, useState, useCallback, type ReactNode } from 'react'
import {
  DndContext,
  rectIntersection,
  KeyboardSensor,
  PointerSensor,
  useSensor,
  useSensors,
  DragOverlay,
  useDroppable,
  type DragStartEvent,
  type DragEndEvent,
  type DragMoveEvent,
  MeasuringStrategy,
} from '@dnd-kit/core'
import { arrayMove, SortableContext, sortableKeyboardCoordinates } from '@dnd-kit/sortable'
import {
  useDashboardConfig,
  useUpdateDashboardConfig,
  type DashboardWidgetConfig,
  type WidgetWidth,
  type WidgetHeight,
} from '@/lib/api'
import { useDashboardLayoutStore } from '@/stores/dashboard'
import { SortableWidget } from './sortable-widget'
import { WidgetPanel } from './widget-panel'
import { cn } from '@/lib/utils/cn'
import { Settings2, X, GripVertical } from 'lucide-react'

// Droppable container for the entire grid
function DroppableGrid({ editMode, children }: { editMode: boolean; children: ReactNode }) {
  const { setNodeRef, isOver } = useDroppable({
    id: 'widget-grid-container',
  })

  return (
    <div
      ref={setNodeRef}
      className={cn(
        // 12-column grid with row-based heights and dense packing
        'grid gap-4 md:grid-cols-12 grid-flow-row-dense',
        // Base row height for spanning (180px per row unit)
        'auto-rows-[180px]',
        editMode && 'rounded-sm border border-dashed border-border/50 p-4 min-h-[200px]',
        editMode && isOver && 'border-terminal-green/50 bg-terminal-green/5'
      )}
    >
      {children}
    </div>
  )
}

export interface WidgetDefinition {
  id: string
  title: string
  defaultWidth: WidgetWidth
  defaultHeight?: WidgetHeight
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
          height: d.defaultHeight,
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
    setWidgetHeight,
    reorderWidgets,
  } = useDashboardLayoutStore()

  const [showPanel, setShowPanel] = useState(false)
  const [activeId, setActiveId] = useState<string | null>(null)
  const [dropIndex, setDropIndex] = useState<number | null>(null)
  const lastSavedRef = useRef<string>('')
  const hasLoadedRef = useRef(false)
  const saveTimer = useRef<number | null>(null)
  const widgetRefs = useRef<Map<string, HTMLElement>>(new Map())
  const pointerPositionRef = useRef<{ x: number; y: number } | null>(null)

  const sensors = useSensors(
    useSensor(PointerSensor, {
      activationConstraint: {
        distance: 10,
      },
    }),
    useSensor(KeyboardSensor, {
      coordinateGetter: sortableKeyboardCoordinates,
    })
  )

  // Measuring configuration to prevent layout thrashing during drag
  const measuringConfig = {
    droppable: {
      strategy: MeasuringStrategy.Always,
    },
  }

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
    const configSnapshot = JSON.stringify(config)
    if (configSnapshot === lastSavedRef.current) return

    if (saveTimer.current) window.clearTimeout(saveTimer.current)
    saveTimer.current = window.setTimeout(() => {
      // Capture the snapshot that was used to trigger this save
      const snapshotToSave = configSnapshot
      save.mutate(config, {
        onSuccess: () => {
          // Only update lastSavedRef if we're saving the same snapshot
          // This prevents race conditions where newer changes arrive during save
          if (lastSavedRef.current !== snapshotToSave) {
            // A newer change has been made, don't update the ref
            // The newer change will trigger its own save
            return
          }
          lastSavedRef.current = snapshotToSave
        },
      })
    }, 650)

    return () => {
      if (saveTimer.current) window.clearTimeout(saveTimer.current)
    }
  }, [config, save])

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
        height: w.height ?? (w.def as WidgetDefinition).defaultHeight ?? 2,
        hidden: w.hidden,
        title: (w.def as WidgetDefinition).title,
        render: (w.def as WidgetDefinition).render,
      }))
  }, [config, widgets])

  const visible = useMemo(() => ordered.filter((w) => !w.hidden), [ordered])
  const visibleIds = useMemo(() => visible.map((w) => w.id), [visible])

  // Calculate drop position based on pointer coordinates
  const calculateDropIndex = useCallback(
    (pointerX: number, pointerY: number, activeId: string): number | null => {
      const refs = widgetRefs.current
      const positions: { id: string; rect: DOMRect; originalIndex: number }[] = []

      visibleIds.forEach((id, index) => {
        const el = refs.get(id)
        if (el) {
          positions.push({ id, rect: el.getBoundingClientRect(), originalIndex: index })
        }
      })

      if (positions.length === 0) return null

      const activeIndex = visibleIds.indexOf(activeId)

      // Sort positions by their vertical then horizontal position (reading order)
      positions.sort((a, b) => {
        const rowA = Math.floor(a.rect.top / 50) // Group by approximate rows
        const rowB = Math.floor(b.rect.top / 50)
        if (rowA !== rowB) return rowA - rowB
        return a.rect.left - b.rect.left
      })

      // Find where the pointer falls in the sorted order
      let insertBeforeIndex = positions.length // Default: insert at end

      for (let i = 0; i < positions.length; i++) {
        const pos = positions[i]
        const rect = pos.rect

        // Check if pointer is above this widget's vertical center
        // or in the same row but before its horizontal center
        const verticalCenter = rect.top + rect.height / 2
        const horizontalCenter = rect.left + rect.width / 2

        if (pointerY < verticalCenter - rect.height * 0.3) {
          // Pointer is clearly above this row
          insertBeforeIndex = pos.originalIndex
          break
        } else if (
          pointerY >= rect.top - 20 &&
          pointerY <= rect.bottom + 20 &&
          pointerX < horizontalCenter
        ) {
          // Same row, pointer is to the left
          insertBeforeIndex = pos.originalIndex
          break
        }
      }

      // Convert "insert before" index to final target index
      let targetIndex = insertBeforeIndex

      // If we're inserting after the active item's original position,
      // we need to account for it being removed
      if (targetIndex > activeIndex) {
        targetIndex--
      }

      // Clamp to valid range
      return Math.max(0, Math.min(targetIndex, visibleIds.length - 1))
    },
    [visibleIds]
  )

  const handleDragStart = (event: DragStartEvent) => {
    setActiveId(event.active.id as string)
    setDropIndex(null)
    pointerPositionRef.current = null
  }

  const handleDragMove = (event: DragMoveEvent) => {
    if (!event.active) return

    // Track pointer position from the delta
    const initialRect = event.active.rect.current.initial
    const delta = event.delta

    if (initialRect) {
      const pointerX = initialRect.left + initialRect.width / 2 + delta.x
      const pointerY = initialRect.top + initialRect.height / 2 + delta.y
      pointerPositionRef.current = { x: pointerX, y: pointerY }

      const newDropIndex = calculateDropIndex(pointerX, pointerY, event.active.id as string)
      setDropIndex(newDropIndex)
    }
  }

  const handleDragEnd = (event: DragEndEvent) => {
    const { active } = event
    let finalDropIndex = dropIndex

    // If dropIndex wasn't set via onDragMove, calculate from final position
    if (finalDropIndex === null && pointerPositionRef.current) {
      finalDropIndex = calculateDropIndex(
        pointerPositionRef.current.x,
        pointerPositionRef.current.y,
        active.id as string
      )
    }

    setActiveId(null)
    setDropIndex(null)
    pointerPositionRef.current = null

    if (finalDropIndex === null) return

    const oldIndex = visibleIds.indexOf(active.id as string)
    if (oldIndex === -1) return
    if (oldIndex === finalDropIndex) return

    const newOrder = arrayMove(visibleIds, oldIndex, finalDropIndex)
    reorderWidgets(newOrder)
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
        collisionDetection={rectIntersection}
        onDragStart={handleDragStart}
        onDragMove={handleDragMove}
        onDragEnd={handleDragEnd}
        measuring={measuringConfig}
      >
        <SortableContext items={visibleIds}>
          <DroppableGrid editMode={editMode}>
            {visible.map((w, index) => {
              // Show drop indicator before this widget
              const showIndicatorBefore =
                activeId !== null && dropIndex === index && visibleIds.indexOf(activeId) !== index

              return (
                <SortableWidget
                  key={w.id}
                  id={w.id}
                  title={w.title}
                  width={w.width}
                  height={w.height}
                  editMode={editMode}
                  onWidthChange={(width) => setWidgetWidth(w.id, width)}
                  onHeightChange={(height) => setWidgetHeight(w.id, height)}
                  onHide={() => setWidgetHidden(w.id, true)}
                  showDropIndicator={showIndicatorBefore}
                  onRefChange={(el) => {
                    if (el) {
                      widgetRefs.current.set(w.id, el)
                    } else {
                      widgetRefs.current.delete(w.id)
                    }
                  }}
                >
                  {w.render()}
                </SortableWidget>
              )
            })}
            {/* Drop indicator at the end */}
            {activeId !== null && dropIndex === visible.length - 1 && (
              <div className="col-span-12 flex items-center justify-center py-2">
                <div className="h-1 w-full max-w-md rounded-full bg-terminal-green animate-pulse" />
              </div>
            )}
          </DroppableGrid>
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
