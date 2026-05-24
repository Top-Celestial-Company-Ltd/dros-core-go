package dros

import (
	"strings"
	"sync"
	"testing"
)

func getMockManifest() DrosManifest {
	nodes := make(map[string]DrosNode)

	// Node T0001
	weightsT0001 := make(map[string]float64)
	weightsT0001["tiantai"] = 0.95
	weightsT0001["yogacara"] = 0.90

	nodes["T0001"] = DrosNode{
		ID:         "T0001",
		Canonical:  "真如",
		Aliases:    []string{"如如", "法性", "Suchness"},
		Weights:    weightsT0001,
		Definition: "一切諸法之真實本性，非虛妄、非變異，離言絕慮之究竟實在。",
		Synapses: []DrosSynapse{
			{
				Target:   "T0002",
				Relation: "等同",
				Weight:   1.0,
			},
			{
				Target:   "T0003",
				Relation: "依止",
				Weight:   0.8,
			},
		},
	}

	// Node T0002
	weightsT0002 := make(map[string]float64)
	weightsT0002["tiantai"] = 0.90
	nodes["T0002"] = DrosNode{
		ID:         "T0002",
		Canonical:  "實相",
		Aliases:    []string{"真諦", "Ultimate Reality"},
		Weights:    weightsT0002,
		Definition: "諸法之真實相狀，無生無滅，離一切虛妄之虛空常住。",
		Synapses:   []DrosSynapse{},
	}

	// Node T0003
	weightsT0003 := make(map[string]float64)
	weightsT0003["tiantai"] = 0.98
	nodes["T0003"] = DrosNode{
		ID:         "T0003",
		Canonical:  "般若波羅蜜多",
		Aliases:    []string{"般若", "大智慧"},
		Weights:    weightsT0003,
		Definition: "能度脫生死彼岸之究竟大智慧，照見五蘊皆空。",
		Synapses:   []DrosSynapse{},
	}

	return DrosManifest{
		Version: "7.0.test",
		Nodes:   nodes,
	}
}

func TestDrosCoreGoPipeline(t *testing.T) {
	manifest := getMockManifest()
	engine := NewDrosEngine(manifest)

	// 1. 測試 O(1) 內存尋址
	nodeT0001, exists := engine.Weaver.GetNode("T0001")
	if !exists || nodeT0001.Canonical != "真如" {
		t.Errorf("FAIL: O(1) node retrieval failed for T0001")
	}
	if _, exists99 := engine.Weaver.GetNode("T9999"); exists99 {
		t.Errorf("FAIL: retrieved non-existent node T9999")
	}
	t.Log("PASS: O(1) address retrieval successful.")

	// 2. 測試最長匹配優先滑動掃描 (LMF Scanner)
	inputText := "當知「般若波羅蜜多」即是諸法之「如如」之相，不可言說。"
	result := engine.Process(inputText, "vajra", 0.5)

	hasPrajnaParamita := false
	hasPrajnaShort := false
	hasRuru := false

	for _, m := range result.Matches {
		if m.MatchedText == "般若波羅蜜多" {
			hasPrajnaParamita = true
		}
		if m.MatchedText == "般若" {
			hasPrajnaShort = true
		}
		if m.MatchedText == "如如" {
			hasRuru = true
		}
	}

	if !hasPrajnaParamita {
		t.Errorf("FAIL: missed long word '般若波羅蜜多'")
	}
	if hasPrajnaShort {
		t.Errorf("FAIL: should have melted short word '般若'")
	}
	if !hasRuru {
		t.Errorf("FAIL: missed alias '如如'")
	}
	t.Log("PASS: LMF scanner successfully matched and melted overlaps.")

	// 3. 測試拓撲鄰居與 Decay 權重衰減
	if len(result.ActiveNeighbors) != 1 {
		t.Errorf("FAIL: expected exactly 1 neighbor, got %d", len(result.ActiveNeighbors))
	} else {
		n := result.ActiveNeighbors[0]
		if n.NodeID != "T0002" {
			t.Errorf("FAIL: expected T0002 neighbor, got %s", n.NodeID)
		}
		if n.Weight != 0.5 {
			t.Errorf("FAIL: expected decayed weight 0.5, got %f", n.Weight)
		}
	}
	t.Log("PASS: topology neighbor navigation and decay weight correct.")

	// 4. 測試 GuardVM 金剛模式 ( Vajra - Strict)
	if !strings.Contains(result.ContextPrompt, "DROS 拓撲義理網格") {
		t.Errorf("FAIL: missing grid header")
	}
	if !strings.Contains(result.ContextPrompt, "一切諸法之真實本性") {
		t.Errorf("FAIL: missing definition of T0001")
	}
	if strings.Contains(result.ContextPrompt, "關聯拓撲鄰居") {
		t.Errorf("FAIL: Vajra contract must not contain synaptic neighbors")
	}
	if !strings.Contains(result.ContextPrompt, "金剛契約已生效") {
		t.Errorf("FAIL: missing Vajra enforcement command")
	}
	t.Log("PASS: GuardVM Vajra Prompt strict enforcement.")

	// 5. 測試 GuardVM 菩薩模式 ( Prajna - Interpretive)
	prajnaResult := engine.Process(inputText, "prajna", 0.5)
	if !strings.Contains(prajnaResult.ContextPrompt, "關聯拓撲鄰居") {
		t.Errorf("FAIL: Prajna contract must contain synaptic neighbors")
	}
	if !strings.Contains(prajnaResult.ContextPrompt, "實相 (T0002)") {
		t.Errorf("FAIL: missing active neighbor representation")
	}
	if !strings.Contains(prajnaResult.ContextPrompt, "菩薩契約已生效") {
		t.Errorf("FAIL: missing Prajna command")
	}
	t.Log("PASS: GuardVM Prajna Prompt warm interpretation.")
}

func TestConcurrentMapAndAdmonition(t *testing.T) {
	manifest := getMockManifest()
	manifest.Nodes["T0262"] = DrosNode{ID: "T0262", Canonical: "妙法蓮華經"}
	engine := NewDrosEngine(manifest)

	// 1. 1000 Goroutine 高頻撞擊壓測 (500 Write / 500 Read)
	var wg sync.WaitGroup
	tCoords := engine.Weaver.TCoords

	for i := 0; i < 1000; i++ {
		wg.Add(1)
		if i%2 == 0 {
			// Write Goroutine: 模擬大模型接口更新座標
			go func() {
				defer wg.Done()
				tCoords.Mu.Lock()
				tCoords.Cache["T0262"] = 40.0
				tCoords.Mu.Unlock()
			}()
		} else {
			// Read Goroutine: 模擬高頻 Weave 掃描尋址
			go func() {
				defer wg.Done()
				engine.Process("查詢 T0262", "vajra", 0.5)
			}()
		}
	}
	wg.Wait()
	t.Log("PASS: SafeCoordinateMap (1000 Goroutines) sustained high-concurrency without DATA RACE.")

	// 2. 測試 ExtractAdmonitionSafe 針對《成唯識論》極端原典的擷取
	text := `前言文本，無關緊要

> [!QUOTE]
> 然《成唯識論》有云：
> 「諸心心所。若細若麤。若相若見。」
> 
>   (不規則的縮排空行)
> 乃至阿賴耶識，亦復如是。

中斷文本
> [!NOTE]
> 這裡是一般的筆記`

	quotes, err := ExtractAdmonitionSafe(text, "QUOTE")
	if err != nil {
		t.Fatalf("ExtractAdmonitionSafe returned error: %v", err)
	}
	if len(quotes) != 1 || !strings.Contains(quotes[0], "成唯識論") || !strings.Contains(quotes[0], "乃至阿賴耶識") {
		t.Errorf("FAIL: Failed to extract complex QUOTE block correctly. Got: %v", quotes)
	}

	notes, err := ExtractAdmonitionSafe(text, "NOTE")
	if err != nil {
		t.Fatalf("ExtractAdmonitionSafe returned error: %v", err)
	}
	if len(notes) != 1 || !strings.Contains(notes[0], "這裡是一般的筆記") {
		t.Errorf("FAIL: Failed to extract NOTE block correctly. Got: %v", notes)
	}
	t.Log("PASS: Complex Cheng Weishi Lun Admonition extraction worked perfectly.")
}
