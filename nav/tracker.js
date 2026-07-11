/**
 * VTC Tracker — 页面访问上报组件
 *
 * Tab 模式（iframe 内）：仅在切到对应 Tab 时上报（postMessage）
 * 单页面模式（直接打开）：页面加载时上报一次
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
    var data = JSON.stringify({ tool: tool });
    if (navigator.sendBeacon) {
      navigator.sendBeacon('/api/visit', data);
    } else {
      var xhr = new XMLHttpRequest();
      xhr.open('POST', '/api/visit', true);
      xhr.setRequestHeader('Content-Type', 'application/json');
      xhr.send(data);
    }
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
