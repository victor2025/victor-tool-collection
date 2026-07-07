# AGENTS.md — Victor Tool Collection 项目规范

> 此文件面向 AI Agent 和工作协作者，说明项目维护规则。

## 📌 核心规则

### 🚫 禁止直接提交和推送

**所有代码修改必须先展示给老板确认，严禁直接 `git commit` 和 `git push`。**

只有在老板明确说出「可以提交」或「提交吧」之后，才允许执行 commit 和 push 操作。

> 违反此规则 = 信用破产，切记。

### 每次修改项目 → 必须更新 README.md

凡是对本项目做出以下任何变更后，**必须同步更新项目根目录的 `README.md`**：

- 新增、删除或重命名工具模块
- 更换技术栈或构建方式
- 变更项目结构或目录约定
- 修改部署方式或 nginx 配置
- 新增外部依赖或运行时要求

### 每个模块 → 必须有自己的 README.md

`tools/` 下的每一个工具子目录，都必须包含一个 `README.md`，至少说明：

- 该工具的用途和功能
- 技术栈（框架、语言、库）
- 构建与部署方式
- 目录结构说明

修改某个模块后，必须同步更新其模块 `README.md`。

### 每次 Git 提交前 → 检查 README 一致性

在每次执行 `git commit` 之前，**必须先确认以下文件已同步更新**：

- 本次变更涉及了哪些模块？对应的 `tools/<模块>/README.md` 是否已更新？
- 项目结构或工具有变化？根目录 `README.md` 是否已更新？
- 部署方式或 nginx 配置有变化？`deploy/README.md` 是否已更新？

> 📌 提交信息中若涉及某模块的变更，该模块的 README.md 必须同步反映最新状态。

### 每个文本框 → 右上角必须有「清空」「复制」按钮

参考 `tools/base64/` 的 UI 模式，每个文本框（textarea）和输出框的**右上角**必须放置「清空」和「复制」两个按钮：

- **清空按钮**（`.act.clear`）：清空当前文本框的内容
- **复制按钮**（`.act`）：将当前文本框的内容复制到剪贴板
- 采用 `.editor-label`（或 `.pane-header`）左右分栏布局：左侧放名称/计数，右侧放操作按钮

```html
<div class="pane-header">
  <span class="left">输入 <span class="badge">...</span></span>
  <span class="right">
    <span class="act clear" onclick="...">清空</span>
    <span class="act" onclick="...">复制</span>
  </span>
</div>
```

### 导航页 Tab 模式行为

导航页 (`nav/index.html`) 使用多 iframe 显隐切换方案：

- **每个工具 Tab 创建独立的 iframe**，只加载一次
- **切 Tab 只切换 `display: none/block`**，不销毁/重建 DOM，不重新加载页面
- **Tab 模式下不展示返回按钮**——顶部 Tab 栏始终有 🏠 首页 Tab 可点击回到主页
- 关闭 Tab（✕）时同步移除对应的 iframe 容器，释放内存

### 工具页顶部边距规范

所有工具的标题（h1）距离页面顶部的高度必须统一。

- **当前值：** `30px`
- **实现方式：** 在 `<style>` 中定义 CSS 变量 `:root { --title-top: 30px; }`，body 的 padding-top 引用该变量：`padding: var(--title-top) 20px`
- **修改方法：** 只需改 `--title-top` 的值，所有工具的标题顶部距离就会同步更新
- **已覆盖工具：** base64、qrcode、json-formatter、jwt-decoder、webshell
- **新增工具：** 必须遵循此规范，添加 `--title-top` CSS 变量并引用到 body padding-top

### 已有 README.md 的模块

- `tools/webshell/README.md` — 已有较完整文档，维护时需同步更新
- `deploy/README.md` — 部署说明文档，有变更时应同步更新

---

## 📂 项目结构

```
victor-tool-collection/
├── AGENTS.md          ← 本文件（项目规范）
├── README.md          ← 项目总介绍（必须保持最新）
├── nav/               ← 导航页
│   └── index.html
├── tools/             ← 工具模块（每个子目录必须有 README.md）
│   ├── base64/        Base64 编解码
│   ├── json-formatter/ JSON 格式化/转义工具
│   ├── qrcode/        二维码工具
│   ├── score-board/   记分板 (React)
│   ├── timestamp/           # 时间戳转换
│   └── webshell/      Web 终端 (ttyd)
├── deploy/            ← 部署配置
│   ├── README.md
│   └── nginx/
│       └── port-8001.conf
```

---

## ✅ 提交前检查清单

**每次 `git commit` 前**，走一遍这个清单：

### 必做
- [ ] 涉及变更的模块 `README.md` 已同步更新
- [ ] 根目录 `README.md` 已同步更新（如项目结构、工具列表有变化）

### 视情况
- [ ] `nav/index.html` — 导航页是否需添加/修改入口
- [ ] `deploy/README.md` — 部署说明是否需更新
- [ ] `deploy/nginx/port-8001.conf` — nginx 配置是否已添加/修改

> `AGENTS.md` 本身也记录了项目规范，如有更新也记得提交。
