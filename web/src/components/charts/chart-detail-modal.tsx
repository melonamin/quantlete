import { type ReactNode } from 'react'
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { X } from 'lucide-react'

export interface ChartDetailData {
  title: string
  subtitle?: string
  stats?: { label: string; value: string | number }[]
  content?: ReactNode
}

interface ChartDetailModalProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  data: ChartDetailData | null
}

export function ChartDetailModal({ open, onOpenChange, data }: ChartDetailModalProps) {
  if (!data) return null

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-lg">
        <DialogHeader>
          <DialogTitle className="flex items-center justify-between">
            <span>{data.title}</span>
          </DialogTitle>
          {data.subtitle && (
            <p className="text-sm text-muted-foreground">{data.subtitle}</p>
          )}
        </DialogHeader>

        {data.stats && data.stats.length > 0 && (
          <div className="grid grid-cols-2 gap-3">
            {data.stats.map((stat, idx) => (
              <div key={idx} className="rounded-lg border border-border p-3">
                <div className="text-xs text-muted-foreground">{stat.label}</div>
                <div className="text-lg font-semibold">{stat.value}</div>
              </div>
            ))}
          </div>
        )}

        {data.content && <div className="mt-4">{data.content}</div>}
      </DialogContent>
    </Dialog>
  )
}

// Hook for managing chart detail modal state
import { useState, useCallback } from 'react'

export function useChartDetailModal() {
  const [open, setOpen] = useState(false)
  const [data, setData] = useState<ChartDetailData | null>(null)

  const showDetail = useCallback((detail: ChartDetailData) => {
    setData(detail)
    setOpen(true)
  }, [])

  const hideDetail = useCallback(() => {
    setOpen(false)
  }, [])

  return {
    open,
    data,
    showDetail,
    hideDetail,
    setOpen,
  }
}
