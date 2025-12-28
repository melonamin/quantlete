import { useSortable } from '@dnd-kit/sortable'
import { cn } from '@/lib/utils/cn'
import type { WidgetWidth } from '@/lib/api/dashboard'
import { GripVertical, EyeOff, Columns2, Columns3, Columns4, Square } from 'lucide-react'
import type { ReactNode, CSSProperties } from 'react'

interface SortableWidgetProps {
  id: string
  title: string
  width: WidgetWidth
  editMode: boolean
  onWidthChange: (width: WidgetWidth) => void
  onHide: () => void
  children: ReactNode
}

const WIDTH_OPTIONS: { value: WidgetWidth; icon: typeof Square; label: string }[] = [
  { value: 4, icon: Columns4, label: '33%' },
  { value: 6, icon: Columns3, label: '50%' },
  { value: 8, icon: Columns2, label: '66%' },
  { value: 12, icon: Square, label: '100%' },
]

export function SortableWidget({
  id,
  title,
  width,
  editMode,
  onWidthChange,
  onHide,
  children,
}: SortableWidgetProps) {
  const {
    attributes,
    listeners,
    setNodeRef,
    isDragging,
    isOver,
  } = useSortable({ id, disabled: !editMode })

  // For variable-width grid items, we don't apply transform animations
  // The grid handles positioning - we only use DragOverlay for drag visual
  const style: CSSProperties = {
    opacity: isDragging ? 0.3 : 1,
    // Prevent the item from "jumping" by not applying transforms
  }

  return (
    <div
      ref={setNodeRef}
      style={style}
      className={cn(
        'relative md:col-span-12',
        width === 4 && 'md:col-span-4',
        width === 6 && 'md:col-span-6',
        width === 8 && 'md:col-span-8',
        width === 12 && 'md:col-span-12',
        // Drop target indicator
        isOver && !isDragging && 'ring-2 ring-terminal-green ring-offset-2 ring-offset-background',
        // Edit mode visual
        editMode && 'group'
      )}
    >
      {/* Edit Mode Overlay */}
      {editMode && (
        <div
          className={cn(
            'absolute inset-0 z-10 rounded-sm',
            'border-2 border-transparent',
            !isDragging && 'group-hover:border-terminal-green/50 group-hover:bg-terminal-green/5',
            isDragging && 'border-dashed border-terminal-green/50 bg-terminal-green/5'
          )}
        >
          {/* Top Control Bar - hidden during drag */}
          {!isDragging && (
            <div
              className={cn(
                'absolute -top-px left-0 right-0 flex items-center justify-between',
                'rounded-t-sm bg-card/95 backdrop-blur border border-border',
                'opacity-0 group-hover:opacity-100'
              )}
            >
            {/* Drag Handle */}
            <button
              {...attributes}
              {...listeners}
              className={cn(
                'flex items-center gap-2 px-3 py-2 cursor-grab active:cursor-grabbing',
                'text-muted-foreground hover:text-terminal-green transition-colors',
                'border-r border-border'
              )}
            >
              <GripVertical className="h-4 w-4" />
              <span className="text-xs font-medium uppercase tracking-wider">{title}</span>
            </button>

            {/* Width Controls */}
            <div className="flex items-center">
              <div className="flex items-center border-r border-border">
                {WIDTH_OPTIONS.map((opt) => {
                  const Icon = opt.icon
                  const isActive = width === opt.value
                  return (
                    <button
                      key={opt.value}
                      onClick={() => onWidthChange(opt.value)}
                      className={cn(
                        'flex items-center justify-center w-8 h-8 transition-colors',
                        isActive
                          ? 'bg-terminal-green/20 text-terminal-green'
                          : 'text-muted-foreground hover:text-foreground hover:bg-accent'
                      )}
                      title={opt.label}
                    >
                      <Icon className="h-3.5 w-3.5" />
                    </button>
                  )
                })}
              </div>

              {/* Hide Button */}
              <button
                onClick={onHide}
                className="flex items-center justify-center w-8 h-8 text-muted-foreground hover:text-destructive transition-colors"
                title="Hide widget"
              >
                <EyeOff className="h-3.5 w-3.5" />
              </button>
            </div>
          </div>

          {/* Resize Indicator Corners */}
          <div className="absolute bottom-1 right-1 opacity-0 group-hover:opacity-100 transition-opacity">
            <div className="flex items-center gap-0.5 text-[10px] text-muted-foreground bg-card/80 px-1.5 py-0.5 rounded-sm">
              {width === 4 && '1/3'}
              {width === 6 && '1/2'}
              {width === 8 && '2/3'}
              {width === 12 && 'FULL'}
            </div>
          </div>
        </div>
      )}

      {/* Widget Content */}
      <div className={cn(editMode && 'pointer-events-none')}>{children}</div>
    </div>
  )
}
