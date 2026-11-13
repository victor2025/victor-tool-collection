# 乐谱夹 (Score Library)

个人乐谱收藏工具：一份乐谱对应多张图片，支持本地上传或公网图片链接导入，按名称 / 作曲家搜索，点开可整页翻看。

## 功能

- **多图乐谱**：一份乐谱 = N 张图片（数据库 `scores` + `score_images` 两张表，一对多）
- **两种导入方式**
  - 上传图片：可多选 / 拖拽，整批入库
  - 图片链接：每行一个 URL；`http://` 自动镜像到本地，`https://` 保留外链
- **乐谱类型**：吉他谱 / 钢琴谱 / 简谱 / 五线谱 / 其他，可按类型筛选（类型存在数据库 `scores.type`）
- **搜索**：名称 / 作曲家 / 备注 模糊搜索，输入即搜（220ms 防抖）
- **翻页预览**：全屏查看，左右翻页 + 缩略图条 + 键盘 ←/→，支持「原图」新窗口打开
- **鉴权**：**查看公开**（列表 / 搜索 / 类型 / 图片无需登录）；新增、修改、删除需要管理密码（`vtc_session` cookie）

## 技术栈

- 前端：纯静态单文件 `dist/index.html`（无构建）
- 后端：Gin + GORM（`tools/backend/handlers/scores.go`）
- 文件服务：nginx 静态直出 `/scores/files/`
- 存储：图片在 `<repo>/data/scores/`，元数据在 Postgres

## 为什么 http 图片要镜像

源站（如 keepgoing.cn）有 Referer 防盗链：带其他站点 Referer 返回 403，无 Referer 才 200。
同时该站 443 端口不通，https 页面直接引用 http 图片会被浏览器按混合内容拦截。
因此链接导入时：**http → 下载到本地存 `/scores/files/`**（记录 `source_url` 保留出处），**https → 直接引用**。

## 目录结构

```
tools/score-library/
├── README.md
└── dist/
    └── index.html      # 单文件前端
```

## 数据库

```
scores                               score_images
------                               ------------
id, title, composer,                 id, score_id, url,
type, note, created_at               source(local|remote), source_url,
                                     stored_name, sort_order, created_at
```

- `score_images.url` 是前端真正加载的地址：本地为 `/scores/files/xxx.png`，外链为原始 URL
- `source_url` 保留原始出处，便于回溯
- 删除乐谱会级联删除图片行，并删除对应本地文件

## 接口（读取公开，写入需登录 cookie）

| 端点 | 方法 | 说明 |
|------|------|------|
| `/api/scores?q=&type=` | GET | **公开** · 列表 / 搜索，返回每份乐谱含 `images[]`、`cover`、`image_count` |
| `/api/scores/types` | GET | **公开** · 返回已使用的乐谱类型（去重） |
| `/api/scores` | POST | 新建乐谱：`{title, composer, type, note, images:[url...]}` |
| `/api/scores/:id` | PATCH | 修改 title / composer / type / note |
| `/api/scores/:id` | DELETE | 删除乐谱 + 图片行 + 本地文件 |
| `/api/scores/:id/images` | POST | 追加图片：multipart `files`（可多张）或 JSON `{urls:[...]}` |
| `/api/scores/images/:imageId` | DELETE | 删除单张图片（含本地文件） |

## 部署

前端静态，改完即生效；后端改动需重新编译并重启：

```bash
cd /home/pi/projects/Frontend/victor-tool-collection/tools/backend
GO111MODULE=on GOPATH=/home/pi/go GOMODCACHE=/home/pi/go/pkg/mod go build -o backend .
sudo systemctl restart vtc-backend
```

nginx（已在 `deploy/nginx/port-8001.conf`）：

```nginx
location /score-library {
    alias /home/pi/projects/Frontend/victor-tool-collection/tools/score-library/dist;
    try_files $uri $uri/ /score-library/index.html;
}
location /scores/files/ {
    alias /home/pi/projects/Frontend/victor-tool-collection/data/scores/;
    expires 7d;
}
```

上传走 `/api/` 代理，该 location 已设 `client_max_body_size 50m`。

## 维护提示

- `data/scores/` 需存在，`pi` 可写、`nginx` 可读（755/644）
- 备份 = 打包 `data/scores/` + 导出 `scores` / `score_images` 两张表
- 外链乐谱（source=remote）不占本地磁盘，但依赖源站可用
