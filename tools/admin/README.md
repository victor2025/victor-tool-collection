# 🔐 管理后台 · 访问统计

密码保护的管理后台，用于查看 Victor Tool Collection 的访问统计数据和图表。

## 功能

- **密码登录** — 默认密码 `admin888`（可在 `index.html` 中修改）
- **概览卡片** — 总访问量、近 7 天、独立 IP、日均访问
- **各工具排行** — 横向柱状图展示每个工具的访问次数
- **时段分布** — 24 小时折线图，观察访问高峰时段
- **IP 明细** — 每个工具下各 IP 的访问次数和最近访问时间

## 技术栈

- 纯 HTML + CSS + JavaScript
- [Chart.js 4.4](https://www.chartjs.org/)（通过 CDN 加载）
- 数据来源：`/api/stats` 接口（Python Logger）
- 无构建步骤

## 访问方式

- 导航页 ⚙ 设置 → 🔐 管理后台
- 直接访问 `/admin/`

## 修改密码

编辑 `dist/index.html`，搜索 `ADMIN_PASSWORD` 变量：

```javascript
const ADMIN_PASSWORD = 'admin888';
```

## 目录结构

```
admin/
├── README.md
└── dist/
    └── index.html
```
