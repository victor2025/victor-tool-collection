# 记分板 🏀

> 一个简洁的记分板应用，支持可编辑队名、翻页动画、本地持久化存储。

## 技术栈

- **框架**: React 18
- **构建工具**: Vite 5
- **无外部状态管理**（使用 React 内置 useState + localStorage）

## 功能

- 两队分数实时显示（0~999）
- 翻页动画数字显示（FlipDigit 组件）
- 可编辑主队/客队队名
- 加分（+1/+2/+3）和减分操作
- 一键重置所有分数
- 自动保存到 `localStorage`，刷新页面不丢数据

## 访问

通过 nginx 代理：`http://localhost:8001/score-board/`

## 目录结构

```
tools/score-board/
├── README.md
├── index.html             # Vite 入口 HTML
├── package.json           # 依赖配置
├── vite.config.js         # Vite 构建配置
├── src/
│   ├── main.jsx           # React 入口
│   ├── App.jsx            # 主组件（状态管理）
│   ├── App.css            # 主样式
│   └── components/
│       ├── ScoreBoard.jsx # 记分板布局组件
│       ├── ScoreBoard.css
│       ├── FlipDigit.jsx  # 翻页动画数字组件
│       └── FlipDigit.css
├── public/                # 静态资源
└── dist/                  # 构建产物（nginx 服务目录）
```

## 本地开发

```bash
cd tools/score-board
npm install          # 首次安装依赖
npm run dev          # 启动 Vite 开发服务器（热更新）
npm run build        # 构建到 dist/
npm run preview      # 预览构建产物
```

## 部署

构建后需确保 nginx 配置已添加对应 location：

```nginx
location /score-board {
    alias /home/pi/projects/Frontend/victor-tool-collection/tools/score-board/dist;
    try_files $uri $uri/ /score-board/index.html;
}
```

修改后需更新 `nav/index.html`（如需）和 `deploy/nginx/port-8001.conf`（如需）。
