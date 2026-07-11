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

### 首次启动

首次启动时自动创建 `admins` 表并写入默认管理员密码 `admin888`。**请登录后立即修改密码。**

## API 接口

| 端点 | 方法 | 鉴权 | 说明 |
|------|------|------|------|
| `/api/login` | POST | 无 | 登录，设置 `vtc_session` cookie |
| `/api/check-session` | GET | Cookie | 验证 session 是否有效 |
| `/api/visit` | POST | 无 | 记录访问（body: `{"tool":"..."}`） |
| `/api/change-password` | POST | Cookie | 修改密码 |
| `/api/stats` | GET | Cookie | 访问统计 |
| `/api/health` | GET | 无 | 健康检查 |

## 数据库表

- **visits** — 访问记录（id, ip, tool, user_agent, visited_at）
- **admins** — 管理员密码（id, password, created_at, updated_at）
- **sessions** — 登录会话（id, token, created_at, expires_at）

## 管理

```bash
sudo systemctl start vtc-backend
sudo systemctl stop vtc-backend
sudo systemctl restart vtc-backend
sudo journalctl -u vtc-backend -n 50 -f
```
