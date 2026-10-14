# RFI 圖台一鍵啟動（學生課上用，越簡單越好）
#
# 用法：在 sample 資料夾按右鍵 →「使用 PowerShell 執行」，或：
#   .\start.ps1
#
# 會開兩個視窗：一個跑後端（:8080，loopback 不跳防火牆），一個跑前端（:5173）。
# 跑起來後瀏覽器開 http://localhost:5173；關掉兩個視窗即停止。
$ErrorActionPreference = "Stop"
$root = $PSScriptRoot

# 基本檢查：缺什麼先講，不直接炸掉
foreach ($cmd in @("go", "node", "npm")) {
  if (-not (Get-Command $cmd -ErrorAction SilentlyContinue)) {
    Write-Host "缺 $cmd，先照 sample/README.md「課前裝好」裝完再跑" -ForegroundColor Red
    exit 1
  }
}

Write-Host "[1/2] 啟動後端（Go + SQLite，http://127.0.0.1:8080）..." -ForegroundColor Green
Start-Process powershell -ArgumentList "-NoExit", "-Command", "cd '$root\backend'; go run ."

Write-Host "[2/2] 啟動前端（React，http://localhost:5173）..." -ForegroundColor Green
Start-Process powershell -ArgumentList "-NoExit", "-Command", "cd '$root\frontend'; if (-not (Test-Path node_modules)) { Write-Host '第一次跑，先裝依賴（約1分鐘）...'; npm install }; npm run dev"

Write-Host ""
Write-Host "兩個視窗都出現 ready / listening 後，瀏覽器開 http://localhost:5173" -ForegroundColor Cyan
Write-Host "外掛另裝：cd revit-plugin；dotnet build -c Release.R24；.\install.ps1 -Version 2024（重開 Revit）" -ForegroundColor Cyan
