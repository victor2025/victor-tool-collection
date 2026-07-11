var CACHE = 'vtc-v4';
var PRECACHE = [
  '/manifest.json',
  '/icon.svg',
];

self.addEventListener('install', function (e) {
  self.skipWaiting();
  e.waitUntil(
    caches.open(CACHE).then(function (cache) { return cache.addAll(PRECACHE); })
  );
});

self.addEventListener('activate', function (e) {
  e.waitUntil(
    Promise.all([
      caches.keys().then(function (keys) {
        return Promise.all(
          keys.filter(function (k) { return k !== CACHE; }).map(function (k) { return caches.delete(k); })
        );
      }),
      self.clients.claim(),
    ])
  );
});

self.addEventListener('fetch', function (e) {
  var url = new URL(e.request.url);
  var path = url.pathname;

  // tracker.js 无论版本号都不缓存
  if (path === '/tracker.js') {
    e.respondWith(fetch(e.request));
    return;
  }

  if (path.match(/\.(js|css|png|jpg|svg|ico|woff2?)$/)) {
    e.respondWith(
      caches.match(e.request).then(function (r) { return r || fetch(e.request); })
    );
    return;
  }

  e.respondWith(fetch(e.request));
});
