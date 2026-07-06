# JWT 解码工具 🔐

> 浏览器端解析 JSON Web Token，解码 Header 和 Payload，展示注册声明。

## 技术栈

- 纯静态 HTML + CSS + JavaScript（零外部依赖）
- 所有解析在浏览器本地完成，Token 不会发送到服务端

## 功能

- **JWT 解码**: 输入三段式 JWT Token，自动解析
- **Header 展示**: 显示算法（alg）和类型（typ）等头部信息
- **Payload 展示**: 格式化显示所有声明
- **注册声明解析**: 自动识别并展示标准注册声明：
  - `iss` — 签发者
  - `sub` — 主题
  - `aud` — 受众
  - `exp` — 过期时间（含中文时间显示 + 过期状态标识）
  - `nbf` — 生效时间
  - `iat` — 签发时间
  - `jti` — JWT ID
- **一键复制**: 复制 Header 或 Payload 的 JSON 内容
- **Token 段标记**: 实时显示各段字符数

## 访问

通过 nginx 代理：`http://localhost:8001/jwt-decoder/`

## 目录结构

```
tools/jwt-decoder/
├── README.md
└── dist/
    └── index.html     # 构建产物（单文件应用）
```

## 开发与部署

直接编辑 `dist/index.html` 后刷新即可。

修改后需更新 `nav/index.html`（如需）和 `deploy/nginx/port-8001.conf`（如需）。
