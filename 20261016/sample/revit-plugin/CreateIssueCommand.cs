using Autodesk.Revit.Attributes;
using Autodesk.Revit.DB;
using Autodesk.Revit.UI;
using System;
using System.IO;
using System.Linq;
using System.Net.Http;

namespace RfiPlugin
{
    /// <summary>
    /// 「建立議題」指令（SPEC Ch4）：
    /// 1. 讀目前選取（取第一個；沒選就提示先選）
    /// 2. 匯出使用中視圖截圖（ImageExportOptions）
    /// 3. 小視窗輸入標題＋說明＋回報人（WinForms）
    /// 4. HttpClient POST 到 POST /api/issues（multipart：欄位＋screenshot）
    /// </summary>
    [Transaction(TransactionMode.Manual)]
    public class CreateIssueCommand : IExternalCommand
    {
        // 課堂內網 demo 用；各組做自己題目時改成自己的後端位址。
        private const string BackendBase = "http://localhost:8080";

        public Result Execute(ExternalCommandData data, ref string message, ElementSet elements)
        {
            var uiDoc = data.Application.ActiveUIDocument;
            if (uiDoc == null) { message = "請先開啟一個專案文件"; return Result.Failed; }
            var doc = uiDoc.Document;

            // 1. 選取
            var selIds = uiDoc.Selection.GetElementIds().ToList();
            if (selIds.Count == 0)
            {
                TaskDialog.Show("RFI", "請先在模型裡選一個元件，再按「建立議題」。");
                return Result.Cancelled;
            }
            string elementId = selIds[0].Value.ToString();

            // 2. 截圖：匯出使用中視圖
            string shotPath = Path.Combine(Path.GetTempPath(), $"rfi_{DateTime.Now:yyyyMMddHHmmss}.png");
            try
            {
                var opt = new ImageExportOptions
                {
                    ExportRange = ImageExportRange.CurrentView,
                    FilePath = shotPath,
                    ImageResolution = ImageResolution.DPI_150,
                    HLRandWFViewsFileType = ImageFileType.PNG,
                    ShadowViewsFileType = ImageFileType.PNG,
                };
                doc.ExportImage(opt);
            }
            catch (Exception ex)
            {
                TaskDialog.Show("RFI", "截圖匯出失敗：" + ex.Message);
                return Result.Failed;
            }

            // 3. 輸入標題＋說明
            using (var form = new IssueForm())
            {
                if (form.ShowDialog() != System.Windows.Forms.DialogResult.OK)
                    return Result.Cancelled;

                // 4. POST 到後端
                try
                {
                    using (var http = new HttpClient())
                    using (var content = new MultipartFormDataContent())
                    {
                        content.Add(new StringContent(form.IssueTitle), "title");
                        content.Add(new StringContent(form.IssueDescription ?? ""), "description");
                        content.Add(new StringContent(form.Reporter ?? ""), "reporter");
                        content.Add(new StringContent("revit"), "source");
                        content.Add(new StringContent(elementId), "revit_element_id");
                        if (File.Exists(shotPath))
                        {
                            var bytes = File.ReadAllBytes(shotPath);
                            var img = new ByteArrayContent(bytes);
                            img.Headers.ContentType = new System.Net.Http.Headers.MediaTypeHeaderValue("image/png");
                            content.Add(img, "screenshot", Path.GetFileName(shotPath));
                        }
                        var resp = http.PostAsync(BackendBase + "/api/issues", content).Result;
                        string body = resp.Content.ReadAsStringAsync().Result;
                        if (!resp.IsSuccessStatusCode)
                        {
                            TaskDialog.Show("RFI", $"後端回錯誤 {(int)resp.StatusCode}：{body}");
                            return Result.Failed;
                        }
                        TaskDialog.Show("RFI", "議題已送出！回到圖台網頁查看。");
                    }
                }
                catch (Exception ex)
                {
                    TaskDialog.Show("RFI", "送出失敗（後端有跑嗎？go run ./backend）：" + ex.Message);
                    return Result.Failed;
                }
            }
            return Result.Succeeded;
        }
    }
}
