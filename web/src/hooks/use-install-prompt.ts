import { useCallback, useEffect, useState } from 'react'

type UserChoiceOutcome = 'accepted' | 'dismissed'

type BeforeInstallPromptEvent = Event & {
  prompt: () => Promise<void>
  userChoice: Promise<{ outcome: UserChoiceOutcome; platform: string }>
}

export function useInstallPrompt() {
  const [deferredPrompt, setDeferredPrompt] = useState<BeforeInstallPromptEvent | null>(null)
  const [isInstalled, setIsInstalled] = useState(false)

  useEffect(() => {
    const onBeforeInstallPrompt = (e: Event) => {
      // Chrome on Android fires this; preventing default lets us show our own UI.
      e.preventDefault()
      setDeferredPrompt(e as BeforeInstallPromptEvent)
    }

    const onAppInstalled = () => {
      setIsInstalled(true)
      setDeferredPrompt(null)
    }

    window.addEventListener('beforeinstallprompt', onBeforeInstallPrompt)
    window.addEventListener('appinstalled', onAppInstalled)

    return () => {
      window.removeEventListener('beforeinstallprompt', onBeforeInstallPrompt)
      window.removeEventListener('appinstalled', onAppInstalled)
    }
  }, [])

  const promptInstall = useCallback(async () => {
    if (!deferredPrompt) return null
    await deferredPrompt.prompt()
    const choice = await deferredPrompt.userChoice
    setDeferredPrompt(null)
    return choice
  }, [deferredPrompt])

  return {
    canInstall: !!deferredPrompt && !isInstalled,
    promptInstall,
    isInstalled,
  }
}

