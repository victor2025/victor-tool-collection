# 乐谱夹 (Score Library)

个人乐谱收藏工具：一份乐谱对应多张图片，支持本地上传或公网图片链接导入，按名称 / 作曲家搜索，点开后连页滚动翻看，可缩放、可自动滚动。

## 功能

- **多图乐谱**：一份乐谱 = N 张图片（数据库 `scores` + `score_images` 两张表，一对多）
- **两种导入方式**
  - 上传图片：可多选 / 拖拽，整批入库
  - 图片链接：每行一个 URL；`http://` 自动镜像到本地，`https://` 保留外链
- **乐谱类型**：吉他谱 / 钢琴谱 / 简谱 / 五线谱 / 其他，可按类型筛选（类型存在数据库 `scores.type`）
- **搜索**：名称 / 作曲家 / 备注 模糊搜索，输入即搜（220ms 防抖）
- **连页预览**：所有页竖向拼接在同一滚动容器，滚动连续不中断；缩略图条 / `‹ ›` / 键盘 ←→ 跳转到指定页
- **缩放**：50%–300%，按钮 ±25% 或 Ctrl+滚轮（±10%），「适应宽度」一键复位
- **自动滚动**：右侧控制条开关，5–200 px/s 可调（步长 5），支持暂停/继续、空格键切换；手动拖动时临时暂停，停手 1.2s 自动从当前位置续滚；滚到末尾自动停止
- **鉴权**：**查看公开**（列表 / 搜索 / 类型 / 图片无需登录）；写操作（新增/修改/删除/补图/删图）**有登录态（`vtc_session`）直接放行，无登录态则每次校验管理密码**（请求头 `X-VTC-Password`）

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

## 接口（读取公开，写入需登录态或密码）

| 端点 | 方法 | 说明 |
|------|------|------|
| `/api/scores?q=&type=` | GET | **公开** · 列表 / 搜索，返回每份乐谱含 `images[]`、`cover`、`image_count` |
| `/api/scores/types` | GET | **公开** · 返回已使用的乐谱类型（去重） |
| `/api/scores` | POST | 新建乐谱：`{title, composer, type, note, images:[url...]}`，需登录态或 `X-VTC-Password` |
| `/api/scores/:id` | PATCH | 修改 title / composer / type / note，需登录态或 `X-VTC-Password` |
| `/api/scores/:id` | DELETE | 删除乐谱 + 图片行 + 本地文件，需登录态或 `X-VTC-Password` |
| `/api/scores/:id/images` | POST | 追加图片：multipart `files`（可多张）或 JSON `{urls:[...]}`，需登录态或 `X-VTC-Password` |
| `/api/scores/images/:imageId` | DELETE | 删除单张图片（含本地文件），需登录态或 `X-VTC-Password` |

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
