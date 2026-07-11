/**
 * VTC Tracker — 页面访问上报组件
 *
 * 仅在 iframe（Tab 模式）中通过 nav 的 postMessage 触发上报。
 * 非 tab 模式（直接打开页面）一律不报。
 */
(function () {
  if (window.__vtc_track_init) return;
  window.__vtc_track_init = true;

  var tool = window.location.pathname
    .replace(/\/$/, '')
    .split('/')
    .pop() || 'home';

  // 非 iframe 模式（直接打开页面）：彻底不报
  if (window.self === window.top) return;

  // Tab 模式（iframe）：nav 切 Tab 时发 postMessage 才上报
  function report() {
    fetch('/api/visit', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ tool: tool }),
      keepalive: true,
    }).catch(function () {});
  }

  window.addEventListener('message', function (e) {
    if (e.data && e.data.type === 'vtc_track_visit') report();
  });
})();
