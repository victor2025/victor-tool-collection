const CACHE = 'vtc-v1';
const PRECACHE = [
  '/',
  '/manifest.json',
  '/icon.svg',
  '/base64/',
  '/qrcode/',
  '/json-formatter/',
  '/jwt-decoder/',
  '/webshell/',
  '/timestamp/',
];

self.addEventListener('install', (e) => {
  self.skipWaiting();
  e.waitUntil(
    caches.open(CACHE).then(cache => cache.addAll(PRECACHE))
  );
});

self.addEventListener('activate', (e) => {
  e.waitUntil(
    caches.keys().then(keys => Promise.all(
      keys.filter(k => k !== CACHE).map(k => caches.delete(k))
    ))
  );
});

self.addEventListener('fetch', (e) => {
  e.respondWith(
    caches.match(e.request).then(r => r || fetch(e.request))
  );
});
