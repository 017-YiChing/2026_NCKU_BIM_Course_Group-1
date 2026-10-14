using Autodesk.Revit.Attributes;
using Autodesk.Revit.DB;
using Autodesk.Revit.UI;
using System;
using System.Collections.Generic;
using System.Linq;
using System.Net.Http;
using System.Text.RegularExpressions;
using System.Windows.Forms;

namespace RfiPlugin
{
    /// <summary>
    /// 「開議題視圖」：從後端抓議題列表，選一筆就在 Revit 切到它的視圖。
    /// 找視圖先用 revit_view_id（UniqueId，最穩），沒有才用 view_name 比對。
    /// </summary>
    [Transaction(TransactionMode.Manual)]
    public class OpenIssueViewCommand : IExternalCommand
    {
        private const string BackendBase = "http://localhost:8080";

        private class Row
        {
            public long Id; public string Title; public string ViewName; public string ViewUid;
            public override string ToString() => $"#{Id} {Title}（{ViewName}）";
        }

        public Result Execute(ExternalCommandData data, ref string message, ElementSet elements)
        {
            var uiDoc = data.Application.ActiveUIDocument;
            if (uiDoc == null) { message = "請先開啟一個專案文件"; return Result.Failed; }
            var doc = uiDoc.Document;

            List<Row> rows;
            try
            {
                using (var http = new HttpClient())
                {
                    string json = http.GetStringAsync(BackendBase + "/api/issues").Result;
                    rows = ParseRows(json);
                }
            }
            catch (Exception ex)
            {
                TaskDialog.Show("RFI", "抓不到議題（後端有跑嗎？go run ./backend）：" + ex.Message);
                return Result.Failed;
            }
            rows = rows.Where(r => !string.IsNullOrWhiteSpace(r.ViewName)).ToList();
            if (rows.Count == 0)
            {
                TaskDialog.Show("RFI", "圖台上還沒有綁視圖的議題（舊議題重建才會有視圖）。");
                return Result.Cancelled;
            }

            using (var form = new System.Windows.Forms.Form())
            {
                form.Text = "開議題視圖";
                form.Size = new System.Drawing.Size(440, 360);
                form.StartPosition = System.Windows.Forms.FormStartPosition.CenterParent;
                var list = new System.Windows.Forms.ListBox { Dock = System.Windows.Forms.DockStyle.Fill };
                list.Items.AddRange(rows.Cast<object>().ToArray());
                var ok = new System.Windows.Forms.Button { Text = "開啟視圖", DialogResult = System.Windows.Forms.DialogResult.OK, Dock = System.Windows.Forms.DockStyle.Bottom, Height = 36 };
                form.Controls.AddRange(new System.Windows.Forms.Control[] { list, ok });
                form.AcceptButton = ok;
                if (form.ShowDialog() != System.Windows.Forms.DialogResult.OK || list.SelectedItem == null)
                    return Result.Cancelled;

                var sel = (Row)list.SelectedItem;
                // 先用 UniqueId（最穩），失敗退回用名稱找
                Autodesk.Revit.DB.View target = null;
                if (!string.IsNullOrWhiteSpace(sel.ViewUid))
                {
                    try { target = doc.GetElement(sel.ViewUid) as Autodesk.Revit.DB.View; } catch { }
                }
                if (target == null)
                {
                    target = new FilteredElementCollector(doc)
                        .OfClass(typeof(Autodesk.Revit.DB.View))
                        .Cast<Autodesk.Revit.DB.View>()
                        .FirstOrDefault(v => !v.IsTemplate && v.Name == sel.ViewName);
                }
                if (target == null)
                {
                    TaskDialog.Show("RFI", $"找不到視圖「{sel.ViewName}」（模型不是當初那個？）。");
                    return Result.Failed;
                }
                try
                {
                    uiDoc.ActiveView = target;
                    return Result.Succeeded;
                }
                catch (Exception ex)
                {
                    TaskDialog.Show("RFI", "切視圖失敗：" + ex.Message);
                    return Result.Failed;
                }
            }
        }

        // 最小 JSON 解析：只抽 id/title/view_name/revit_view_id（不引 JSON 套件，學生好懂）
        private static List<Row> ParseRows(string json)
        {
            var out_ = new List<Row>();
            foreach (Match m in Regex.Matches(json, @"\{[^{}]*\}"))
            {
                string s = m.Value;
                out_.Add(new Row
                {
                    Id = long.TryParse(Str(s, "id"), out long n) ? n : 0,
                    Title = Str(s, "title"),
                    ViewName = Str(s, "view_name"),
                    ViewUid = Str(s, "revit_view_id"),
                });
            }
            return out_;
        }

        private static string Str(string s, string key)
        {
            var m = Regex.Match(s, "\"" + key + "\"\\s*:\\s*\"([^\"]*)\"");
            if (m.Success) return m.Groups[1].Value;
            m = Regex.Match(s, "\"" + key + "\"\\s*:\\s*(\\d+)");
            return m.Success ? m.Groups[1].Value : "";
        }
    }
}
