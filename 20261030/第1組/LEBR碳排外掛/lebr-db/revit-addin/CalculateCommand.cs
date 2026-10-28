using System;
using System.Collections.Generic;
using System.IO;
using System.Linq;
using System.Reflection;
using System.Windows.Interop;
using Autodesk.Revit.Attributes;
using Autodesk.Revit.DB;
using Autodesk.Revit.UI;
using LebrCarbon.Revit;
using LebrCarbon.UI;
using LebrCarbon;

namespace LebrCarbon
{
    /// <summary>「開始計算」：計算前檢查 → 計算 → 結果視窗。模型不會被修改。</summary>
    [Transaction(TransactionMode.ReadOnly)]
    public class CalculateCommand : IExternalCommand
    {
        private static LebrDatabase _database;

        public Result Execute(ExternalCommandData commandData, ref string message, ElementSet elements)
        {
            var folder = Path.GetDirectoryName(Assembly.GetExecutingAssembly().Location);
            LebrDatabase db;
            try
            {
                db = _database ?? (_database = LebrDatabase.Load(Path.Combine(folder, "lebr_2025.json")));
            }
            catch (Exception ex)
            {
                TaskDialog.Show("LEBR", "無法讀取 LEBR 資料庫 lebr_2025.json：\n" + ex.Message);
                return Result.Failed;
            }

            var uiApp = commandData.Application;
            var doc = uiApp.ActiveUIDocument?.Document;
            var calculator = new LebrCalculator(db);
            Func<QuantityInput> readModel = null;
            if (doc != null && !doc.IsFamilyDocument)
                readModel = () => new RevitQuantityProvider(doc, db).Read();
            var precheck = new PrecheckWindow(calculator.Calculate, Path.Combine(folder, "samples", "appendix3_quantities.json"),
                                              readModel, doc?.Title);
            new WindowInteropHelper(precheck).Owner = uiApp.MainWindowHandle;
            if (precheck.ShowDialog() != true || precheck.Result == null)
                return Result.Cancelled;

            var uiDoc = uiApp.ActiveUIDocument;
            var window = new ResultWindow(precheck.Result, ids => SelectElements(uiDoc, ids));
            new WindowInteropHelper(window).Owner = uiApp.MainWindowHandle;
            window.ShowDialog();
            return Result.Succeeded;
        }

        /// <summary>在模型中選取明細列對應的元件；範例資料沒有元件 ID 時不動作。</summary>
        private static bool SelectElements(UIDocument uiDoc, IList<long> ids)
        {
            if (uiDoc == null || ids == null || ids.Count == 0) return false;
            var existing = ids.Select(i => new ElementId(i)).Where(id => uiDoc.Document.GetElement(id) != null).ToList();
            if (existing.Count == 0) return false;
            uiDoc.Selection.SetElementIds(existing);
            uiDoc.ShowElements(existing);
            return true;
        }
    }
}
