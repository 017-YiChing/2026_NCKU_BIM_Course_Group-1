# RFI 議題追蹤圖台 sample（10/16 課上範例）

對應教材：[`SPEC.md`](../SPEC.md)。第一版最小可用（已驗證：後端 `go test` 過、前端 `vite build` 過）：

- `backend/` — Go + SQLite（`modernc.org/sqlite`，pure Go 免裝 CGO），`go run .` 即跑 `:8080`
- `frontend/` — React 19 + Vite，列表＋詳情＋手填新增（第一版無 IFC viewer，決議見下）
- `revit-plugin/` — C# 外掛，「建立議題」：讀選取 ElementId → 匯出視圖截圖 → POST 到後端

課堂決議（已定案）：前端第一版**先只做列表+詳情**，IFC viewer 為延伸；
外掛用**傳統 `.addin` 部署**（跟 10/02 MCP add-in 同模式，需重開 Revit）。

## 跑起來（三步）

```powershell
# 1. 後端（:8080）
cd sample/backend
go run .

# 2. 前端（:5173，proxy 轉到 :8080）
cd ../frontend
npm install
npm run dev
# 瀏覽器開 http://localhost:5173

# 3. 外掛（教室電腦有 Revit 2024 才能 build；備課機沒裝 Revit 屬正常）
cd ../revit-plugin
dotnet build -c Release
.\install.ps1
# 重開 Revit 2024 → "RFI Sample" 分頁 → 選一個元件 → 按「建立議題」
```

## API（SPEC Ch5 最小集，後端都已實作＋測試）

| 方法 | 路徑 | 說明 |
|------|------|------|
| POST | `/api/issues` | 新增（multipart：欄位＋screenshot 可選） |
| GET | `/api/issues?status=open` | 列表（status 篩選） |
| GET | `/api/issues/:id` | 詳情 |
| PATCH | `/api/issues/:id` | 改狀態／指派（JSON） |
| POST | `/api/models` | 上傳 IFC（存檔＋解析身分＋回 element_count） |
| GET | `/api/models/:id/file` | 下載 IFC（給 viewer 用，viewer 為延伸） |
| GET | `/api/models/:id/elements` | 元件身分列表 |

驗證：`cd backend && go test ./...`（含議題 CRUD＋IFC 身分解析測試）。

## 給各組的提示（SPEC Ch11）

各組**不沿用這個框架改**，做自己的 AI 應用。可參考的只有架構思路：

- Ch2 分層：擷取（外掛）／存取（後端）／呈現（前端）拆開
- Ch3 綁定設計：議題一律綁模型物件（ElementId 或 GlobalId），圖台一點即定位

各組起手骨架（超薄版 starter）另開一輪再做。
