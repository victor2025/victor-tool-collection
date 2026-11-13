# Go Backend — Victor Tool Collection

Gin + GORM 后端服务，提供访问记录、认证鉴权、统计查询功能。

## 配置文件

启动时读取同目录下的 `config.json`，**不提交到 git**。

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
| `scores_dir` | 否 | `<repo>/data/scores` | 乐谱文件存储目录（相对路径按可执行文件所在目录解析） |

### 首次启动

首次启动时自动创建 `admins` 表并写入默认管理员密码 `admin888`。**请登录后立即修改密码。**

## API 接口

| 端点 | 方法 | 鉴权 | 说明 |
|------|------|------|------|
| `/api/login` | POST | 无 | 登录，设置 `vtc_session` cookie |
| `/api/logout` | POST | 无 | 退出，删除服务端 session |
| `/api/check-session` | GET | Cookie | 验证 session 是否有效 |
| `/api/visit` | POST | 无 | 记录访问（body: `{"tool":"...", "device_id":"..."}`） |
| `/api/change-password` | POST | Cookie | 修改密码 |
| `/api/stats` | GET | Cookie | 访问统计 |
| `/api/visits` | GET | Cookie | 分页访问记录（支持 `page`, `page_size`, `tool`, `show_label`） |
| `/api/device-labels` | GET | Cookie | 列出所有设备标签 |
| `/api/device-labels` | POST | Cookie | 创建/更新设备标签（body: `{"device_id":"...","label":"..."}`） |
| `/api/device-labels/:device_id` | DELETE | Cookie | 删除设备标签 |
| `/api/scores` | GET | 无 | 乐谱列表 / 搜索（`?q=关键词&type=类型`），含 `images[]` |
| `/api/scores/types` | GET | 无 | 已使用的乐谱类型（去重） |
| `/api/scores` | POST | Cookie | 新建乐谱（JSON：`title, composer, type, note, images[]`） |
| `/api/scores/:id` | PATCH | Cookie | 修改乐谱信息 |
| `/api/scores/:id` | DELETE | Cookie | 删除乐谱 + 图片行 + 本地文件 |
| `/api/scores/:id/images` | POST | Cookie | 追加图片（multipart `files` 或 JSON `urls`） |
| `/api/scores/images/:imageId` | DELETE | Cookie | 删除单张图片 |
| `/api/system` | GET | Cookie | 服务器实时状态（CPU/内存/swap/温度/电压/频率/负载/磁盘/运行时长，树莓派 vcgencmd） |
| `/api/health` | GET | 无 | 健康检查 |

## 数据库表

- **visits** — 访问记录（id, ip, tool, user_agent, device_id, visited_at）
- **admins** — 管理员密码（id, password, created_at, updated_at）
- **sessions** — 登录会话（id, token, created_at, expires_at）
- **device_labels** — 设备标签（id, device_id, label, created_at, updated_at）
- **scores** — 乐谱（id, title, composer, type, note, created_at）
- **score_images** — 乐谱页面图片（id, score_id, url, source, source_url, stored_name, sort_order, created_at）

## 管理

```bash
sudo systemctl start vtc-backend
sudo systemctl stop vtc-backend
sudo systemctl restart vtc-backend
sudo journalctl -u vtc-backend -n 50 -f
```
