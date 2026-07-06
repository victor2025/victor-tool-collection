# Base64 编解码工具 🔣

> 轻量级的 UTF-8 Base64 编解码工具，支持双向同步，一键复制。

## 技术栈

- 纯静态 HTML + CSS + JavaScript（零外部依赖）
- 仅在浏览器端运行，无后端需求

## 功能

- **编码**: 将 UTF-8 文本编码为 Base64
- **解码**: 将 Base64 解码回 UTF-8 文本
- **双向同步**: 支持自动双向同步（例如输入编码或解码任一侧自动联动）
- **一键复制**: 点击复制按钮直接复制结果到剪贴板

## 访问

通过 nginx 代理：`http://localhost:8001/base64/`

## 目录结构

```
tools/base64/
├── README.md
└── dist/
    └── index.html     # 构建产物（单文件应用）
```

## 开发与部署

该工具为单 HTML 文件应用，直接编辑 `dist/index.html` 后刷新即可。

修改后需更新 `nav/index.html`（如需）和 `deploy/nginx/port-8001.conf`（如需）。
