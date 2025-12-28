import { create } from 'zustand'
import type { DashboardConfig, DashboardWidgetConfig, WidgetWidth } from '@/lib/api/dashboard'

interface DashboardLayoutState {
  config: DashboardConfig | null
  editMode: boolean
  setConfig: (config: DashboardConfig) => void
  setEditMode: (editMode: boolean) => void
  reset: (config: DashboardConfig) => void
  ensureWidget: (widget: DashboardWidgetConfig) => void
  moveWidget: (activeId: string, overId: string) => void
  setWidgetHidden: (id: string, hidden: boolean) => void
  setWidgetWidth: (id: string, width: WidgetWidth) => void
}

function upsertWidget(widgets: DashboardWidgetConfig[], widget: DashboardWidgetConfig) {
  const idx = widgets.findIndex((w) => w.id === widget.id)
  if (idx === -1) return [...widgets, widget]
  const next = widgets.slice()
  next[idx] = { ...next[idx], ...widget }
  return next
}

export const useDashboardLayoutStore = create<DashboardLayoutState>((set, get) => ({
  config: null,
  editMode: false,
  setConfig: (config) => set({ config }),
  setEditMode: (editMode) => set({ editMode }),
  reset: (config) => set({ config }),
  ensureWidget: (widget) => {
    const config = get().config
    if (!config) return
    set({ config: { ...config, widgets: upsertWidget(config.widgets, widget) } })
  },
  moveWidget: (activeId, overId) => {
    const config = get().config
    if (!config) return
    if (activeId === overId) return
    const widgets = config.widgets.slice()
    const from = widgets.findIndex((w) => w.id === activeId)
    const to = widgets.findIndex((w) => w.id === overId)
    if (from === -1 || to === -1) return
    const [moved] = widgets.splice(from, 1)
    widgets.splice(to, 0, moved)
    set({ config: { ...config, widgets } })
  },
  setWidgetHidden: (id, hidden) => {
    const config = get().config
    if (!config) return
    const widgets = config.widgets.map((w) => (w.id === id ? { ...w, hidden } : w))
    set({ config: { ...config, widgets } })
  },
  setWidgetWidth: (id, width) => {
    const config = get().config
    if (!config) return
    const widgets = config.widgets.map((w) => (w.id === id ? { ...w, width } : w))
    set({ config: { ...config, widgets } })
  },
}))
