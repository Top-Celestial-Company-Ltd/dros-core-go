package dros

import (
	"fmt"
	"strings"
)

type DrosGuardVM struct {
	weaver *DrosWeaver
}

func NewDrosGuardVM(weaver *DrosWeaver) *DrosGuardVM {
	return &DrosGuardVM{weaver: weaver}
}

// Compile 將內存拓撲網格編譯為大模型提示詞 Context
func (dg *DrosGuardVM) Compile(matches []DrosMatch, activeNeighbors []ActiveNeighbor, mode string) string {
	var builder strings.Builder

	builder.WriteString("<!-- DROS_SOVEREIGN_CONTEXT_START -->\n")
	builder.WriteString("## 📿 DROS 拓撲義理網格 (Sovereign Context Grid)\n")
	builder.WriteString("當前文本中已成功編織以下義理突觸：\n\n")

	// 1. 輸出核心名相定義
	builder.WriteString("### 核心名相定義 (Canonical Core Nodes)\n")
	if len(matches) == 0 {
		builder.WriteString("- *無直接匹配核心名相*\n")
	} else {
		seenCoreIDs := make(map[string]bool)
		for _, m := range matches {
			if seenCoreIDs[m.NodeID] {
				continue
			}
			seenCoreIDs[m.NodeID] = true

			if node, exists := dg.weaver.GetNode(m.NodeID); exists {
				builder.WriteString(fmt.Sprintf("- **%s (%s)**：%s\n", node.Canonical, node.ID, node.Definition))
			}
		}
	}
	builder.WriteString("\n")

	// 2. 輸出關聯拓撲鄰居 (僅在 Prajna 模式下輸出)
	if mode == "prajna" {
		builder.WriteString("### 關聯拓撲鄰居 (Active Synaptic Neighbors)\n")
		if len(activeNeighbors) == 0 {
			builder.WriteString("- *無關聯拓撲鄰居*\n")
		} else {
			for _, neighbor := range activeNeighbors {
				sourceName := neighbor.SourceNodeID
				if sourceNode, exists := dg.weaver.GetNode(neighbor.SourceNodeID); exists {
					sourceName = sourceNode.Canonical
				}
				builder.WriteString(fmt.Sprintf(
					"- **%s (%s)** (共鳴權重: %.2f)：與 [%s] 具有 [%s] 關係。\n",
					neighbor.Canonical, neighbor.NodeID, neighbor.Weight, sourceName, neighbor.Relation,
				))
			}
		}
		builder.WriteString("\n")
	}

	// 3. 輸出合約熔斷指令
	builder.WriteString(fmt.Sprintf("### 推理合約熔斷規則 (GuardVM Execution Mode: %s)\n", strings.ToUpper(mode)))
	if mode == "vajra" {
		builder.WriteString("[金剛契約已生效]：你的一切推論必須 100% 侷限在上述給定的【核心名相定義】中。你必須保持極致的學術客觀，逐字對齊定義，不得添加 any 未經定義的宗教發揮或主觀推演。如果用戶的問題超出了上述定義的範疇，你必須坦誠回答「非本合約所及」。\n")
	} else {
		builder.WriteString("[菩薩契約已生效]：在立足於【核心名相定義】的前提下，你可以沿著【關聯拓撲鄰居】所勾勒的突觸網格（特別是各鄰居之間的關係，如等同、依止、生起等），進行溫和、融會貫通的跨學科或現代化義理詮釋。請結合當前筆記內容，引導用戶體會空性與智慧的隨流運用。\n")
	}

	builder.WriteString("<!-- DROS_SOVEREIGN_CONTEXT_END -->")

	return builder.String()
}
