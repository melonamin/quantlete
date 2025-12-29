import { create } from 'zustand'

interface SyncModalState {
  open: boolean
  openModal: () => void
  closeModal: () => void
}

export const useSyncModalStore = create<SyncModalState>()((set) => ({
  open: false,
  openModal: () => set({ open: true }),
  closeModal: () => set({ open: false }),
}))
