# 移除 RFI sample 外掛（跟 install.ps1 配對）：
#   .\uninstall.ps1 -Version 2024
# 會刪掉 Addins\{版號}\ 下的 RfiPlugin.addin + RfiPlugin.dll，重開 Revit 即消失。
# 注意：Revit 要先關掉，不然 DLL 被咬住刪不掉。
param([string]$Version = "2024")
$ErrorActionPreference = "Stop"
$addinDir = "$env:APPDATA\Autodesk\Revit\Addins\$Version"
foreach ($f in @("RfiPlugin.addin", "RfiPlugin.dll")) {
  $p = Join-Path $addinDir $f
  if (Test-Path $p) { Remove-Item $p -Force; Write-Host "已刪 $p" }
  else { Write-Host "沒有 $p（本來就沒裝）" }
}
Write-Host "完成，重開 Revit $Version 確認分頁消失" -ForegroundColor Green
