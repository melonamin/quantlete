import { Button } from '@/components/ui/button'

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
    <div className="inline-flex items-center gap-1 rounded-md border border-border bg-muted/30 p-1">
      {options.map((opt) => (
        <Button
          key={opt.value}
          type="button"
          size="sm"
          variant={opt.value === value ? 'secondary' : 'ghost'}
          className="h-7 px-2"
          onClick={() => onChange(opt.value)}
        >
          {opt.label}
        </Button>
      ))}
    </div>
  )
}
