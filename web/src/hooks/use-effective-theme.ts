import { useEffect, useState } from 'react'

import { useSettingsStore } from '@/stores/settings'

/**
 * Returns the current effective theme ('dark' or 'light'), resolving 'system'
 * to the actual OS preference. Reactively updates when system preference changes.
 */
export function useEffectiveTheme(): 'dark' | 'light' {
  const theme = useSettingsStore((s) => s.theme)
  const [systemPrefersDark, setSystemPrefersDark] = useState(() =>
    window.matchMedia('(prefers-color-scheme: dark)').matches
  )

  useEffect(() => {
    if (theme !== 'system') return

    const mql = window.matchMedia('(prefers-color-scheme: dark)')
    const onChange = (e: MediaQueryListEvent) => setSystemPrefersDark(e.matches)

    mql.addEventListener('change', onChange)
    return () => mql.removeEventListener('change', onChange)
  }, [theme])

  if (theme === 'system') {
    return systemPrefersDark ? 'dark' : 'light'
  }
  return theme
}
