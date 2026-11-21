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
│   │   ├── port-8001.conf           # 站点配置（已含安全头 include 与限流）
│   │   ├── 00-vtc-limits.conf       # limit_req_zone / map（http 上下文）
│   │   ├── 10-vtc-gzip.conf         # 压缩类型（CSS/JS/JSON/SVG）
│   │   └── vtc-security-headers.conf # 安全响应头片段（每个 location include）
│   ├── apply-p0.sh          # P0 安全加固一键安装脚本（sudo）
│   ├── deploy.sh            # 日常部署：测试 + 编译 + 同步配置 + 重启
│   ├── backup.sh            # 数据库与乐谱图片备份（cron 每日调用）
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
    sub_filter '</body>' '<script src="/tracker.js?v=6"></script></body>';
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

首次启动时自动创建数据表并写入默认管理员账号，密码以 **bcrypt 哈希**存储；初始密码可用环境变量 `VTC_INITIAL_PASSWORD` 指定，默认 `admin888`（不会再打印到日志）。**请登录管理后台后立即修改，新密码至少 8 位。**

历史遗留的明文密码会在服务启动时自动升级为 bcrypt 哈希（幂等，已升级则跳过）。

### API 接口

| 端点 | 方法 | 鉴权 | 说明 |
|------|------|------|------|
| `/api/login` | POST | 无 | 登录，设置 `vtc_session` cookie |
| `/api/logout` | POST | 无 | 退出，删除服务端 session |
| `/api/check-session` | GET | Cookie | 验证 session 是否有效 |
| `/api/visit` | POST | 无 | 记录访问（body: `{"tool":"...","device_id":"..."}`，来源 IP 取自 nginx 的 `X-Real-IP`，单 IP 限流） |
| `/api/stats` | GET | Cookie | 访问统计 |
| `/api/visits` | GET | Cookie | 分页访问记录（支持 `page`, `page_size`, `tool`, `show_label` 参数） |
| `/api/device-labels` | GET | Cookie | 列出所有设备标签 |
| `/api/device-labels` | POST | Cookie | 创建/更新设备标签（body: `{"device_id":"...","label":"..."}`） |
| `/api/device-labels/:device_id` | DELETE | Cookie | 删除设备标签 |
| `/api/change-password` | POST | Cookie | 修改管理员密码 |
| `/api/health` | GET | 无 | 健康检查 |

### 数据库表

| 表名 | 说明 |
|------|------|
| `visits` | 访问记录（id, ip, tool, user_agent, device_id, visited_at） |
| `admins` | 管理员密码（id, password（bcrypt 哈希）, created_at, updated_at） |
| `sessions` | 登录会话（id, token, created_at, expires_at） |
| `device_labels` | 设备标签（id, device_id, label, created_at, updated_at） |

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
- 公共样式通过 `/shared/` 直出仓库 `shared/` 目录（`/shared/vtc.css`）
- API 请求 `/api/` 代理到 Go Backend（127.0.0.1:8003）
- 添加新工具时记得加上 `sub_filter` 两行以启用访问跟踪

### 生效配置

实际运行的 nginx 配置文件位于 `/etc/nginx/sites-available/port-8001`，修改后需执行：

```bash
sudo nginx -t && sudo nginx -s reload
```

> 注意：仓库里的 `deploy/nginx/port-8001.conf` 是**唯一基线**，线上文件必须由它同步过去，不要反向手改线上文件，否则会再次出现两端漂移。

---

## 安全加固（P0）

一次性的安全基线，安装方式：

```bash
sudo bash deploy/apply-p0.sh
```

脚本做了什么：

| 项 | 实现 |
|---|---|
| 管理员密码 | bcrypt 哈希存储（cost 10），启动时自动把历史明文升级；改密最少 8 位，改密后失效其它会话 |
| 登录限流 | nginx `vtc_login`（10r/m，burst 5）+ 后端进程内限流（10 分钟 8 次），超限返回 429 |
| 访问上报限流 | nginx `vtc_visit`（60r/m，burst 20）+ 后端 10 秒 10 次，防止刷库写满 SD 卡 |
| 终端限流 | `vtc_term` 限制 `/webshell/term/` 握手（30r/m，burst 5） |
| API 限流 | `vtc_api` 兜底 120r/m（burst 20） |
| 会话 Cookie | `HttpOnly` + `SameSite=Lax`，仅当请求走 HTTPS（看 `X-Forwarded-Proto`）时附带 `Secure` |
| 过期会话 | 登录时与启动时清理 `sessions` 表过期记录 |
| 安全响应头 | `X-Content-Type-Options`、`X-Frame-Options: SAMEORIGIN`、`frame-ancestors 'self'`、`Referrer-Policy`、`Permissions-Policy`（相机仅允许自身）、HTTPS 下的 HSTS |
| 回环监听 | 后端改为监听 `127.0.0.1:8003`，外部只能经 nginx；排障时可用 `VTC_LISTEN=0.0.0.0:8003` 临时覆盖 |

装完后的验证清单：

```bash
# 1) 后端仍健康、且只监听回环
ss -lntp | grep 8003
curl -s http://127.0.0.1:8003/api/health

# 2) 密码已变 bcrypt
psql -d vtc -At -c "select left(password, 4) from admins;"

# 3) 登录限流生效（连续 6 次后应出现 429）
for i in $(seq 1 7); do curl -s -o /dev/null -w "%{http_code} " -X POST \
  -H 'Content-Type: application/json' -d '{"password":"wrong"}' \
  http://127.0.0.1:8001/api/login; done; echo

# 4) 安全头
curl -sI http://127.0.0.1:8001/ | grep -Ei "x-frame|x-content|referrer|frame-ancestors"
```

---

## 日常部署

```bash
bash deploy/deploy.sh                # 前后端常规发布
bash deploy/deploy.sh --scoreboard   # 顺带重建记分板（npm ci + build）
```

脚本流程：`go test` → `go build` → 按需同步 nginx 片段（有不一致才覆盖）→ `nginx -t` → reload → 重启后端。
任何一步失败即中止，不会带着半成品 reload。

修改静态工具的 `dist/index.html` 不需要跑脚本，nginx `try_files` 直接读文件；只有改 nginx 配置、后端 Go 代码、或 score-board 源码时才需要。

---

## 备份

`deploy/backup.sh` 每次产出两份文件到 `~/backups/vtc/`：

| 文件 | 内容 |
|---|---|
| `vtc-db-<时间戳>.dump` | PostgreSQL 自定义格式全库备份（`pg_restore -l` 校验通过才算成功） |
| `vtc-scores-<时间戳>.tar.gz` | `data/scores` 乐谱图片 |

默认保留最近 7 份，可用 `OUT=` / `KEEP=` 覆盖。

已安装的定时任务（`crontab -l` 可见）：

```
40 3 * * * cd /home/pi/projects/Frontend/victor-tool-collection && /bin/bash deploy/backup.sh >> /home/pi/backups/vtc/backup.log 2>&1
```

恢复示例：

```bash
# 数据库
pg_restore -h /var/run/postgresql -U vtc -d vtc --clean --if-exists ~/backups/vtc/vtc-db-YYYYmmdd-HHMMSS.dump

# 乐谱图片
tar xzf ~/backups/vtc/vtc-scores-YYYYmmdd-HHMMSS.tar.gz -C ~/projects/Frontend/victor-tool-collection/data
```

> 备份与源数据在同一块 SD 卡上，能防"误删/写坏"，但防不了整卡损坏。真正要紧的话，把 `~/backups/vtc` 定期同步到外部存储或对象存储。

---

## 压缩

`10-vtc-gzip.conf` 把 CSS / JS / JSON / SVG 纳入 gzip（系统默认只压 `text/html`）。
安装位置 `/etc/nginx/conf.d/10-vtc-gzip.conf`，由 `deploy.sh` 自动同步。验证：

```bash
curl -sI -H 'Accept-Encoding: gzip' http://127.0.0.1:8001/base64/ | grep -i content-encoding
```
