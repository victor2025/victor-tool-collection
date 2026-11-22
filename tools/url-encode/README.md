# URL 编解码

一个纯静态的 URL 编解码小工具，输入即转换，适合快速处理百分号编码文本。

## 功能

- **双向同步转换**：左侧输入原文自动生成 URL 编码，右侧输入编码串自动还原原文（`oninput` 实时触发）
- **两种编码模式**：
  - 完整模式（勾选「完整编码」）→ `encodeURIComponent`，连 `: / ? & =` 等保留字符一并编码
  - 常规模式 → `encodeURI`，保留 URL 结构字符，适合整条链接
- **中文与 Unicode**：按标准百分号编码处理，避免出现 `%uXXXX` 之类的非标准结果
- **一键操作**：每个文本框右上角有「清空」「复制」（遵循 AGENTS.md 的 UI 规范），复制失败时回退到 `document.execCommand`
- **轻提示**：操作结果通过页面底部 toast 反馈

## 技术栈

- 纯静态 HTML / CSS / JavaScript，零第三方依赖
- 样式取自 `/shared/vtc.css` 的 `--vtc-*` 设计变量（暗/亮两套真实色板，不再依赖整页反相滤镜），是本项目第一个完成共享样式迁移的页面
- 单文件实现：`dist/index.html`

## 目录结构

```
tools/url-encode/
├── README.md        ← 本文件
└── dist/
    └── index.html   ← 工具本体（nginx 直接静态托管）
```

## 主题适配

- `<head>` 引入 `/shared/vtc.css`，颜色全部用 `var(--vtc-*)`
- `<body>` 带 `vtc-native` 类，与全站其它工具页共用同一套变量主题
- 主题切换仍由 `/theme.js` + 导航页的 postMessage 驱动，无需页面自己监听

## 构建与部署

无需构建，直接编辑 `dist/index.html` 即可。

部署方式（已配置在 `deploy/nginx/port-8001.conf`）：

```nginx
location /url-encode {
    sub_filter '</body>' '<script src="/tracker.js?v=6"></script></body>';
    sub_filter_once off;
    alias /home/pi/projects/Frontend/victor-tool-collection/tools/url-encode/dist;
    try_files $uri $uri/ /url-encode/index.html;
    expires 1h;
    add_header Cache-Control "public";
}
```

访问地址：`https://tools.victor2022.dpdns.org/url-encode/`
