# Go Backend — Victor Tool Collection

Gin + GORM 后端服务，替代原 Python 日志服务。

## API 接口

| 端点 | 方法 | 鉴权 | 说明 |
|------|------|------|------|
| `/api/login` | POST | 无 | 登录，返回 token |
| `/api/check-session` | POST | 无 | 验证 token 是否有效 |
| `/api/visit` | POST | 无 | 记录访问（body: `{"tool":"..."}`） |
| `/api/change-password` | POST | Bearer token | 修改密码 |
| `/api/stats` | GET | Bearer token | 访问统计 |
| `/api/health` | GET | 无 | 健康检查 |

## 数据库

支持 PostgreSQL（默认），可通过环境变量切换到 SQLite 或 MySQL。

```bash
# 当前配置
DB_TYPE=postgres
DB_DSN=host=/var/run/postgresql user=vtc password=*** dbname=vtc sslmode=disable

# 切换到 SQLite
DB_TYPE=sqlite DSN=backend.db

# 切换到 MySQL（需安装 mysql driver）
DB_TYPE=mysql DSN=user:pass@tcp(host:3306)/dbname
```

## 环境变量

| 变量 | 默认值 | 说明 |
|------|--------|------|
| `DB_TYPE` | postgres | 数据库类型 |
| `DSN` | (postgres 连接串) | 连接字符串 |
| `SERVER_PORT` | 8003 | 监听端口 |
| `ADMIN_PASSWORD` | mima123123 | 管理员密码 |

## 管理

```bash
# 启动/停止/重启
sudo systemctl start vtc-backend
sudo systemctl stop vtc-backend
sudo systemctl restart vtc-backend

# 查看日志
sudo journalctl -u vtc-backend -n 50 -f
```
