self.addEventListener('install', (e) => {
  self.skipWaiting()
})

self.addEventListener('activate', (e) => {
  e.waitUntil(clients.claim())
})

self.addEventListener('fetch', (e) => {
  // Only cache same-origin GET requests
  if (e.request.method !== 'GET' || !e.request.url.startsWith(self.location.origin)) return
  // Don't cache API calls
  if (e.request.url.includes('/api/')) return

  e.respondWith(
    caches.open('stock-resonance-v1').then(cache => {
      return fetch(e.request).then(response => {
        cache.put(e.request, response.clone())
        return response
      }).catch(() => {
        return cache.match(e.request).then(cached => cached || new Response('Offline', { status: 503 }))
      })
    })
  )
})
