import { create } from 'zustand'
import { persist } from 'zustand/middleware'

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
      theme: 'system',
      setUnitSystem: (unitSystem) => set({ unitSystem }),
      setTheme: (theme) => set({ theme }),
    }),
    {
      name: 'stata-settings',
    }
  )
)
