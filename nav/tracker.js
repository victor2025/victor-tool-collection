/**
 * VTC Tracker — 页面访问上报组件
 * 仅在工具焦点激活时上报（postMessage、focus、visibilitychange）
 * 防抖 30s 避免重复上报
 */
(function () {
  if (window.__vtc_track_init) return;
  window.__vtc_track_init = true;

  var tool = window.location.pathname
    .replace(/\/$/, '')
    .split('/')
    .pop() || 'home';

  var lastReport = 0;
  var REPORT_DEBOUNCE = 30000;

  function reportVisit() {
    var now = Date.now();
    if (now - lastReport < REPORT_DEBOUNCE) return;
    lastReport = now;

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

  // 非 iframe 场景（直接打开工具页面）：初始化时上报一次
  if (window.self === window.top) {
    reportVisit();
  }

  // 来自导航页 Tab 切换的消息
  window.addEventListener('message', function (e) {
    if (e.data && e.data.type === 'vtc_track_visit') {
      reportVisit();
    }
  });

  // 浏览器标签页切换 / iframe 内点击获得焦点
  window.addEventListener('focus', reportVisit);

  // 浏览器标签页可见性变化
  document.addEventListener('visibilitychange', function () {
    if (!document.hidden) reportVisit();
  });
})();
