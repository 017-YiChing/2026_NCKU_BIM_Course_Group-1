using System;
using System.Globalization;
using System.Reflection;
using System.Windows;
using System.Windows.Media;
using System.Windows.Media.Imaging;
using Autodesk.Revit.UI;
// Revit SDK 的全域 using 也有同名類別
using Color = System.Windows.Media.Color;
using FormattedText = System.Windows.Media.FormattedText;
using Point = System.Windows.Point;

namespace LebrCarbon
{
    /// <summary>在 Revit 功能區建立「LEBR」頁籤與「開始計算」按鈕。</summary>
    public class App : IExternalApplication
    {
        public const string TabName = "LEBR";

        public Result OnStartup(UIControlledApplication application)
        {
            try
            {
                application.CreateRibbonTab(TabName);
            }
            catch (Autodesk.Revit.Exceptions.ArgumentException)
            {
                // 頁籤已存在（例如重複載入），沿用即可
            }

            var panel = application.CreateRibbonPanel(TabName, "低碳建築評估");
            panel.AddItem(new PushButtonData(
                "LebrSetup", "建立\n參數", Assembly.GetExecutingAssembly().Location, typeof(Revit.SetupParametersCommand).FullName)
            {
                ToolTip = "在專案中加入 LEBR 共用參數：構件類型的 LEBR_Code、外牆本體 LEBR_Code_Body、不計入的 LEBR_Exclude，以及專案資訊的建築類別與結構參數",
                LargeImage = Icon(32, "P", Color.FromRgb(0x44, 0x54, 0x6A)),
                Image = Icon(16, "P", Color.FromRgb(0x44, 0x54, 0x6A)),
            });
            panel.AddItem(new PushButtonData(
                "LebrProjectSettings", "專案\n設定", Assembly.GetExecutingAssembly().Location, typeof(ProjectSettingsCommand).FullName)
            {
                ToolTip = "設定建築條件、主結構（含估算鎖定）、非主結構與優惠係數（LL），儲存到專案資訊的 LEBR_* 參數",
                LargeImage = Icon(32, "S", Color.FromRgb(0x2E, 0x6F, 0x9E)),
                Image = Icon(16, "S", Color.FromRgb(0x2E, 0x6F, 0x9E)),
            });
            panel.AddItem(new PushButtonData(
                "LebrCodes", "代號\nJSON", Assembly.GetExecutingAssembly().Location, typeof(Revit.CodesCommand).FullName)
            {
                ToolTip = "把這個模型填好的 LEBR 代號匯出成 JSON，或把別人的代號 JSON 套用到這個模型（會自動建立參數）",
                LargeImage = Icon(32, "{ }", Color.FromRgb(0x8A, 0x5A, 0x1F)),
                Image = Icon(16, "{ }", Color.FromRgb(0x8A, 0x5A, 0x1F)),
            });
            var button = new PushButtonData(
                "LebrCalculate", "開始\n計算", Assembly.GetExecutingAssembly().Location, typeof(CalculateCommand).FullName)
            {
                ToolTip = "依《2025 低碳（低蘊含碳）建築評估手冊》計算蘊含碳排、減碳率與等級",
                LongDescription = "先做計算前檢查（未對應元件、缺少的專案參數），再顯示結果視窗，可匯出結果 JSON 給網頁。",
                LargeImage = Icon(32, "CO₂", Color.FromRgb(0x2E, 0x7D, 0x5B)),
                Image = Icon(16, "CO₂", Color.FromRgb(0x2E, 0x7D, 0x5B)),
            };
            panel.AddItem(button);
            return Result.Succeeded;
        }

        public Result OnShutdown(UIControlledApplication application)
        {
            return Result.Succeeded;
        }

        /// <summary>以程式繪製的按鈕圖示（色塊加文字），不需額外圖檔。</summary>
        private static BitmapSource Icon(int size, string label, Color color)
        {
            var visual = new DrawingVisual();
            using (var dc = visual.RenderOpen())
            {
                var fill = new SolidColorBrush(color);
                dc.DrawRoundedRectangle(fill, null, new Rect(0, 0, size, size), size * 0.18, size * 0.18);
                var text = new FormattedText(label, CultureInfo.InvariantCulture, FlowDirection.LeftToRight,
                    new Typeface(new FontFamily("Segoe UI"), FontStyles.Normal, FontWeights.Bold, FontStretches.Normal),
                    size * (label.Length > 1 ? 0.38 : 0.6), Brushes.White, 1.0);
                dc.DrawText(text, new Point((size - text.Width) / 2, (size - text.Height) / 2));
            }
            var bmp = new RenderTargetBitmap(size, size, 96, 96, PixelFormats.Pbgra32);
            bmp.Render(visual);
            bmp.Freeze();
            return bmp;
        }
    }
}
