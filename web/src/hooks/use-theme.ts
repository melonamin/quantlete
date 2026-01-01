import { useEffect } from 'react'

import { useSettingsStore } from '@/stores/settings'

function supportsMatchMediaEventListener(mql: MediaQueryList): mql is MediaQueryList & {
  addEventListener: (type: 'change', listener: (ev: MediaQueryListEvent) => void) => void
  removeEventListener: (type: 'change', listener: (ev: MediaQueryListEvent) => void) => void
} {
  return typeof (mql as MediaQueryList).addEventListener === 'function'
}

export function useTheme() {
  const theme = useSettingsStore((s) => s.theme)

  useEffect(() => {
    const root = document.documentElement
    const mql = window.matchMedia('(prefers-color-scheme: dark)')

    const apply = () => {
      const isDark = theme === 'dark' || (theme === 'system' && mql.matches)
      root.classList.toggle('dark', isDark)
    }

    apply()

    if (theme !== 'system') return

    const onChange = () => apply()
    if (supportsMatchMediaEventListener(mql)) {
      mql.addEventListener('change', onChange)
      return () => mql.removeEventListener('change', onChange)
    }

    // Safari < 14 - uses deprecated addListener/removeListener API
    mql.addListener(onChange)
    return () => mql.removeListener(onChange)
  }, [theme])
}

