const CACHE_NAME = 'sismomonitor-v5';
const API_CACHE_NAME = 'sismomonitor-api-v1';
const OFFLINE_URL = '/offline.html';
const API_CACHE_TTL_MS = 24 * 60 * 60 * 1000; // 24h aligned with live data window

self.addEventListener('install', (event) => {
  self.skipWaiting();
  event.waitUntil(
    caches.open(CACHE_NAME).then((cache) => {
      return cache.add(new Request(OFFLINE_URL, { cache: 'reload' }));
    })
  );
});

self.addEventListener('activate', (event) => {
  event.waitUntil(
    caches.keys().then((cacheNames) => {
      return Promise.all(
        cacheNames.map((cacheName) => {
          // Keep runtime API cache; clean old app/offline caches
          if (cacheName !== CACHE_NAME && cacheName !== API_CACHE_NAME) {
            return caches.delete(cacheName);
          }
        })
      );
    }).then(() => self.clients.claim())
  );
});

async function getCachedEarthquakes(request) {
  const cache = await caches.open(API_CACHE_NAME);
  const cached = await cache.match(request);
  const now = Date.now();

  if (cached) {
    const cachedTime = parseInt(cached.headers.get('x-sismo-cached-at') || '0', 10);
    if (now - cachedTime < API_CACHE_TTL_MS) {
      return cached;
    }
  }

  try {
    const networkResponse = await fetch(request);
    if (networkResponse && networkResponse.status === 200) {
      // Store a copy with a timestamp header
      const headers = new Headers(networkResponse.headers);
      headers.set('x-sismo-cached-at', String(now));
      const body = await networkResponse.clone().blob();
      const responseToCache = new Response(body, { status: networkResponse.status, statusText: networkResponse.statusText, headers });
      await cache.put(request, responseToCache);
    }
    return networkResponse;
  } catch (err) {
    if (cached) {
      return cached;
    }
    throw err;
  }
}

self.addEventListener('fetch', (event) => {
  const url = new URL(event.request.url);

  // Cache initial earthquakes with a 24h TTL
  if (url.pathname === '/api/earthquakes') {
    event.respondWith(getCachedEarthquakes(event.request));
    return;
  }

  // CRITICAL: We MUST NOT intercept EventSource streams (/api/stream) or we hang the app.
  if (event.request.url.includes('/api/')) {
    return; // Do nothing, let the browser handle it natively
  }

  // Handle navigation requests (when the user requests an HTML page)
  if (event.request.mode === 'navigate') {
    event.respondWith(
      fetch(event.request).catch(() => {
        // If the network request fails (e.g., offline), serve the offline fallback page
        return caches.match(OFFLINE_URL);
      })
    );
  }
});

self.addEventListener('push', (event) => {
  let data = { title: 'Nuevo Sismo', body: 'Alerta de sismo detectada.', url: '/' };
  
  if (event.data) {
    try {
      data = event.data.json();
    } catch {
      data.body = event.data.text();
    }
  }

  const options = {
    body: data.body,
    icon: '/favicon/android-chrome-192x192.png',
    badge: '/favicon/android-chrome-192x192.png',
    vibrate: [200, 100, 200, 100, 200, 100, 200],
    data: {
      url: data.url
    }
  };

  event.waitUntil(
    self.registration.showNotification(data.title, options)
  );
});

self.addEventListener('notificationclick', (event) => {
  event.notification.close();

  event.waitUntil(
    clients.matchAll({ type: 'window', includeUncontrolled: true }).then((windowClients) => {
      // If a window is already open, focus it
      for (let client of windowClients) {
        if (client.url === self.registration.scope && 'focus' in client) {
          return client.focus();
        }
      }
      // Otherwise, open a new window
      if (clients.openWindow) {
        return clients.openWindow(event.notification.data.url || '/');
      }
    })
  );
});
