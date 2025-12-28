import { useSortable } from '@dnd-kit/sortable'
import { CSS } from '@dnd-kit/utilities'
import { cn } from '@/lib/utils/cn'
import type { WidgetWidth } from '@/lib/api/dashboard'
import { GripVertical, EyeOff, Columns2, Columns3, Columns4, Square } from 'lucide-react'
import type { ReactNode } from 'react'

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
    transform,
    transition,
    isDragging,
    isOver,
  } = useSortable({ id, disabled: !editMode })

  const style = {
    transform: CSS.Transform.toString(transform),
    transition,
  }

  return (
    <div
      ref={setNodeRef}
      style={style}
      className={cn(
        'relative md:col-span-12 transition-all duration-200',
        width === 4 && 'md:col-span-4',
        width === 6 && 'md:col-span-6',
        width === 8 && 'md:col-span-8',
        width === 12 && 'md:col-span-12',
        // Dragging states
        isDragging && 'opacity-40 scale-[0.98]',
        isOver && !isDragging && 'ring-2 ring-terminal-green ring-offset-2 ring-offset-background',
        // Edit mode visual
        editMode && 'group'
      )}
    >
      {/* Edit Mode Overlay */}
      {editMode && (
        <div
          className={cn(
            'absolute inset-0 z-10 rounded-sm transition-all duration-200',
            'border-2 border-transparent',
            'group-hover:border-terminal-green/50 group-hover:bg-terminal-green/5',
            isDragging && 'border-terminal-green bg-terminal-green/10'
          )}
        >
          {/* Top Control Bar */}
          <div
            className={cn(
              'absolute -top-px left-0 right-0 flex items-center justify-between',
              'rounded-t-sm bg-card/95 backdrop-blur border border-border',
              'transform transition-all duration-200',
              'opacity-0 -translate-y-2 group-hover:opacity-100 group-hover:translate-y-0',
              isDragging && 'opacity-100 translate-y-0'
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
