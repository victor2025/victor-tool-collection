# WebShell - Web 终端工具

基于 [ttyd](https://github.com/tsl0922/ttyd) 的 Web 终端，支持本地终端（系统账密鉴权）和远程 SSH 连接。

## 架构

```
┌─ 本地终端 ───────────────────────────────────────────────┐
│ 浏览器 → nginx (:8001) → ttyd (:8022) → webshell-wrapper  │
│                           (loopback)          ↓ 无参数 ↓  │
│                                            su - pi → 输入密码 │
│                                                      → shell │
└─────────────────────────────────────────────────────────┘

┌─ SSH 连接 ───────────────────────────────────────────────┐
│ 浏览器 → nginx (:8001) → ttyd (:8022) → webshell-wrapper  │
│                           (loopback)     ↓ 有参数 ↓       │
│                              ssh -p 22 user@host → SSH 账密  │
│                                                     → shell │
└──────────────────────────────────────────────────────────┘

## 鉴权原理

WebShell 的鉴权不依赖 nginx，而是由目标命令自身完成：

| 模式 | 执行流程 | 鉴权方式 |
|------|---------|---------|
| 本地终端 | `webshell-wrapper` (无参数) → `su - pi` | 输入系统密码 |
| SSH 连接 | `webshell-wrapper` (有参数) → `ssh -p 22 user@host` | SSH 密码/密钥 |

> 使用 `/usr/local/bin/webshell-wrapper` 作为 ttyd 入口命令，避免 `-a` 标志传递 URL 参数给 `su` 时导致 bash 误将 `ssh` 作为脚本加载的 bug。
```

- nginx 纯代理，不参与鉴权
- 终端鉴权由目标程序负责：`su -l pi`（本地）或 `ssh`（远程）
- ttyd 只监听 `127.0.0.1:8022`（loopback），外部无法直接访问

## 文件说明

- `static/index.html` — WebShell 连接页面（源码）
- `dist/index.html` — 构建产物（拷贝自 static）

## 部署

```bash
# ttyd 服务管理
systemctl --user start ttyd-webshell.service   # 启动
systemctl --user stop ttyd-webshell.service    # 停止
systemctl --user status ttyd-webshell.service  # 状态
systemctl --user enable ttyd-webshell.service  # 开机自启
journalctl --user -u ttyd-webshell.service -f  # 查看日志
```

## 使用

### 🔌 本地终端

点击「本地终端」按钮，新标签页打开 ttyd，运行 `su -l pi`。终端中会提示输入系统密码，验证通过即进入 shell。

> 💡 按 `Ctrl+D` 或输入 `exit` 退出终端

### 🔗 SSH 远程连接

在表单中填写主机地址、端口、用户名，点击「连接 SSH」。ttyd 以 `ssh -p <port> <user>@<host>` 启动，SSH 协议自身处理认证（密码或密钥）。

### 📋 最近连接

最近连接的 SSH 记录保存在浏览器 localStorage 中，点击可快速重连，点击 ✕ 可删除。

## 变更密码

本地终端鉴权直接对接系统 PAM：

```bash
passwd pi  # 修改 pi 用户密码后，新密码即时生效
```

## ttyd 启动参数

```
-t fontSize=14                      # 终端字号
-t themeBackground=#0d1117           # 暗色背景
-t themeForeground=#c9d1d9           # 亮色文字
-t themeCursor=#4ecdc4               # 青色光标
-W                                   # 允许写入（可交互）
-i 127.0.0.1                         # 仅 loopback，安全
-a                                   # 支持 URL 参数
/usr/local/bin/webshell-wrapper       # 入口包装脚本
  └─ 无参数 → su - pi (系统密码)      # 本地终端
  └─ 有参数 → exec "$@"              # SSH 连接
```

## 包装脚本说明

`/usr/local/bin/webshell-wrapper` 解决了两个关键问题：

### 1️⃣ 避免 `-a` 参数传递 bug

带 `-a` 时，URL 参数追加到默认命令末尾。如果默认命令是 `su - pi`，
SSH 连接的 URL 参数 `?arg=ssh&arg=-p&arg=22&arg=user@host` 会导致：

```
su - pi ssh -p 22 user@host
               ↑ bash 把 ssh 当作脚本文件加载 → 报错！
```

包装脚本自动判断：无参数 → 走 `su` 鉴权；有参数 → 直接 `exec` 执行参数。

### 2️⃣ 强制 SSH 密码认证

本机已配置 SSH 密钥（`~/.ssh/id_rsa`），直接用 `ssh user@host` 会
无密码登录。包装脚本检测到 SSH 命令时，自动添加：

```
-o PubkeyAuthentication=no
-o PreferredAuthentications=keyboard-interactive,password
```

确保即使在配置了密钥的机器上，通过 WebShell 连接也必须输入密码。
