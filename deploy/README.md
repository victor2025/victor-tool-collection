# Victor Tool Collection - 部署说明

## 项目结构

```
victor-tool-collection/
├── nav/                     # 导航页
│   └── index.html
├── tools/                   # 工具目录
│   ├── admin/               # 管理后台（访问统计仪表盘）
│   ├── backend/             # Go 后端服务
│   ├── score-board/         # 记分板 (React/Vite)
│   └── ...                  # 其他工具
├── deploy/
│   ├── nginx/               # Nginx 配置
│   │   └── port-8001.conf
│   └── README.md
```

## 添加新工具

1. **创建项目** — 在 `tools/` 下新建目录，如 `tools/my-tool/`

2. **构建产物** — 确保构建输出到 `dist/` 目录

3. **配置 nginx** — 编辑 `deploy/nginx/port-8001.conf`，添加 location 块：
   ```nginx
   location /my-tool {
       alias /home/pi/projects/Frontend/victor-tool-collection/tools/my-tool/dist;
       try_files $uri $uri/ /my-tool/index.html;
   }
   ```

4. **更新导航页** — 在 `nav/index.html` 的 `.tool-list` 中添加卡片：
   ```html
   <a href="/my-tool/" class="tool-card">
     <div class="tool-icon">🎯</div>
     <div class="tool-info">
       <div class="tool-name">工具名</div>
       <div class="tool-desc">描述</div>
       <div class="tool-tags"><span class="tool-tag">React</span></div>
     </div>
     <span class="tool-arrow">→</span>
   </a>
   ```

5. **重载 nginx**：
   ```bash
   sudo cp deploy/nginx/port-8001.conf /etc/nginx/sites-available/port-8001
   sudo ln -sf /etc/nginx/sites-available/port-8001 /etc/nginx/sites-enabled/
   sudo nginx -t && sudo nginx -s reload
   ```

## 访问跟踪

所有工具页面（包括导航页）通过 `nav/tracker.js` 自动上报访问记录到 Go 后端。

### 架构

```
浏览器打开工具页面 → tracker.js → POST /api/visit → nginx → Go Backend (8003) → PostgreSQL
```

### 组件

| 组件 | 说明 |
|------|------|
| `nav/tracker.js` | 前端上报脚本，自动捕获页面路径作为工具名 |
| `tools/backend/` | Go + Gin + GORM 后端服务，监听 127.0.0.1:8003 |
| PostgreSQL | 存储访问记录、管理员密码、登录会话 |

### 跟踪脚本自动注入

`nav/tracker.js` 通过 nginx `sub_filter` 自动注入到所有 HTML 页面：

```nginx
location /my-tool {
    sub_filter '</body>' '<script src="/tracker.js"></script></body>';
    sub_filter_once off;
    ...
}
```

---

## Go Backend 服务

Go 后端提供认证鉴权、访问统计、数据查询等功能。

### 配置文件

启动时读取 `tools/backend/config.json`，**不提交到 git**（已在 `.gitignore` 中）。

```json
{
  "db_type": "postgres",
  "dsn": "host=/var/run/postgresql user=vtc password=*** dbname=vtc sslmode=disable",
  "server_port": "8003"
}
```

| 字段 | 必填 | 默认值 | 说明 |
|------|------|--------|------|
| `db_type` | 否 | `postgres` | 数据库类型，支持 `postgres` / `sqlite` / `mysql` |
| `dsn` | 是 | — | 数据库连接字符串 |
| `server_port` | 否 | `8003` | 监听端口 |

### 首次启动

首次启动时自动创建数据表并写入默认管理员密码 `admin888`。**请登录管理后台后立即修改密码。**

### API 接口

| 端点 | 方法 | 鉴权 | 说明 |
|------|------|------|------|
| `/api/login` | POST | 无 | 登录，设置 `vtc_session` cookie |
| `/api/logout` | POST | 无 | 退出，删除服务端 session |
| `/api/check-session` | GET | Cookie | 验证 session 是否有效 |
| `/api/visit` | POST | 无 | 记录访问（body: `{"tool":"..."}`） |
| `/api/stats` | GET | Cookie | 访问统计 |
| `/api/visits` | GET | Cookie | 分页访问记录（支持 `page`, `page_size`, `tool` 参数） |
| `/api/change-password` | POST | Cookie | 修改管理员密码 |
| `/api/health` | GET | 无 | 健康检查 |

### 数据库表

| 表名 | 说明 |
|------|------|
| `visits` | 访问记录（id, ip, tool, user_agent, visited_at） |
| `admins` | 管理员密码（id, password, created_at, updated_at） |
| `sessions` | 登录会话（id, token, created_at, expires_at） |

### 服务管理

```bash
# 编译
cd tools/backend && GO111MODULE=on go build -o backend .

# 服务控制（systemd）
sudo systemctl start vtc-backend       # 启动
sudo systemctl stop vtc-backend        # 停止
sudo systemctl restart vtc-backend     # 重启
sudo systemctl status vtc-backend      # 查看状态
sudo journalctl -u vtc-backend -n 50 -f   # 查看实时日志

# 直接查询数据库
psql -d vtc -c "SELECT tool, COUNT(*) AS visits FROM visits GROUP BY tool ORDER BY visits DESC;"
```

---

## 管理后台

密码保护的管理后台，用于查看访问统计、IP 明细和分页访问记录。

- **位置**: `/admin/`
- **入口**: 导航页 ⚙ 设置 → 🔐 管理后台
- **技术栈**: Chart.js 4.4（CDN）+ Go API
- **认证**: 服务端 session（存储于数据库），cookie 传递
- **登录**: 默认密码 `admin888`，登录后可在页面内修改
- **功能**:
  - 概览卡片（总访问、独立 IP、日均访问）
  - 各工具访问排行（柱状图）
  - 时段分布（折线图）
  - IP 明细表（按 IP + 工具分组）
  - 分页访问记录（支持工具筛选、页码跳转）
  - 密码修改

---

## Nginx 配置

- 端口：8001
- 根目录：`nav/`（导航页）
- 各工具通过 `/tool-name/` 路径访问
- API 请求 `/api/` 代理到 Go Backend（127.0.0.1:8003）
- 添加新工具时记得加上 `sub_filter` 两行以启用访问跟踪

### 生效配置

实际运行的 nginx 配置文件位于 `/etc/nginx/sites-available/port-8001`，修改后需执行：

```bash
sudo nginx -t && sudo nginx -s reload
```
