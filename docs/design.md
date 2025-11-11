# CodeCopyBook – System Design

## 1. Problem Overview
Create a “code copybook” web app where users trace source files character‑by‑character. Users upload individual files or folders, select one file, and practice by typing. Typed characters gradually darken the corresponding tokens over a faint reference background, turning copying into an interactive exercise.

Core needs:
- Upload/manage source assets (files or zipped folders).
- Parse and render code with light theme preview.
- Typing workspace that validates keystrokes, shows progress, and darkens correctly typed words/characters.
- Persist practice state per session (to resume mid‑file).
- Deployable in both “pure software” (local binaries + npm) and containerized forms.

## 2. High-Level Architecture
```
┌──────────┐       HTTPS/JSON        ┌──────────────────┐
│ Frontend │  <--------------------> │ Go Backend API   │
│ (React   │                         │ + Storage        │
│  + Vite) │                         │ (local FS)       │
└────┬─────┘                         └──────┬───────────┘
     │                                         │
     │  SPA assets (S3, CDN, or same host)      │
     │                                          ▼
     └───────────── Deployment Layer ───────────┘
```

- **Frontend**: React + TypeScript + Vite；通过双层 `<pre>` + CSS 控制浅色字帖与加深效果，负责上传、文件树和键入互动。
- **Backend**: Go (`net/http`)。负责上传（单文件 / ZIP）、文件树、源码读取、练习会话同步。
- **Storage**: JSON 元数据 + 本地磁盘（`data/uploads/<assetId>`）保存实际源码。
- **Deployment**: 
  - Pure software mode: run `go run` backend + `npm run dev/build` frontend, configure via `.env`.
  - Container mode: multi-stage Docker builds for frontend and backend; docker-compose for local stack.

## 3. Functional Components
### 3.1 Asset Management
- Upload endpoint accepts single source files or zipped folders.
- Backend detects MIME, validates file type/size, extracts zipped folder preserving structure.
- Metadata stored with `asset_id`, `name`, `path`, `language`, `hash`.
- Provide listing API returning tree per asset to select file for practice.

### 3.2 Practice Session Flow
1. User picks asset + file path.
2. Frontend fetches file content, tokenizes, renders faint baseline using syntax highlight.
3. Typing overlay tracks cursor index.
4. On correct keystroke: overlay letter darkens; update progress meter. On mistake: show subtle highlight and optionally block progress until corrected.
5. Periodically POST progress to backend (session id). Allows resume: GET returns last cursor index and typed history snapshot.

### 3.3 Rendering Model
- Use `Monaco` or `PrismJS`. Implementation plan: Prism for lightweight highlight; computed spans with color classes for faint/dark states.
- Maintain per-character state: `pending`, `active`, `completed`.
- Provide adjustable opacity slider and font size to mimic copybook.
- Overlay层包含一个闪烁的光标（blinking caret），实时标记当前应输入的字符；当用户抵达文件尾部时，光标自动停留在末尾，提示练习已完成。

### 3.4 User Experience
- Landing page with asset list and upload widget.
- Practice page layout:
  - Left sidebar: file tree, session stats.
  - Main pane: code canvas with layered text (background reference + typed overlay).
  - Bottom: typing controls, accuracy meter, per-token progress。
- 练习工具栏提供“跳过当前行”按钮，可在遇到长注释或暂不想练的段落时直接将光标推进到下一行首；行为同样会同步到会话进度。

## 4. API Design (REST JSON)
All endpoints prefixed with `/api`.

| Method | Path | Description |
|--------|------|-------------|
| `POST` | `/assets/upload` | Upload single file (`file`) or zip folder; returns asset metadata. |
| `GET` | `/assets` | List uploaded assets with aggregated metadata. |
| `GET` | `/assets/{id}/tree` | Returns file/folder tree for asset (recursive). |
| `GET` | `/assets/{id}/file?path=...` | Returns file content + detected language. |
| `POST` | `/sessions` | Create practice session `{asset_id, path}`; returns session id. |
| `GET` | `/sessions/{id}` | Fetch session state (cursor, accuracy, timestamps). |
| `PATCH` | `/sessions/{id}` | Update progress `{cursor, errors, duration}`. |
| `DELETE` | `/assets/{id}` | Remove asset + sessions. |

Authentication is out of scope for first iteration; rely on local usage.

## 5. Data Model（当前实现）
- 资产信息（`assets.json`）：`id`、`name`、`rootPath`、`sizeBytes`、`fileCount`、`createdAt/UpdatedAt`、`sourceName`。
- 会话信息（`sessions.json`）：`id`、`assetId`、`relPath`、`cursor`、`errors`、`durationSeconds`、`createdAt/UpdatedAt`。

JSON 元数据配合磁盘目录 `data/uploads/<assetId>` 保存提取后的源码树。初版避免外部依赖（SQLite 等），后续可平滑迁移到数据库。

## 6. Technology Choices
- Frontend: React 18 + TypeScript + Vite，使用原生 hooks 管理状态，CSS 实现图层式字帖效果。
- Backend: Go 1.21+，原生 `net/http` + 自定义路由，JSON 文件持久化 + 本地文件系统存储上传内容。
- 构建：npm（前端）+ Go toolchain（后端）。
- 测试策略：后端以 `go test` 验证存储/处理逻辑，前端采用 Vitest/VitePreview（后续迭代）。

## 7. Deployment Strategy
- 核心环境变量：`ADDR`、`DATA_DIR`、`FRONTEND_DIR`。
- 纯软件：`npm run build` 产出 `frontend/dist`，`go build` 生成二进制并通过 `FRONTEND_DIR` 同站托管静态资源。
- 容器：多阶段 `Dockerfile` 同时构建前后端，`docker compose` 暴露 8080 并挂载数据卷。
- `scripts/deploy.sh` 提供 `local` / `docker` 两种模式，完成上述步骤。

## 8. Iterative Implementation Plan
1. Scaffold repo：Vite + Go module。
2. 实作资产/会话 API 与 JSON 存储。
3. 构建前端 UI/UX（上传、文件树、临摹画布、键入逻辑）。
4. 增强会话同步、提示、错误处理。
5. 添加部署脚本、Dockerfile、Compose 与文档。

## 9. Risks & Mitigations
- **Large uploads**: enforce size limits (config) and stream to disk.
- **Syntax highlight accuracy**: fallback to plain text if language detection fails.
- **State sync**: use optimistic updates and periodic server sync with debounced PATCH.
- **Folder upload UX**: require zipped folder to keep backend simple.
