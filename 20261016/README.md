# 2026-10-16 Revit 外掛開發（以 RFI 議題追蹤圖台為例）

- 主題：Revit 外掛開發（以 RFI 議題追蹤圖台為例）
- 教材：[SPEC.md](SPEC.md)（RFI 圖台框架規格）
- 骨架：[sample/](sample/)（最小可用，課上直接拿去改；`.\start.ps1` 一鍵啟動）
- 課堂決議：前端第一版先只做列表+詳情（無 IFC viewer）；外掛用傳統 `.addin` 部署

## 基本開發原則（課上只認這幾條）

1. **議題一律綁模型物件**：Revit 送 `ElementId`，IFC 送 `GlobalId`；圖台一點即定位，不接受「只用文字描述位置」的議題。
2. **分層不越界**：外掛只管擷取（選取＋截圖＋送出）、後端只管存取（存議題/存檔/解析）、前端只管呈現。加功能先問屬於哪一層。
3. **最小可用先行**：先讓鏈路跑通（送得出、看得到），再加欄位、美化、viewer。沒跑通不加東西。
4. **可追溯**：模型大改前先另存副本；程式改動走 commit；議題狀態只走 open → in_progress → closed。
5. **先驗證再相信**：agent 寫的任何東西，回 Revit 或瀏覽器親眼確認；截圖就是證據。

## 必備知識（0 經驗版，每項夠用就好）

- **HTTP**：GET 查、POST 送、PATCH 改；狀態碼 2xx 才算成功。看 `backend/handlers_issues.go` 對 `frontend/src/api.js`。
- **C# 最小集**：class、方法、try/catch；WinForms 就是拉三個輸入框。看 `revit-plugin/IssueForm.cs`。
- **Revit API 最小集**：`Selection.GetElementIds()` 讀選取、`ImageExportOptions` 匯出截圖、`PushButtonData` 掛按鈕。三個都在 `revit-plugin/` 裡，用到才學。
- **Go 最小集**：handler 就是「收請求→碰資料庫→回 JSON」。照抄 `handlers_issues.go` 改。
- **React 最小集**：`useState` 存資料、`fetch` 打 API、列表 onClick 切詳情。看 `frontend/src/App.jsx`。
- **SQLite**：單一檔案資料庫，`issues`/`models`/`elements` 三張表（見 SPEC Ch3）；壞了刪掉重跑自動重建。

## Add-in Manager（可以用，但課上不用）

- 它是什麼：Revit 外掛熱載入工具，build 出 DLL 直接載入，**免寫 .addin、免重開 Revit**，調試迴圈最快。
- 為什麼課上不用：10/02 學生已學會傳統 `.addin` 部署（`install.ps1` 一鍵裝），全班同一套流程最省事；多引一個工具就多一處卡關。
- 自學指路：裝好 Add-in Manager 後，用它載入 `revit-plugin/bin/Release.R24/net48/RfiPlugin.dll` 即可，程式碼不用改。

## 課程流程

課上 demo 骨架全鏈路 → 各組發散改成自己的題目 → 收斂交件（見 `sample/README.md`「交件」）。
