package main

import (
	"bufio"
	"os"
	"regexp"
	"strings"
)

// ParseIFCIdentity 第一版：不跑幾何核心，只讀 STEP 文字抽身分資料。
// 抽：GlobalId、類型(IfcWall/IfcDoor/...)、Name、所在樓層(盡力而為)。
// 回傳 elements 列表，供寫入 elements 表（定位夠用）。
func ParseIFCIdentity(path string) ([]Element, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	// 例：#123=IFCWALL('2a$xyz...',#42,'Wall-001',$,'Basic Wall',...);
	reEntity := regexp.MustCompile(`^#(\d+)\s*=\s*(IFC[A-Z0-9]+)\s*\((.*)\)\s*;`)
	reQuoted := regexp.MustCompile(`'([^']*)'`)
	// 物件化關係：#n=IFCRELCONTAINEDINSPATIALSTRUCTURE(..., (#a,#b,...), #storey,...)
	reRel := regexp.MustCompile(`IFCRELCONTAINEDINSPATIALSTRUCTURE\s*\((.*)\)`)
	reRefs := regexp.MustCompile(`#(\d+)`)

	type entInfo struct {
		ID       string
		Type     string
		GlobalID string
		Name     string
	}
	ents := map[string]*entInfo{} // #id -> info
	storeyName := map[string]string{}
	relContained := map[string]string{} // element #id -> storey #id

	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 1024*1024), 1024*1024)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "ISO-") {
			continue
		}
		m := reEntity.FindStringSubmatch(line)
		if m == nil {
			continue
		}
		num, typ, args := m[1], m[2], m[3]
		key := "#" + num
		qs := reQuoted.FindAllStringSubmatch(args, -1)

		switch typ {
		case "IFCBUILDINGSTOREY":
			nm := ""
			if len(qs) >= 3 {
				nm = qs[2][1] // 常見：第3個字串是 Name/LongName
			} else if len(qs) >= 1 {
				nm = qs[len(qs)-1][1]
			}
			if nm == "" {
				nm = key
			}
			storeyName[key] = nm
		case "IFCRELCONTAINEDINSPATIALSTRUCTURE":
			rm := reRel.FindStringSubmatch(line)
			if rm == nil {
				continue
			}
			refs := reRefs.FindAllStringSubmatch(rm[1], -1)
			// 最後一個 # 引用通常是 RelatingStructure（樓層）
			if len(refs) >= 2 {
				storeyRef := "#" + refs[len(refs)-1][1]
				for _, r := range refs[:len(refs)-1] {
					relContained["#"+r[1]] = storeyRef
				}
			}
		default:
			if !strings.HasPrefix(typ, "IFC") {
				continue
			}
			// 只收常見可定位元件，避免收關係/屬性等雜項
			if !isTrackableIFCType(typ) {
				continue
			}
			gid, name := "", ""
			if len(qs) >= 1 {
				gid = qs[0][1]
			}
			if len(qs) >= 3 {
				name = qs[2][1]
			} else if len(qs) >= 2 {
				name = qs[1][1]
			}
			if gid == "" {
				continue
			}
			ents[key] = &entInfo{ID: key, Type: typ, GlobalID: gid, Name: name}
		}
	}
	if err := sc.Err(); err != nil {
		return nil, err
	}

	out := make([]Element, 0, len(ents))
	for key, e := range ents {
		st := ""
		if sref, ok := relContained[key]; ok {
			st = storeyName[sref]
		}
		out = append(out, Element{
			GlobalID: e.GlobalID,
			Type:     e.Type,
			Name:     e.Name,
			Storey:   st,
		})
	}
	return out, nil
}

func isTrackableIFCType(t string) bool {
	switch t {
	case "IFCWALL", "IFCWALLSTANDARDCASE",
		"IFCDOOR", "IFCWINDOW", "IFCSLAB", "IFCBEAM",
		"IFCCOLUMN", "IFCSTAIR", "IFCSTAIRFLIGHT",
		"IFCROOF", "IFCCURTAINWALL", "IFCMEMBER",
		"IFCPLATE", "IFCSPACE", "IFCZONE",
		"IFCFURNISHINGELEMENT", "IFCDISTRIBUTIONELEMENT",
		"IFCBUILDINGELEMENTPROXY", "IFCPILE", "IFCFOOTING":
		return true
	default:
		return false
	}
}
