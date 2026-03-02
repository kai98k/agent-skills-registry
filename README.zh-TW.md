# AgentSkills Registry

[English](./README.md) | **中文**

AgentSkills 是一個 AI Agent Skill 的集中式 Registry 平台，類似 npm 或 Docker Hub。開發者可透過 CLI 工具上傳（push）與下載（pull）標準化的 **Skill Bundle**，平台負責版本控制、Metadata 解析與檔案儲存。

## 為什麼需要 AgentSkills？

AI Agent（例如 Claude Code、AutoGPT、LangChain Agent）越來越強大，但目前缺乏一套標準化的方式來**分享與重用能力**。AgentSkills 透過以下方式解決這個問題：

- **標準化 Skill 格式** — 每個 Skill 是一個目錄，包含 `SKILL.md`（YAML frontmatter + Markdown 指令）、可選的 `scripts/`（Agent 可呼叫的腳本）、`references/`（RAG / few-shot 參考文件）和 `assets/`（靜態模板與資源）。
- **一鍵分享** — `agentskills push` 打包並上傳 Skill；`agentskills pull` 下載並解壓。就像 `npm publish` / `npm install` 一樣簡單。
- **Agent 框架無關** — Skill Bundle 是純檔案格式，任何 Agent 框架都可以透過讀取 `SKILL.md` 獲取指令，並載入附帶的腳本與參考文件。
- **版本控制** — 每個 Skill 使用嚴格的 semver 版本號，團隊可以鎖定特定版本，按自己的節奏升級。
- **可搜尋性** — `agentskills search` 讓開發者透過關鍵字或標籤（如 `code-review`、`data-analysis`、`devops`）找到社群貢獻的 Skill。

### Skill Bundle 結構

```
my-skill/
├── SKILL.md         # 必填：YAML frontmatter（name, version, description, author, tags）+ Markdown 指令
├── scripts/         # 選填：Agent 可執行的腳本
├── references/      # 選填：RAG / few-shot 參考文件
└── assets/          # 選填：靜態模板與資源
```

## 運作方式

```
AI Agent（Claude Code、AutoGPT 等）
        │
        │  在 shell 中執行 CLI 指令
        ▼
   agentskills CLI ──── HTTP REST API ────► agentskills server
   (Go binary)                              (Go binary，使用 -tags server 編譯)
```

CLI 是 Agent 與 Registry Server 之間的唯一介面。Agent 只需在 shell 中執行指令即可操作平台，不需要了解底層 HTTP API：

- `agentskills search code-review` → 在 server 上搜尋符合的 Skill
- `agentskills pull code-review` → 下載並解壓 Skill Bundle 到本地
- `agentskills push ./my-skill` → 打包並上傳 Skill 到 server

## 架構

- **後端 API + CLI 工具**：Go + Cobra，位於 `cli/`（server 使用 `-tags server` 編譯）
- **儲存**：檔案式儲存（bundle 以 `.tar.gz` 保存，metadata 為 JSON）
- **規格文件**：完整設計請參閱 [`reference/SDD.md`](./reference/SDD.md)

## 快速開始

### 啟動 Server（Docker）

```bash
docker build -t agentskills-server .
docker run -p 8000:8000 -v agentskills-data:/data agentskills-server
```

### 建置 CLI

```bash
cd cli
go build -o bin/agentskills .
```

### 驗證

```bash
# 使用 CLI
./cli/bin/agentskills --help

# 搜尋 Skills
./cli/bin/agentskills search test
```

## CLI 指令

| 指令 | 說明 |
|------|------|
| `agentskills init <name>` | 建立 Skill 骨架目錄 |
| `agentskills login` | 儲存 API Token 至本地設定 |
| `agentskills push <path>` | 打包並上傳 Skill Bundle |
| `agentskills pull <name>[@version]` | 下載並解壓 Skill Bundle |
| `agentskills search <keyword>` | 搜尋平台上的 Skills |

## 專案結構

```
.
├── cli/                  # Go CLI + Server
│   ├── cmd/              # Cobra 指令（含 serve）
│   ├── server/           # HTTP handler 與檔案儲存
│   ├── internal/         # 內部套件
│   ├── Dockerfile
│   ├── Makefile
│   └── go.mod
├── reference/            # 設計文件
│   └── SDD.md
├── Dockerfile            # Server Docker image
└── docker-compose.yml
```

## FAQ

**Q: 為什麼會有這個專案？**

AgentSkills Registry 讓 AI Agent 有一個標準化的方式分享與重用能力——就像 npm 對 JavaScript 開發者的意義。不用每個 Agent 都從零開始，Skill 發佈一次即可供所有人使用。

**Q: Agent 可以自主建立和發佈 Skill 嗎？**

可以。從建立 `SKILL.md` 到 push 的整個流程，都可以完全由 AI Agent 驅動。人類只需表達意圖，Agent 負責執行。這個 repo 本身也大量由 Agent 開發完成。

**Q: 可以自建私有的 Registry 嗎？**

可以。Server 是一個獨立的 Go binary，用 Docker 一行指令即可部署。CLI 透過 `--api-url` 指向你的私有 server，就像自建 GitLab 一樣，適合企業內部或離線環境使用。

```bash
# 自建範例
docker run -p 8000:8000 -v my-data:/data agentskills-server
agentskills search test --api-url http://my-server:8000
```

**Q: 誰負責發佈的 Skill 品質？**

Skill 作者（人類或 Agent）負責。使用者應在 production 環境使用前審查 `SKILL.md` 和相關腳本。建議鎖定特定版本並審計 Skill 內容，就像對待任何其他依賴一樣。

**Q: Agent 驅動開發的最佳實踐是什麼？**

讓 AI Agent（Claude Code、Cursor 等）可以存取 `agentskills` CLI。人類提供方向，Agent 負責搜尋現有 Skill、pull 作為參考、建立新 Skill、push 到 Registry——全部透過 shell 指令完成。

```
                  意圖 / 方向
  人類擁有者 ─────────────────────────► AI Agent
                                          │
                                          │ agentskills CLI
                                          ▼
                                    AgentSkills Registry
```

## 授權

請參閱 [LICENSE](./LICENSE)。
