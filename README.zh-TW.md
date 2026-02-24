# AgentSkills Registry

[English](./README.md) | **中文**

AgentSkills 是一個 AI Agent Skill 的集中式 Registry 平台，類似 npm 或 Docker Hub。開發者可透過 CLI 工具上傳（push）與下載（pull）標準化的 Skill Bundle，平台負責版本控制、Metadata 解析與檔案儲存。

## 架構

- **後端 API**：FastAPI (Python 3.12+)，位於 `api/`
- **CLI 工具**：Go + Cobra，位於 `cli/`
- **資料庫**：PostgreSQL 16（Docker）
- **物件儲存**：MinIO S3 相容（Docker）
- **規格文件**：完整設計請參閱 [`reference/SDD.md`](./reference/SDD.md)

## 快速開始

### 啟動基礎設施

```bash
docker compose up -d
```

這會啟動 PostgreSQL（port 5432）和 MinIO（API: 9000, Console: 9001）。

### 啟動 API（開發模式）

```bash
cd api
pip install -r requirements.txt
uvicorn app.main:app --reload --port 8000
```

### 建置 CLI

```bash
cd cli
go build -o bin/agentskills .
```

### 驗證

```bash
# 健康檢查
curl http://localhost:8000/v1/health

# 使用 CLI
./cli/bin/agentskills --help
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
├── api/                  # FastAPI 後端
│   ├── app/
│   │   ├── routes/       # API 端點
│   │   └── services/     # 業務邏輯
│   ├── tests/
│   ├── Dockerfile
│   └── requirements.txt
├── cli/                  # Go CLI 工具
│   ├── cmd/              # Cobra 指令
│   ├── internal/         # 內部套件
│   ├── Dockerfile
│   ├── Makefile
│   └── go.mod
├── reference/            # 設計文件
│   └── SDD.md
├── docker-compose.yml
└── init.sql
```

## 授權

請參閱 [LICENSE](./LICENSE)。
