# 專案表單 PR 檢查步驟

適用範圍：`feat/lebr-plugin-dev` 分支發 PR 回組 fork 或主 repo 之前。

## 一、分支與範圍

- [ ] 分支從最新 `main` 建立，且只在 `20261030/第1組/` 內修改。
- [ ] 沒有修改其他組的資料夾，也沒有改名「第四組」。
- [ ] 沒有直接推送到 `main`。
- [ ] 移除測試檔 `20261030/第1組/261025_test.md`（內容為「測試用，勿commit」）。

## 二、禁止提交的內容

- [ ] `git status` 沒有 `.rvt` 檔案。
- [ ] 沒有 `bin/`、`obj/` 建置產物。
- [ ] 沒有 DLL、`LEBR_SharedParams.txt`、測試模型或已編譯外掛。
- [ ] 沒有個人的 Revit 路徑、帳號或暫存檔。

## 三、資料與路徑

- [ ] 外掛程式只使用相對路徑引用 `dist/`、`csharp/`、`samples/`，沒有寫死絕對路徑。
- [ ] `dist/lebr_2025.json` 與 `corrections.csv` 一致。若有修正，已執行 `python -I scripts/build_json.py .` 並提交新的 `dist/lebr_2025.json`。
- [ ] 沒有手動修改 `dist/`、CSV 或 JSON 的內容。
- [ ] `REVIEW_STATUS` 仍為 `draft`，除非已完成人工核對並寫入核對紀錄。

## 四、表單規則（專案設定）

- [ ] 表單分為 A 主結構（標題為「基準案同設計案」）、B 非主結構、C 優惠係數（LL）三部分。
- [ ] A-1（I、SaD、Fu、L、D0、Rs、覆寫 S）為設計案與基準案共用。
- [ ] A-2（CSER、覆寫 Sp、覆寫 F）僅用於設計案。
- [ ] 使用估算時，A 區所有欄位鎖定。
- [ ] 主結構的檢查放在 A 區底部。
- [ ] Rs 可選 0.95 與 1.00，預設為 0.95。
- [ ] CSER 只在構造為 RC 或 SRC 時可勾選；未勾選時為 0.00；勾選後可輸入任意數值。
- [ ] LL 的項目 1、2 只在構造為 RC 時可選；項目 3、4 不限構造。
- [ ] LL 合計超過 0.08 時以 0.08 計算，並顯示提醒。
- [ ] 取消勾選覆寫欄位時，已填的值會被清空。

## 五、運算引擎規則

- [ ] 設計案 Sp、F 的缺少檢查只影響設計案；基準案不因此中止計算。
- [ ] 基準案 W、Sp、F 依樓層與 AFa 查表 4-6，不讀取設計案的 Sp、F。
- [ ] 基準案 EBF = 0（Rn = 1.0）。
- [ ] 基準案 `LCCR = 1.0`，不受 CSER 影響。
- [ ] 估算值只套用於設計案；若已接入估算程式，估算來源有記錄在「提醒與未對應」。

## 六、測試與建置

- [ ] 執行 `dotnet run --project csharp/LebrDatabase.Tests -f net8.0`，結果為「全部通過」。
- [ ] 附錄三案例仍為 CFR 20.51%、1+ 級。
- [ ] 執行 `powershell -ExecutionPolicy Bypass -File revit-addin/install.ps1 -RevitVersion 2027 -Folder LebrCarbon_<n>` 能成功編譯。
- [ ] 若有改動 UI，已用 `dotnet run --project revit-addin/Preview -- . shots --interactive` 預覽畫面。
- [ ] 在 Revit 2027 以小模型測試：建立參數、填寫專案設定、開始計算、匯出 JSON。

## 七、文件

- [ ] 若改動交接格式（數量格式、結果格式）、構件代號或參數名稱，已更新 `交接說明_外掛.md` 或 `README.md`。
- [ ] 若有上述改動，已先在群組說明，並提醒已填代號的模型需要重填。
- [ ] 核對紀錄（若有）已填寫日期、核對人與範圍。

## 八、Commit 與 PR

- [ ] Commit 訊息使用 conventional commits 格式，例如 `feat(lebr): 新增專案設定表單`、`fix(lebr): 修正基準案 Sp 查表`。
- [ ] 文件與 PR 說明使用繁體中文（zh-TW）。
- [ ] PR 標題包含組別與題目，例如「第1組｜LEBR 專案設定表單」。
- [ ] PR 說明列出：改動範圍、測試結果、已知限制。
- [ ] 由業師 review 後以 squash merge 合併。
- [ ] 合併前若 `main` 有新的第1組變更，先同步並解衝突。

## 九、合併後

- [ ] 在主 repo 的 `main` 確認 `20261030/第1組/` 內容與分支一致。
- [ ] 通知組員重新安裝外掛（若程式或資料庫有變更）。
- [ ] 刪除已合併的功能分支（若不再使用）。
