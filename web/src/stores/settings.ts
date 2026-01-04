import { create } from 'zustand'
import { persist } from 'zustand/middleware'

import { SETTINGS_STORAGE_KEY } from '@/lib/constants'

export type UnitSystem = 'metric' | 'imperial'
export type Theme = 'light' | 'dark' | 'system'

interface SettingsState {
  unitSystem: UnitSystem
  theme: Theme
  setUnitSystem: (system: UnitSystem) => void
  setTheme: (theme: Theme) => void
}

export const useSettingsStore = create<SettingsState>()(
  persist(
    (set) => ({
      unitSystem: 'metric',
      theme: 'dark',
      setUnitSystem: (unitSystem) => set({ unitSystem }),
      setTheme: (theme) => set({ theme }),
    }),
    {
      name: SETTINGS_STORAGE_KEY,
    }
  )
)
