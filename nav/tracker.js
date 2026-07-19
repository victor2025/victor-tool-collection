// VTC Tracker — 访问统计 + 设备标识
(function() {
  'use strict';

  // 1. 设备唯一标识符（有则复用，无则生成）
  function getDeviceId() {
    var id = localStorage.getItem('vtc_device_id');
    if (id) return id;
    // 生成 UUID v4
    id = 'xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx'.replace(/[xy]/g, function(c) {
      var r = Math.random() * 16 | 0;
      return (c === 'x' ? r : (r & 0x3 | 0x8)).toString(16);
    });
    localStorage.setItem('vtc_device_id', id);
    return id;
  }

  var deviceId = getDeviceId();

  // 2. 上报访问记录
  var tool = location.pathname.replace(/\/$/,"").split("/").pop() || "home";
  // 延迟上报确保页面加载优先
  setTimeout(function() {
    fetch("/api/visit", {
      method: "POST",
      body: JSON.stringify({ tool: tool, device_id: deviceId }),
      headers: { "Content-Type": "application/json" }
    }).catch(function() {});
  }, 100);
})();
