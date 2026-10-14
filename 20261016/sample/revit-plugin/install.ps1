# 部署（傳統 .addin 模式，跟 10/02 MCP add-in 同流程）：
#   學生（Revit 2024）：dotnet build -c Release.R24; .\install.ps1 -Version 2024
#   備課機（Revit 2026）：dotnet build -c Release.R26; .\install.ps1 -Version 2026
#   跑完重開對應版 Revit → "RFI Sample" 分頁 → 選一個元件 → 按「建立議題」
param([string]$Version = "2024")
$ErrorActionPreference = "Stop"
$addinDir = "$env:APPDATA\Autodesk\Revit\Addins\$Version"
New-Item -ItemType Directory -Force $addinDir | Out-Null
$dll = "bin\Release.R$Version\net48\RfiPlugin.dll"
if (-not (Test-Path $dll)) { throw "找不到 $dll，先跑 dotnet build -c Release.R$Version" }
Copy-Item RfiPlugin.addin (Join-Path $addinDir "RfiPlugin.addin") -Force
Copy-Item $dll (Join-Path $addinDir "RfiPlugin.dll") -Force
Write-Host "已部署到 $addinDir，重開 Revit $Version 即可使用" -ForegroundColor Green
