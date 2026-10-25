/**
 * VTC 主题引导脚本
 * 读取 localStorage.vtc_theme（dark | light）并同步到 <html data-theme="...">。
 * same-origin 的其它页面/iframe 切换主题时，通过 storage 事件或 postMessage 实时跟随。
 */
(function () {
  function normalize(value) {
    return value === 'light' ? 'light' : 'dark';
  }

  function readTheme() {
    try {
      return normalize(localStorage.getItem('vtc_theme'));
    } catch (e) {
      return 'dark';
    }
  }

  function apply(theme, notify) {
    document.documentElement.setAttribute('data-theme', theme);
    if (notify) {
      window.dispatchEvent(new CustomEvent('vtc-theme-change', { detail: { theme: theme } }));
    }
  }

  apply(readTheme(), false);

  window.addEventListener('storage', function (e) {
    if (e.key === 'vtc_theme') apply(normalize(e.newValue), true);
  });

  window.addEventListener('message', function (e) {
    var data = e.data;
    if (data && data.type === 'vtc-theme') apply(normalize(data.theme), true);
  });

  window.vtcSetTheme = function (theme) {
    var next = normalize(theme);
    apply(next, true);
    try { localStorage.setItem('vtc_theme', next); } catch (e) {}
  };
})();
