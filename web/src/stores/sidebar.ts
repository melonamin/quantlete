import { create } from 'zustand'
import { persist } from 'zustand/middleware'

interface SidebarState {
  collapsed: boolean
  mobileOpen: boolean
  setCollapsed: (collapsed: boolean) => void
  toggleCollapsed: () => void
  setMobileOpen: (open: boolean) => void
  toggleMobileOpen: () => void
}

export const useSidebarStore = create<SidebarState>()(
  persist(
    (set, get) => ({
      collapsed: false,
      mobileOpen: false,
      setCollapsed: (collapsed) => set({ collapsed }),
      toggleCollapsed: () => set({ collapsed: !get().collapsed }),
      setMobileOpen: (mobileOpen) => set({ mobileOpen }),
      toggleMobileOpen: () => set({ mobileOpen: !get().mobileOpen }),
    }),
    {
      name: 'stata-sidebar',
      partialize: (state) => ({ collapsed: state.collapsed }),
    }
  )
)
