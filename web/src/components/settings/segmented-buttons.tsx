import { cn } from '@/lib/utils'

interface SegmentedButtonsProps<T extends string> {
  value: T
  options: { value: T; label: string }[]
  onChange: (v: T) => void
}

export function SegmentedButtons<T extends string>({
  value,
  options,
  onChange,
}: SegmentedButtonsProps<T>) {
  return (
    <div className="inline-flex items-center gap-0.5 rounded-md border border-border bg-muted/50 p-1">
      {options.map((opt) => {
        const isSelected = opt.value === value
        return (
          <button
            key={opt.value}
            type="button"
            className={cn(
              'h-7 px-3 text-xs font-medium rounded-sm transition-all',
              isSelected
                ? 'bg-background text-foreground shadow-sm ring-1 ring-border/50'
                : 'text-muted-foreground hover:text-foreground hover:bg-muted/50'
            )}
            onClick={() => onChange(opt.value)}
          >
            {opt.label}
          </button>
        )
      })}
    </div>
  )
}
