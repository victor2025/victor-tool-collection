/**
 * VTC Tracker — 页面访问上报组件
 *
 * 单页面模式（直接打开）：页面加载时 fetch 上报一次，永不重发
 * Tab 模式（iframe 内）：仅在切到对应 Tab 时上报（postMessage）
 * 其他情况一律不上报
 */
(function () {
  if (window.__vtc_track_init) return;
  window.__vtc_track_init = true;

  var tool = window.location.pathname
    .replace(/\/$/, '')
    .split('/')
    .pop() || 'home';

  function report() {
    // 页面级锁：同一个页面只会发出一次请求
    if (window.__vtc_report_sent) return;
    window.__vtc_report_sent = true;

    fetch('/api/visit', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ tool: tool }),
      keepalive: true,
    }).catch(function () {});
  }

  // Tab 模式（iframe）：仅通过 postMessage 触发
  if (window.self !== window.top) {
    window.addEventListener('message', function (e) {
      if (e.data && e.data.type === 'vtc_track_visit') report();
    });
    return;
  }

  // 单页面模式（直接打开）：页面加载时上报一次
  report();
})();
