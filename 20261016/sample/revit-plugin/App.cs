using Autodesk.Revit.UI;
using System;
using System.Reflection;

namespace RfiPlugin
{
    /// <summary>
    /// 外掛進入點：Revit 啟動時建一個 Ribbon 面板，放一顆「建立議題」按鈕。
    /// 傳統 .addin 部署：RfiPlugin.addin + RfiPlugin.dll 一起丟到
    /// %APPDATA%\Autodesk\Revit\Addins\2024\，重開 Revit 即生效（跟 10/02 MCP add-in 同模式）。
    /// </summary>
    public class App : IExternalApplication
    {
        public Result OnStartup(UIControlledApplication app)
        {
            const string tabName = "RFI Sample";
            try { app.CreateRibbonTab(tabName); } catch { /* 已存在就跳過 */ }

            var panel = app.CreateRibbonPanel(tabName, "議題追蹤");
            var asm = Assembly.GetExecutingAssembly().Location;
            var btnData = new PushButtonData(
                "CreateIssue", "建立議題",
                asm, typeof(CreateIssueCommand).FullName);
            btnData.ToolTip = "把目前選取的元件 + 使用中視圖截圖，送到 RFI 圖台後端";
            panel.AddItem(btnData);
            var openData = new PushButtonData(
                "OpenIssueView", "開議題視圖",
                asm, typeof(OpenIssueViewCommand).FullName);
            openData.ToolTip = "從圖台選一筆議題，切到它當初的視圖";
            panel.AddItem(openData);
            return Result.Succeeded;
        }

        public Result OnShutdown(UIControlledApplication app) => Result.Succeeded;
    }
}
