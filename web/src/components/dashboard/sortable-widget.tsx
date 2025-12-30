import { useCallback } from 'react'
import { useSortable } from '@dnd-kit/sortable'
import { cn } from '@/lib/utils/cn'
import type { WidgetWidth, WidgetHeight } from '@/lib/api'
import {
  GripVertical,
  EyeOff,
  Columns2,
  Columns3,
  Columns4,
  Square,
  ChevronsUpDown,
  MoreHorizontal,
} from 'lucide-react'
import { Popover, PopoverContent, PopoverTrigger } from '@/components/ui/popover'
import type { ReactNode, CSSProperties } from 'react'

interface SortableWidgetProps {
  id: string
  title: string
  width: WidgetWidth
  height: WidgetHeight
  editMode: boolean
  onWidthChange: (width: WidgetWidth) => void
  onHeightChange: (height: WidgetHeight) => void
  onHide: () => void
  children: ReactNode
  showDropIndicator?: boolean
  onRefChange?: (el: HTMLElement | null) => void
}

const WIDTH_OPTIONS: { value: WidgetWidth; icon: typeof Square; label: string }[] = [
  { value: 4, icon: Columns4, label: '33%' },
  { value: 6, icon: Columns3, label: '50%' },
  { value: 8, icon: Columns2, label: '66%' },
  { value: 12, icon: Square, label: '100%' },
]

const HEIGHT_OPTIONS: { value: WidgetHeight; label: string }[] = [
  { value: 1, label: 'Compact' },
  { value: 2, label: 'Standard' },
  { value: 3, label: 'Tall' },
]

// Height classes - grid row spans (works with grid-auto-rows on container)
const HEIGHT_CLASSES: Record<WidgetHeight, string> = {
  1: 'row-span-1',
  2: 'row-span-2',
  3: 'row-span-3',
}

export function SortableWidget({
  id,
  title,
  width,
  height,
  editMode,
  onWidthChange,
  onHeightChange,
  onHide,
  children,
  showDropIndicator,
  onRefChange,
}: SortableWidgetProps) {
  const { attributes, listeners, setNodeRef, isDragging, isOver } = useSortable({
    id,
    disabled: !editMode,
  })

  // Combined ref for both sortable and position tracking
  const combinedRef = useCallback(
    (el: HTMLElement | null) => {
      setNodeRef(el)
      onRefChange?.(el)
    },
    [setNodeRef, onRefChange]
  )

  // For variable-width grid items, we don't apply transform animations
  // The grid handles positioning - we only use DragOverlay for drag visual
  const style: CSSProperties = {
    opacity: isDragging ? 0.3 : 1,
  }

  // Use compact mode for narrowest widgets (col-span-4 = ~300px)
  const isCompact = width === 4

  // Compact controls popover content
  const CompactControls = () => (
    <div className="flex flex-col gap-3 p-3">
      {/* Width controls */}
      <div className="flex flex-col gap-1.5">
        <span className="text-[10px] text-muted-foreground uppercase tracking-wider">Width</span>
        <div className="flex gap-1">
          {WIDTH_OPTIONS.map((opt) => {
            const Icon = opt.icon
            const isActive = width === opt.value
            return (
              <button
                key={opt.value}
                onClick={() => onWidthChange(opt.value)}
                className={cn(
                  'flex items-center justify-center w-8 h-8 rounded-sm transition-colors',
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
      </div>
      {/* Height controls */}
      <div className="flex flex-col gap-1.5">
        <span className="text-[10px] text-muted-foreground uppercase tracking-wider">Height</span>
        <div className="flex gap-1">
          {HEIGHT_OPTIONS.map((opt) => {
            const isActive = height === opt.value
            return (
              <button
                key={opt.value}
                onClick={() => onHeightChange(opt.value)}
                className={cn(
                  'flex items-center justify-center w-8 h-8 rounded-sm text-[10px] font-medium transition-colors',
                  isActive
                    ? 'bg-terminal-green/20 text-terminal-green'
                    : 'text-muted-foreground hover:text-foreground hover:bg-accent'
                )}
                title={opt.label}
              >
                {opt.value}×
              </button>
            )
          })}
        </div>
      </div>
      {/* Hide button */}
      <button
        onClick={onHide}
        className="flex items-center gap-2 w-full px-2 py-1.5 rounded-sm text-xs text-muted-foreground hover:text-destructive hover:bg-destructive/10 transition-colors"
      >
        <EyeOff className="h-3.5 w-3.5" />
        Hide widget
      </button>
    </div>
  )

  return (
    <>
      {/* Drop indicator before this widget */}
      {showDropIndicator && (
        <div className="col-span-12 flex items-center justify-center py-1">
          <div className="h-1 w-full max-w-md rounded-full bg-terminal-green animate-pulse" />
        </div>
      )}
      <div
        ref={combinedRef}
        style={style}
        className={cn(
          'relative md:col-span-12',
          width === 4 && 'md:col-span-4',
          width === 6 && 'md:col-span-6',
          width === 8 && 'md:col-span-8',
          width === 12 && 'md:col-span-12',
          // Height classes on grid cell
          HEIGHT_CLASSES[height],
          // Drop target indicator
          isOver &&
            !isDragging &&
            'ring-2 ring-terminal-green ring-offset-2 ring-offset-background',
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
                    'border-r border-border',
                    isCompact && 'flex-1'
                  )}
                >
                  <GripVertical className="h-4 w-4" />
                  <span className={cn(
                    'text-xs font-medium uppercase tracking-wider',
                    isCompact && 'truncate max-w-[80px]'
                  )}>
                    {title}
                  </span>
                </button>

                {/* Controls: Compact (popover) or Full (inline) */}
                {isCompact ? (
                  <Popover>
                    <PopoverTrigger asChild>
                      <button
                        className="flex items-center justify-center w-10 h-10 text-muted-foreground hover:text-terminal-green transition-colors"
                        title="Widget settings"
                      >
                        <MoreHorizontal className="h-4 w-4" />
                      </button>
                    </PopoverTrigger>
                    <PopoverContent align="end" className="w-auto p-0">
                      <CompactControls />
                    </PopoverContent>
                  </Popover>
                ) : (
                  /* Full mode: inline controls */
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

                  {/* Height Controls */}
                  <div className="flex items-center border-r border-border">
                    <ChevronsUpDown className="h-3 w-3 mx-1 text-muted-foreground" />
                    {HEIGHT_OPTIONS.map((opt) => {
                      const isActive = height === opt.value
                      return (
                        <button
                          key={opt.value}
                          onClick={() => onHeightChange(opt.value)}
                          className={cn(
                            'flex items-center justify-center w-8 h-8 text-[10px] font-medium transition-colors',
                            isActive
                              ? 'bg-terminal-green/20 text-terminal-green'
                              : 'text-muted-foreground hover:text-foreground hover:bg-accent'
                          )}
                          title={opt.label}
                        >
                          {opt.value}×
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
                )}
              </div>
            )}

            {/* Resize Indicator Corners - also hidden during drag */}
            {!isDragging && (
              <div className="absolute bottom-1 right-1 opacity-0 group-hover:opacity-100">
                <div className="flex items-center gap-1 text-[10px] text-muted-foreground bg-card/80 px-1.5 py-0.5 rounded-sm">
                  <span>
                    {width === 4 && '1/3'}
                    {width === 6 && '1/2'}
                    {width === 8 && '2/3'}
                    {width === 12 && 'FULL'}
                  </span>
                  <span className="text-muted-foreground/60">×</span>
                  <span>{height}×</span>
                </div>
              </div>
            )}
          </div>
        )}

        {/* Widget Content - fills container height */}
        <div className={cn('h-full overflow-hidden', editMode && 'pointer-events-none')}>
          {children}
        </div>
      </div>
    </>
  )
}
