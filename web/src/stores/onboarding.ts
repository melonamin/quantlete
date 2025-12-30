import { create } from 'zustand'

const ONBOARDING_STORAGE_KEY = 'quantlete_onboarding_seen'

interface OnboardingState {
  dismissed: boolean
  dismiss: () => void
  isComplete: () => boolean
}

export const useOnboardingStore = create<OnboardingState>()((set) => ({
  dismissed: localStorage.getItem(ONBOARDING_STORAGE_KEY) === 'true',
  dismiss: () => {
    localStorage.setItem(ONBOARDING_STORAGE_KEY, 'true')
    set({ dismissed: true })
  },
  isComplete: () => localStorage.getItem(ONBOARDING_STORAGE_KEY) === 'true',
}))

/**
 * Mark onboarding as complete (called after successful OAuth).
 */
export function markOnboardingComplete(): void {
  localStorage.setItem(ONBOARDING_STORAGE_KEY, 'true')
  useOnboardingStore.setState({ dismissed: true })
}
