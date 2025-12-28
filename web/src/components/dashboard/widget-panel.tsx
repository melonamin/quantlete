import { useEffect, useState } from 'react'
import {
  DndContext,
  closestCenter,
  KeyboardSensor,
  PointerSensor,
  useSensor,
  useSensors,
  type DragEndEvent,
} from '@dnd-kit/core'
import {
  arrayMove,
  SortableContext,
  sortableKeyboardCoordinates,
  verticalListSortingStrategy,
  useSortable,
} from '@dnd-kit/sortable'
import { CSS } from '@dnd-kit/utilities'
import { cn } from '@/lib/utils/cn'
import { X, GripVertical, Eye, EyeOff, LayoutGrid } from 'lucide-react'

interface Widget {
  id: string
  title: string
  hidden: boolean
}

interface WidgetPanelProps {
  open: boolean
  onClose: () => void
  widgets: Widget[]
  onToggleWidget: (id: string, hidden: boolean) => void
  onReorder: (newOrder: string[]) => void
}

export function WidgetPanel({
  open,
  onClose,
  widgets,
  onToggleWidget,
  onReorder,
}: WidgetPanelProps) {
  const [localWidgets, setLocalWidgets] = useState(widgets)

  // Sync local state when widgets change
  useEffect(() => {
    setLocalWidgets(widgets)
  }, [widgets])

  // Handle escape key
  useEffect(() => {
    if (!open) return
    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === 'Escape') onClose()
    }
    window.addEventListener('keydown', handleKeyDown)
    return () => window.removeEventListener('keydown', handleKeyDown)
  }, [open, onClose])

  const sensors = useSensors(
    useSensor(PointerSensor, {
      activationConstraint: { distance: 5 },
    }),
    useSensor(KeyboardSensor, {
      coordinateGetter: sortableKeyboardCoordinates,
    })
  )

  const handleDragEnd = (event: DragEndEvent) => {
    const { active, over } = event
    if (!over || active.id === over.id) return

    const oldIndex = localWidgets.findIndex((w) => w.id === active.id)
    const newIndex = localWidgets.findIndex((w) => w.id === over.id)
    const newWidgets = arrayMove(localWidgets, oldIndex, newIndex)
    setLocalWidgets(newWidgets)
    onReorder(newWidgets.map((w) => w.id))
  }

  const visibleCount = localWidgets.filter((w) => !w.hidden).length
  const hiddenCount = localWidgets.filter((w) => w.hidden).length

  return (
    <>
      {/* Backdrop */}
      <div
        className={cn(
          'fixed inset-0 z-40 bg-background/60 backdrop-blur-sm transition-opacity duration-300',
          open ? 'opacity-100' : 'opacity-0 pointer-events-none'
        )}
        onClick={onClose}
      />

      {/* Panel */}
      <div
        className={cn(
          'fixed right-0 top-0 z-50 h-full w-full max-w-md',
          'bg-card border-l border-border shadow-2xl',
          'transform transition-transform duration-300 ease-out',
          open ? 'translate-x-0' : 'translate-x-full'
        )}
      >
        {/* Header */}
        <div className="flex items-center justify-between border-b border-border px-6 py-4">
          <div className="flex items-center gap-3">
            <div className="flex h-8 w-8 items-center justify-center rounded-sm bg-terminal-green/10">
              <LayoutGrid className="h-4 w-4 text-terminal-green" />
            </div>
            <div>
              <h2 className="text-sm font-semibold">Widget Manager</h2>
              <p className="text-xs text-muted-foreground">
                {visibleCount} visible • {hiddenCount} hidden
              </p>
            </div>
          </div>
          <button
            onClick={onClose}
            className="flex h-8 w-8 items-center justify-center rounded-sm text-muted-foreground transition-colors hover:bg-accent hover:text-foreground"
          >
            <X className="h-4 w-4" />
          </button>
        </div>

        {/* Instructions */}
        <div className="border-b border-border bg-muted/30 px-6 py-3">
          <p className="text-xs text-muted-foreground">
            Drag to reorder widgets. Click the eye icon to show/hide.
          </p>
        </div>

        {/* Widget List */}
        <div className="h-[calc(100%-140px)] overflow-y-auto px-4 py-4">
          <DndContext
            sensors={sensors}
            collisionDetection={closestCenter}
            onDragEnd={handleDragEnd}
          >
            <SortableContext
              items={localWidgets.map((w) => w.id)}
              strategy={verticalListSortingStrategy}
            >
              <div className="space-y-2">
                {localWidgets.map((widget, index) => (
                  <SortableWidgetItem
                    key={widget.id}
                    widget={widget}
                    index={index}
                    onToggle={() => onToggleWidget(widget.id, !widget.hidden)}
                  />
                ))}
              </div>
            </SortableContext>
          </DndContext>
        </div>

        {/* Footer */}
        <div className="absolute bottom-0 left-0 right-0 border-t border-border bg-card px-6 py-4">
          <button
            onClick={onClose}
            className="w-full rounded-sm bg-terminal-green py-2 text-sm font-medium text-primary-foreground transition-colors hover:bg-terminal-green/90"
          >
            Done
          </button>
        </div>
      </div>
    </>
  )
}

interface SortableWidgetItemProps {
  widget: Widget
  index: number
  onToggle: () => void
}

function SortableWidgetItem({ widget, index, onToggle }: SortableWidgetItemProps) {
  const {
    attributes,
    listeners,
    setNodeRef,
    transform,
    transition,
    isDragging,
  } = useSortable({ id: widget.id })

  const style = {
    transform: CSS.Transform.toString(transform),
    transition,
  }

  return (
    <div
      ref={setNodeRef}
      style={style}
      className={cn(
        'group flex items-center gap-3 rounded-sm border border-border bg-card p-3',
        'transition-all duration-200',
        isDragging && 'shadow-lg shadow-terminal-green/10 border-terminal-green z-10',
        widget.hidden && 'opacity-50'
      )}
    >
      {/* Drag Handle */}
      <button
        {...attributes}
        {...listeners}
        className="flex h-6 w-6 cursor-grab items-center justify-center rounded-sm text-muted-foreground transition-colors hover:bg-accent hover:text-foreground active:cursor-grabbing"
      >
        <GripVertical className="h-4 w-4" />
      </button>

      {/* Index */}
      <div className="flex h-6 w-6 items-center justify-center rounded-sm bg-muted text-xs font-medium text-muted-foreground">
        {index + 1}
      </div>

      {/* Title */}
      <span
        className={cn(
          'flex-1 text-sm transition-colors',
          widget.hidden ? 'text-muted-foreground line-through' : 'text-foreground'
        )}
      >
        {widget.title}
      </span>

      {/* Visibility Toggle */}
      <button
        onClick={onToggle}
        className={cn(
          'flex h-8 w-8 items-center justify-center rounded-sm transition-all',
          widget.hidden
            ? 'text-muted-foreground hover:bg-terminal-green/10 hover:text-terminal-green'
            : 'bg-terminal-green/10 text-terminal-green hover:bg-terminal-green/20'
        )}
        title={widget.hidden ? 'Show widget' : 'Hide widget'}
      >
        {widget.hidden ? (
          <EyeOff className="h-4 w-4" />
        ) : (
          <Eye className="h-4 w-4" />
        )}
      </button>
    </div>
  )
}
