import { create } from 'zustand'
import type { ActivityFilters } from '@/lib/api'

interface ActivityFiltersState {
  filters: ActivityFilters
  setFilters: (filters: Partial<ActivityFilters>) => void
  resetFilters: () => void
}

const defaultFilters: ActivityFilters = {
  page: 1,
  per_page: 50,
  order_by: 'start_date',
  order_dir: 'desc',
}

export const useActivityFiltersStore = create<ActivityFiltersState>((set) => ({
  filters: defaultFilters,
  setFilters: (updates) =>
    set((state) => ({
      filters: { ...state.filters, ...updates },
    })),
  resetFilters: () => set({ filters: defaultFilters }),
}))
