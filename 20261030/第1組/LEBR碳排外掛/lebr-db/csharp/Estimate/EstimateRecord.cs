using System;
using System.Collections.Generic;
using System.Globalization;
using Newtonsoft.Json.Linq;

namespace LebrCarbon
{
    /// <summary>
    /// 估算紀錄（LEBR_EstimateJson，介面規格第三節）。只組成 JSON 字串與讀回摘要，不依賴 Revit API。
    /// </summary>
    public static class EstimateRecord
    {
        public const int Version = 1;

        public static string ToJson(EstimateResult result, IDictionary<string, string> materialBySource, string modelName, DateTime calculatedAt)
        {
            if (result == null) throw new ArgumentNullException("result");

            var levels = new JArray();
            foreach (var l in result.Levels)
            {
                levels.Add(new JObject
                {
                    ["name"] = l.Level,
                    ["elevation_m"] = l.ElevationM,
                    ["below_ground"] = l.IsBelowGround,
                    ["concrete_m3"] = l.ConcreteM3,
                    ["rebar_kg"] = l.RebarKg,
                    ["cfc"] = l.CFc,
                    ["cfr"] = l.CFr,
                    ["cfs_n"] = l.CFsn,
                });
            }

            var selections = new JArray();
            if (materialBySource != null)
            {
                foreach (var pair in materialBySource)
                    selections.Add(new JObject { ["source"] = pair.Key, ["material_id"] = pair.Value });
            }

            var notes = new JArray();
            foreach (var n in result.Notes) notes.Add(n);

            var root = new JObject
            {
                ["version"] = Version,
                ["calculated_at"] = calculatedAt.ToString("yyyy-MM-ddTHH:mm:ss", CultureInfo.InvariantCulture),
                ["model"] = modelName ?? "",
                ["levels"] = levels,
                ["selections"] = selections,
                ["rebar_sources"] = new JObject
                {
                    ["structure_rebar"] = "Rebar（主體為結構柱、結構構架）＋ OST_AreaRein／OST_PathRein 系統內鋼筋（樓板）",
                    ["mass_method"] = "體積 × 7850 kg/m3",
                },
                ["result"] = new JObject
                {
                    ["above"] = result.Above,
                    ["below"] = result.Below,
                    ["cfs"] = result.CFs,
                },
                ["notes"] = notes,
            };
            return root.ToString(Newtonsoft.Json.Formatting.Indented);
        }

        /// <summary>從 LEBR_EstimateJson 讀回地上、地下與合計；格式不符時回傳 null。</summary>
        public static bool TryReadSummary(string json, out double above, out double below, out double cfs)
        {
            above = below = cfs = 0;
            if (string.IsNullOrWhiteSpace(json)) return false;
            try
            {
                var result = JObject.Parse(json)["result"] as JObject;
                if (result == null) return false;
                above = result.Value<double>("above");
                below = result.Value<double>("below");
                cfs = result.Value<double>("cfs");
                return true;
            }
            catch (Newtonsoft.Json.JsonException)
            {
                return false;
            }
        }
    }
}
