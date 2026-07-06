# WebShell - Web 终端工具

基于 [ttyd](https://github.com/tsl0922/ttyd) 的 Web SSH 终端。

## 架构

```
用户浏览器 → nginx (8001) → ttyd (8022) → /bin/bash
```

- nginx 代理 `/webshell/term/` 到 ttyd WebSocket
- ttyd 以 systemd user service 运行，持久化保活
- 支持 SSH 连接和本地终端

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

- **本地终端**：访问 `/webshell/term/`，直接进入树莓派的 shell
- **SSH 连接**：在 `/webshell/` 页面填写主机/端口/用户名，自动打开 SSH 连接
- 最近连接记录存储在浏览器 localStorage 中
