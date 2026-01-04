import type { LucideIcon } from 'lucide-react'
import { cn } from '@/lib/utils'

interface SectionHeaderProps {
  icon: LucideIcon
  title: string
  description?: string
  className?: string
}

export function SectionHeader({ icon: Icon, title, description, className }: SectionHeaderProps) {
  return (
    <div className={cn('mb-4', className)}>
      <h2 className="flex items-center gap-2.5 text-lg font-semibold tracking-tight">
        <Icon className="h-5 w-5 text-muted-foreground" />
        {title}
      </h2>
      {description && <p className="mt-1 text-sm text-muted-foreground ml-7.5">{description}</p>}
    </div>
  )
}
