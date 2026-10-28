using System;
using System.Collections.Generic;
using System.Linq;
using LebrCarbon;

namespace LebrDatabaseTests
{
    /// <summary>結構碳排估算測試：附錄一讀取、CFsn 手算、地上與地下分類、迴圈與錯誤處理。</summary>
    internal static partial class Program
    {
        static partial void TestEstimate(string csvPath)
        {
            Section("估算：附錄一讀取");
            var table = PLccTable.Load(csvPath);
            Equal("附錄一筆數", 120, table.Materials.Count);
            Near("PLCC-040 預拌混凝土（2000psi）總碳排", 243.78, table.Get("PLCC-040").TotalA1A4.Value);
            Equal("PLCC-040 單位", "m3", table.Get("PLCC-040").Unit);
            Near("PLCC-003 鋼筋及鐵件總碳排", 1.15, table.Get("PLCC-003").TotalA1A4.Value);
            Equal("PLCC-003 單位", "kg", table.Get("PLCC-003").Unit);
            Equal("PLCC-013 有手冊說明", true, table.Get("PLCC-013").ManualIssue != null);
            Throws("不存在的材料編號報錯", () => table.Get("PLCC-999"));

            var est = new StructureEstimator(table);

            Section("估算：單層手算");
            var r1 = est.Calculate(new[]
            {
                Item("StructuralColumns", "Level 1", 0, 10, "m3", "PLCC-040", false),
                Item("Rebar", "Level 1", 0, 1000, "kg", "PLCC-003", true),
            });
            Near("地上 CFs = 10×243.78 + 1000×1.15", 3587.8, r1.Above);
            Near("地下 CFs = 0", 0, r1.Below);
            Near("合計 CFs", 3587.8, r1.CFs);
            Near("Level 1 CFc", 2437.8, r1.Levels[0].CFc);
            Near("Level 1 CFr", 1150, r1.Levels[0].CFr);
            Equal("沒有缺鋼筋註明", 0, r1.Notes.Count);

            Section("估算：地下層與無鋼筋");
            var r2 = est.Calculate(new[]
            {
                Item("StructuralColumns", "B1", -4, 5, "m3", "PLCC-040", false),
            });
            Near("地下 CFs = 5×243.78", 1218.9, r2.Below);
            Near("地上 CFs = 0", 0, r2.Above);
            Equal("無鋼筋樓層有註明", true, r2.Notes.Any(n => n.Contains("B1") && n.Contains("Bn = 0")));

            Section("估算：標高邊界（0 為地上，-0.001 為地下）");
            var r3 = est.Calculate(new[]
            {
                Item("Floors", "Level 0", 0, 1, "m3", "PLCC-040", false),
                Item("Floors", "B0", -0.001, 1, "m3", "PLCC-040", false),
            });
            Near("標高 0 歸地上", 243.78, r3.Above);
            Near("標高 -0.001 歸地下", 243.78, r3.Below);

            Section("估算：迴圈（零次、一次、多次）");
            var r4 = est.Calculate(new List<EstimateItem>());
            Near("零次：CFs = 0", 0, r4.CFs);
            Equal("零次：有註明", true, r4.Notes.Any(n => n.Contains("CFs = 0")));

            var loop = new List<EstimateItem>();
            for (var i = 0; i < 76; i++)
                loop.Add(Item("StructuralColumns", "Level 1", 0, 1.8, "m3", "PLCC-040", false));
            var r5 = est.Calculate(loop);
            Near("76 根柱（每根 1.8 m3）合計", 76 * 1.8 * 243.78, r5.CFs);
            Equal("76 根柱仍為一層", 1, r5.Levels.Count);

            Section("估算：錯誤輸入直接報錯");
            Throws("混凝土數量單位不是 m3 報錯", () => est.Calculate(new[] { Item("StructuralColumns", "Level 1", 0, 10, "kg", "PLCC-040", false) }));
            Throws("混凝土係數用 kg 報錯", () => est.Calculate(new[] { Item("StructuralColumns", "Level 1", 0, 10, "m3", "PLCC-003", false) }));
            Throws("鋼筋係數用 m3 報錯", () => est.Calculate(new[] { Item("Rebar", "Level 1", 0, 10, "kg", "PLCC-040", true) }));
            Throws("負數量報錯", () => est.Calculate(new[] { Item("StructuralColumns", "Level 1", 0, -1, "m3", "PLCC-040", false) }));
            Throws("同名樓層標高不一致報錯", () => est.Calculate(new[]
            {
                Item("StructuralColumns", "Level 1", 0, 1, "m3", "PLCC-040", false),
                Item("Floors", "Level 1", 4, 1, "m3", "PLCC-040", false),
            }));
            Throws("不存在的材料編號報錯", () => est.Calculate(new[] { Item("Floors", "Level 1", 0, 1, "m3", "PLCC-999", false) }));
        }

        private static EstimateItem Item(string source, string level, double elevation, double quantity, string unit, string materialId, bool isRebar)
        {
            return new EstimateItem
            {
                Source = source,
                Level = level,
                ElevationM = elevation,
                Quantity = quantity,
                Unit = unit,
                MaterialId = materialId,
                IsRebar = isRebar,
            };
        }
    }
}
