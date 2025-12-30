import { create } from 'zustand'
import type { ActivityFilters } from '@/lib/api'
import { PAGINATION } from '@/lib/constants'

interface ActivityFiltersState {
  filters: ActivityFilters
  setFilters: (filters: Partial<ActivityFilters>) => void
  resetFilters: () => void
}

const defaultFilters: ActivityFilters = {
  page: PAGINATION.DEFAULT_PAGE,
  per_page: PAGINATION.DEFAULT_PER_PAGE,
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
