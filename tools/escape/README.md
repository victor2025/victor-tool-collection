# 转义/反转义工具 🔤

> 支持 HTML 实体、JavaScript 字符串、JSON 字符串三类特殊字符的转义与反转义。

## 技术栈

- 纯静态 HTML + CSS + JavaScript（零外部依赖）
- 仅在浏览器端运行，无后端需求

## 功能

- **格式支持**: 三种转义格式一键切换
  - `HTML 实体`: `&` `<` `>` `"` `'` → `&amp;` `&lt;` `&gt;` `&quot;` `&#39;`
  - `JavaScript 字符串`: `\n` `\t` `\"` `\'` `\\` `\uXXXX` 等
  - `JSON 字符串`: 基于 `JSON.stringify` 规则，保证输出合法 JSON
- **转义**: 原文 → 转义后的字符串
- **反转义**: 转义后的字符串 → 原文（支持 `\xHH`、`\uHHHH`、`\u{...}` 等序列）
- **非 ASCII 转义**: 可选将中文/emoji 等转义为 `\uXXXX`（HTML 模式下自动隐藏）
- **清空/复制**: 每个文本框右上角均带「清空」「复制」按钮

## 访问

通过 nginx 代理：`http://localhost:8001/escape/`

## 目录结构

```
tools/escape/
├── README.md
└── dist/
    └── index.html     # 构建产物（单文件应用）
```

## 开发与部署

该工具为单 HTML 文件应用，直接编辑 `dist/index.html` 后刷新即可。

修改后需更新 `nav/index.html`（如需）和 `deploy/nginx/port-8001.conf`（如需）。
