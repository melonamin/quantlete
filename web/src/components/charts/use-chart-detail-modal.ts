import { useState, useCallback } from 'react'
import type { ChartDetailData } from './chart-detail-modal'

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
