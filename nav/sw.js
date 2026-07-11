const CACHE = 'vtc-v2';
const PRECACHE = [
  '/manifest.json',
  '/icon.svg',
];

self.addEventListener('install', (e) => {
  self.skipWaiting();
  e.waitUntil(
    caches.open(CACHE).then(cache => cache.addAll(PRECACHE))
  );
});

self.addEventListener('activate', (e) => {
  e.waitUntil(
    Promise.all([
      caches.keys().then(keys => Promise.all(
        keys.filter(k => k !== CACHE).map(k => caches.delete(k))
      )),
      self.clients.claim(),
    ])
  );
});

self.addEventListener('fetch', (e) => {
  const url = new URL(e.request.url);
  // 只缓存静态资源，HTML 页面不缓存（方便更新）
  if (url.pathname.match(/\.(js|css|png|jpg|svg|ico|woff2?)$/)) {
    e.respondWith(
      caches.match(e.request).then(r => r || fetch(e.request))
    );
    return;
  }
  // 其他全部走网络
  e.respondWith(fetch(e.request));
});
