# Victor's Tool Collection 🧰

> 一个轻量级的 Web 工具集，运行在树莓派上，通过 nginx 提供便捷的开发工具和娱乐工具。

## 快速访问

- **地址**: [https://tools.victor2022.dpdns.org](https://tools.victor2022.dpdns.org)
- **导航页**: `nav/index.html`（根目录）

---

## 🛠 工具一览

### 开发工具集

| 工具 | 路径 | 技术栈 | 功能 |
|------|------|--------|------|
| 🔣 **Base64 编解码** | `/base64/` | 纯静态 HTML/CSS/JS | UTF-8 Base64 双向同步编解码，一键复制 |
| 📱 **二维码工具** | `/qrcode/` | 纯静态 HTML/CSS/JS + 第三方库 | 生成二维码、从图片扫码、调用摄像头扫码 |
| 🔐 **JWT 解码** | `/jwt-decoder/` | 纯静态 HTML/CSS/JS | 解析 JWT Header/Payload，识别注册声明，过期校验 |
| 🖥 **WebShell 终端** | `/webshell/` | ttyd + webshell-wrapper + su/SSH | 本地终端输入系统密码 · SSH 连接使用 SSH 账密 |
| ⏰ **时间戳转换** | `/timestamp/` | 纯静态 HTML/CSS/JS | Unix 时间戳 · 秒/毫秒 · 日期 ↔ 时间戳双向转换 |
| 🔐 **管理后台** | `/admin/` | Chart.js + Go API | 密码登录 · 访问统计图表 · IP 明细 · 分页访问记录 |

### 娱乐工具集

| 工具 | 路径 | 技术栈 | 功能 |
|------|------|--------|------|
| 🏀 **记分板** | `/score-board/` | React + Vite | 可编辑队名、翻页动画显示、自动本地持久化 |

### 生活工具集

| 工具 | 路径 | 技术栈 | 功能 |
|------|------|--------|------|
| 🌪 **台风观测** | `/typhoon/` | 纯静态 HTML/CSS/JS（iframe 嵌入） | 中央气象台实时台风路径观测 |

---

## 📁 项目结构

```
victor-tool-collection/
├── AGENTS.md                # 项目规范（面向 AI Agent）
├── README.md                ← 本文件
├── nav/
│   ├── index.html           # 导航首页（端口 8001 根目录）
│   ├── tracker.js           # 访问跟踪脚本（上报 /api/visit）
│   ├── sw.js                # Service Worker
│   └── manifest.json        # PWA 清单
├── tools/
│   ├── admin/               # 管理后台（访问统计仪表盘）
│   ├── backend/             # Go 后端服务（Gin + GORM）
│   ├── base64/              # Base64 编解码工具
│   ├── jwt-decoder/         # JWT 解码工具
│   ├── json-formatter/      # JSON 格式化工具
│   ├── qrcode/              # 二维码工具
│   ├── score-board/         # 记分板（React）
│   ├── timestamp/           # 时间戳转换工具
│   ├── typhoon/             # 台风观测工具
│   └── webshell/            # Web 终端（ttyd）
└── deploy/
    ├── README.md             # 部署说明
    └── nginx/
        └── port-8001.conf    # nginx 站点配置（端口 8001）
```

---

## 🏗 如何添加新工具

详见 [`deploy/README.md`](deploy/README.md) 的完整流程，简要步骤：

1. 在 `tools/` 下创建新目录，构建产物输出到 `dist/`
2. 添加该模块的 `README.md`
3. 更新根目录 `README.md` 的工具列表
4. 更新 `nav/index.html` 添加入口卡片
5. 更新 `deploy/nginx/port-8001.conf` 添加 nginx location
6. 重载 nginx 使配置生效

---

## ⚙️ 技术栈

- **运行环境**: 树莓派 (Raspberry Pi) + Nginx
- **后端服务**: Go + Gin + GORM（PostgreSQL）
- **前端框架**: React 18 + Vite（记分板）
- **图表**: Chart.js 4（管理后台）
- **纯静态工具**: 零依赖 HTML/CSS/JS
- **终端服务**: ttyd（Web SSH 终端）
- **访问跟踪**: 前端 tracker.js → Go API → PostgreSQL
- **会话认证**: 服务端 session（DB 存储）+ cookie
- **端口**: 8001（HTTP）- 8003（Go Backend API）

---

## 🔧 本地开发

```bash
# 记分板开发
cd tools/score-board
npm install
npm run dev      # 本地开发服务器
npm run build    # 构建到 dist/

# 静态工具（base64 / qrcode / webshell）
# 直接在 dist/index.html 中编辑
```

---

## 📝 维护说明

- 每次添加/修改工具后，必须同步更新 `README.md`
- 每个工具模块都应有自己的 `README.md`
- 详见 `AGENTS.md` 的完整变更规范
