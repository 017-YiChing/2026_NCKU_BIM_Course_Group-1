# 部署到 Revit 2024（傳統 .addin 模式，跟 10/02 MCP add-in 同流程）：
#   1. 先 build：dotnet build -c Release
#   2. 跑本腳本：.\install.ps1
#   3. 重開 Revit 2024 → 出現 "RFI Sample" 分頁 → 按「建立議題」
$ErrorActionPreference = "Stop"
$addinDir = "$env:APPDATA\Autodesk\Revit\Addins\2024"
New-Item -ItemType Directory -Force $addinDir | Out-Null
$dll = "bin\Release\net48\RfiPlugin.dll"
if (-not (Test-Path $dll)) { throw "找不到 $dll，先跑 dotnet build -c Release" }
Copy-Item RfiPlugin.addin (Join-Path $addinDir "RfiPlugin.addin") -Force
Copy-Item $dll (Join-Path $addinDir "RfiPlugin.dll") -Force
Write-Host "已部署到 $addinDir，重開 Revit 2024 即可使用" -ForegroundColor Green
