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

## Nginx 配置

- 端口：8001
- 根目录：`nav/`（导航页）
- 各工具通过 `/tool-name/` 路径访问
