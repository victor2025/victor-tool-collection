/**
 * VTC Tracker — 页面访问上报组件
 *
 * 规则：
 * - Tab 模式（iframe 内）：nav 发 postMessage 才上报，每次切 Tab 都报
 * - 单页面模式（直接打开）：首次获得焦点/可见时上报一次，加载不报
 * - 其他情况一律不上报
 */
(function () {
  if (window.__vtc_track_init) return;
  window.__vtc_track_init = true;

  var tool = window.location.pathname
    .replace(/\/$/, '')
    .split('/')
    .pop() || 'home';

  function report() {
    // 非 iframe 模式下：页面级锁，只发一次
    if (window.__vtc_report_sent && window.self === window.top) return;
    window.__vtc_report_sent = true;

    fetch('/api/visit', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ tool: tool }),
      keepalive: true,
    }).catch(function () {});
  }

  // ── Tab 模式（iframe）──
  if (window.self !== window.top) {
    window.addEventListener('message', function (e) {
      if (e.data && e.data.type === 'vtc_track_visit') report();
    });
    return; // iframe 不执行下面的逻辑
  }

  // ── 单页面模式（直接打开）──
  // 页面加载时不报，等用户切到该页面/获得焦点才报
  function onActive() {
    report();
    window.removeEventListener('focus', onActive);
    document.removeEventListener('visibilitychange', onVisible);
  }
  function onVisible() {
    if (!document.hidden) onActive();
  }
  window.addEventListener('focus', onActive);
  document.addEventListener('visibilitychange', onVisible);
})();
