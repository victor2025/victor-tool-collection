/**
 * VTC Tracker — 页面访问上报组件
 *
 * 规则：
 * - Tab 模式（iframe 内）：nav 发 postMessage 才上报，每次切 Tab 都报
 * - 单页面模式（直接打开）：首次获得焦点/可见时上报一次，加载不报，刷新不报
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
    return;
  }

  // ── 单页面模式（直接打开）──
  // 刷新页面不报，仅新导航时等焦点/可见才报
  try {
    var nav = performance.getEntriesByType('navigation')[0];
    if (nav && nav.type === 'reload') return; // 刷新操作：彻底不报
  } catch(e) {}

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
