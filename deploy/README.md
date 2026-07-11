# Victor Tool Collection - 部署说明

## 项目结构

```
victor-tool-collection/
├── nav/                     # 导航页
│   └── index.html
├── tools/                   # 工具目录
│   ├── score-board/         # 记分板 (React/Vite)
│   │   ├── src/
│   │   ├── dist/            # 构建输出
│   │   ├── package.json
│   │   └── vite.config.js
│   └── ...                  # 新工具放这里
├── deploy/
│   ├── nginx/               # Nginx 配置
│   │   └── port-8001.conf
│   └── README.md
```

## 添加新工具

1. **创建项目** — 在 `tools/` 下新建目录，如 `tools/my-tool/`

2. **构建产物** — 确保构建输出到 `dist/` 目录

3. **配置 nginx** — 编辑 `deploy/nginx/port-8001.conf`，添加 location 块：
   ```
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
   ```
   sudo cp deploy/nginx/port-8001.conf /etc/nginx/sites-available/port-8001
   sudo ln -sf /etc/nginx/sites-available/port-8001 /etc/nginx/sites-enabled/
   sudo nginx -t && sudo nginx -s reload
   ```

## 访问日志（Visit Logger）

所有工具页面（包括导航页）通过 nginx `sub_filter` 自动注入跟踪脚本，记录每次访问。

### 架构

```
浏览器 → nginx:8001 → 注入 track.js → 发送 POST /_log/visit → Python Logger → SQLite
```

### 组件

| 组件 | 说明 |
|------|------|
| `deploy/visit-logger.py` | Python HTTP 服务，监听 127.0.0.1:8002 |
| `data/visits.db` | SQLite 数据库，存储访问记录 |
| `/etc/systemd/system/visit-logger.service` | systemd 服务，自动重启 |
| nginx `sub_filter` | 在每个 HTML 页面的 `</body>` 前注入 `<script src="/_log/track.js">` |

### 数据库表

```sql
CREATE TABLE visits (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    ip         TEXT NOT NULL,       -- 来源 IP
    tool       TEXT NOT NULL,       -- 工具名（如 base64, typhoon）
    visited_at DATETIME DEFAULT (datetime('now', 'localtime'))
);
```

### API 接口

| 端点 | 方法 | 说明 |
|------|------|------|
| `/_log/track.js` | GET | 返回跟踪脚本（可缓存 1h） |
| `/_log/log/visit` | POST | 记录访问（Body: `{"tool":"..."}`） |
| `/_log/stats` | GET | 统计（7 天内），JSON 格式 |
| `/_log/stats?days=30` | GET | 指定天数 |
| `/_log/stats?detail=1` | GET | 展示每个工具的 IP 明细 |

### 管理服务

```bash
sudo systemctl restart visit-logger  # 重启
sudo systemctl status visit-logger   # 查看状态
sudo journalctl -u visit-logger -n 50  # 查看日志

# 直接查询数据库
sqlite3 ~/projects/Frontend/victor-tool-collection/data/visits.db \\
  "SELECT tool, COUNT(*) as visits FROM visits GROUP BY tool ORDER BY visits DESC;"
```

### 添加新工具到访问跟踪

在 nginx 的 tool location 中添加以下两行即可自动注入跟踪脚本：

```nginx
location /my-tool {
    sub_filter '</body>' '<script src="/_log/track.js"></script></body>';
    sub_filter_once off;
    ...
}
```

## 管理后台（Admin Dashboard）

密码保护的管理后台，用于查看访问统计图表。

- **位置**: `/admin/`
- **入口**: 导航页 ⚙ 设置 → 🔐 管理后台
- **技术栈**: Chart.js 4.4（CDN）
- **登录**: 默认密码 `admin888`，登录后可在页面内修改
- **密码保存**: 修改后的密码存储在浏览器 `localStorage`，清除后会重置为默认密码

## Nginx 配置

- 端口：8001
- 根目录：`nav/`（导航页）
- 各工具通过 `/tool-name/` 路径访问
- 添加新工具时记得加上 `sub_filter` 两行以启用访问跟踪
