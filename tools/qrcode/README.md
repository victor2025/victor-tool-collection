# 二维码工具 📱

> 一站式二维码工具：生成二维码、从图片解码、调用摄像头扫码。

## 技术栈

- 纯静态 HTML + CSS + JavaScript
- 依赖第三方库：
  - `qrcode.min.js` — 用于生成二维码
  - `jsQR.js` — 用于从图片/摄像头解码二维码

## 功能

- **生成二维码**: 输入文本/URL，实时生成二维码
- **图片扫码**: 上传包含二维码的图片，自动解码
- **摄像头扫码**: 调用浏览器摄像头实时扫描二维码

## 访问

通过 nginx 代理：`http://localhost:8001/qrcode/`

## 目录结构

```
tools/qrcode/
├── README.md
└── dist/
    ├── index.html        # 主应用页面
    └── lib/
        ├── qrcode.min.js # 二维码生成库
        └── jsQR.js       # 二维码解码库
```

## 开发与部署

编辑 `dist/index.html` 后刷新即可。如需更新第三方库，替换 `dist/lib/` 下的文件。

修改后需更新 `nav/index.html`（如需）和 `deploy/nginx/port-8001.conf`（如需）。
