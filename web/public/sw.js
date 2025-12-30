const CACHE_NAME = 'quantlete-pwa-v1'

const PRECACHE_URLS = ['/', '/index.html', '/manifest.json', '/icons/icon-192.png', '/icons/icon-512.png']

self.addEventListener('install', (event) => {
  event.waitUntil(
    caches
      .open(CACHE_NAME)
      .then((cache) => cache.addAll(PRECACHE_URLS))
      .then(() => self.skipWaiting())
  )
})

self.addEventListener('activate', (event) => {
  event.waitUntil(
    (async () => {
      const keys = await caches.keys()
      await Promise.all(keys.filter((k) => k !== CACHE_NAME).map((k) => caches.delete(k)))
      await self.clients.claim()
    })()
  )
})

self.addEventListener('fetch', (event) => {
  const { request } = event
  if (request.method !== 'GET') return

  const url = new URL(request.url)
  if (url.origin !== self.location.origin) return

  // Never cache API responses.
  if (url.pathname.startsWith('/api/')) return

  // SPA navigations: network-first with offline fallback to cached shell.
  if (request.mode === 'navigate') {
    event.respondWith(
      (async () => {
        try {
          const res = await fetch(request)
          const cache = await caches.open(CACHE_NAME)
          cache.put('/index.html', res.clone())
          return res
        } catch {
          const cached = await caches.match('/index.html')
          return cached || Response.error()
        }
      })()
    )
    return
  }

  // Static assets: cache-first.
  const dest = request.destination
  const isStatic =
    dest === 'script' ||
    dest === 'style' ||
    dest === 'image' ||
    dest === 'font' ||
    url.pathname.startsWith('/assets/') ||
    url.pathname.startsWith('/icons/')

  if (isStatic) {
    event.respondWith(
      (async () => {
        const cached = await caches.match(request)
        if (cached) return cached
        const res = await fetch(request)
        const cache = await caches.open(CACHE_NAME)
        cache.put(request, res.clone())
        return res
      })()
    )
  }
})
