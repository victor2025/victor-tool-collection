/**
 * VTC Tracker — 页面访问上报组件
 * 非阻塞上报，不影响页面加载
 */
(function () {
  if (window.__vtc_tracked) return;
  window.__vtc_tracked = true;

  var tool = window.location.pathname
    .replace(/\/$/, '')
    .split('/')
    .pop() || 'home';

  var data = JSON.stringify({ tool: tool });

  if (navigator.sendBeacon) {
    navigator.sendBeacon('/api/visit', data);
  } else {
    var xhr = new XMLHttpRequest();
    xhr.open('POST', '/api/visit', true);
    xhr.setRequestHeader('Content-Type', 'application/json');
    xhr.send(data);
  }
})();
