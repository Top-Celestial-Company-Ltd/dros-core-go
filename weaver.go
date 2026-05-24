package dros

import (
	"bufio"
	"fmt"
	"regexp"
	"strings"
)

type trieNode struct {
	children map[rune]*trieNode
	nodeID   string // 記錄詞尾對應的 T-Number
}

type DrosWeaver struct {
	root    *trieNode
	nodeMap map[string]DrosNode
	TCoords *SafeCoordinateMap
}

func NewDrosWeaver(manifest DrosManifest) *DrosWeaver {
	weaver := &DrosWeaver{
		root:    &trieNode{children: make(map[rune]*trieNode)},
		nodeMap: manifest.Nodes,
		TCoords: NewSafeCoordinateMap(),
	}
	weaver.buildTrie()
	return weaver
}

// buildTrie 構建字典樹
func (dw *DrosWeaver) buildTrie() {
	for nodeID, node := range dw.nodeMap {
		dw.insert(node.Canonical, nodeID)
		for _, alias := range node.Aliases {
			dw.insert(alias, nodeID)
		}
	}
}

func (dw *DrosWeaver) insert(word string, nodeID string) {
	cleaned := strings.TrimSpace(word)
	if len(cleaned) == 0 {
		return
	}

	current := dw.root
	for _, c := range cleaned {
		next, exists := current.children[c]
		if !exists {
			next = &trieNode{children: make(map[rune]*trieNode)}
			current.children[c] = next
		}
		current = next
	}
	current.nodeID = nodeID
}

// Weave 核心最長匹配掃描演算法 (LMF - Longest Match First)
// 時間複雜度：O(N)
func (dw *DrosWeaver) Weave(text string) []DrosMatch {
	var matches []DrosMatch
	
	// [Feature] T-Coordinate 直連檢索 (+40.0 溢價分)
	// 使用 RE2 正則尋找經文座標，並透過長度轉換對齊 Rune Index
	tRegex := regexp.MustCompile(`\b(T\d{4}[A-Za-z]?)\b`)
	tMatchIndices := tRegex.FindAllStringSubmatchIndex(text, -1)
	if len(tMatchIndices) > 0 {
		for _, indices := range tMatchIndices {
			if len(indices) >= 4 {
				nodeID := text[indices[2]:indices[3]]
				
				dw.TCoords.Mu.RLock()
				cachedScore, inCache := dw.TCoords.Cache[nodeID]
				dw.TCoords.Mu.RUnlock()
				
				score := 40.0
				if inCache {
					score = cachedScore
				}
				
				if _, exists := dw.nodeMap[nodeID]; exists {
					rStart := len([]rune(text[:indices[2]]))
					rEnd := len([]rune(text[:indices[3]]))
					matches = append(matches, DrosMatch{
						NodeID:      nodeID,
						StartIndex:  rStart,
						EndIndex:    rEnd,
						MatchedText: nodeID,
						Score:       score,
					})
				}
			}
		}
	}

	runes := []rune(text)
	length := len(runes)
	i := 0

	for i < length {
		current := dw.root
		var longestMatchNodeID string
		longestMatchLength := 0

		for j := i; j < length; j++ {
			c := runes[j]
			next, exists := current.children[c]
			if !exists {
				break
			}
			current = next
			if current.nodeID != "" {
				longestMatchNodeID = current.nodeID
				longestMatchLength = j - i + 1
			}
		}

		if longestMatchNodeID != "" && longestMatchLength > 0 {
			matchedText := string(runes[i : i+longestMatchLength])
			matches = append(matches, DrosMatch{
				NodeID:      longestMatchNodeID,
				StartIndex:  i,
				EndIndex:    i + longestMatchLength,
				MatchedText: matchedText,
			})
			i += longestMatchLength // 步進已匹配長度
		} else {
			i++
		}
	}

	return matches
}

func (dw *DrosWeaver) GetNode(nodeID string) (DrosNode, bool) {
	node, exists := dw.nodeMap[nodeID]
	return node, exists
}

func (dw *DrosWeaver) GetAllNodes() []DrosNode {
	nodes := make([]DrosNode, 0, len(dw.nodeMap))
	for _, node := range dw.nodeMap {
		nodes = append(nodes, node)
	}
	return nodes
}

// ExtractAdmonitionSafe 物理平移掃描防截斷 (取代 RE2 貪婪正則)
// 專門針對 `> [!QUOTE]` 等多行引用，解決 RE2 不支援 `(?=...)` 前瞻斷言的缺陷
func ExtractAdmonitionSafe(text string, admonitionType string) ([]string, error) {
	if len(text) == 0 {
		return nil, fmt.Errorf("empty text provided to Admonition extractor")
	}

	var results []string
	var current strings.Builder
	inAdmonition := false
	prefix := "> [!" + admonitionType + "]"

	scanner := bufio.NewScanner(strings.NewReader(text))
	for scanner.Scan() {
		line := scanner.Text()
		trimmedLine := strings.TrimSpace(line)
		
		if strings.HasPrefix(line, prefix) {
			if inAdmonition {
				results = append(results, strings.TrimRight(current.String(), "\n \t\r"))
				current.Reset()
			}
			inAdmonition = true
			current.WriteString(line + "\n")
		} else if inAdmonition {
			// 如果是 `>` 開頭的合法延續段落，或者是空白行，皆將其吞噬
			if strings.HasPrefix(trimmedLine, ">") || trimmedLine == "" {
				current.WriteString(line + "\n")
			} else {
				results = append(results, strings.TrimRight(current.String(), "\n \t\r"))
				current.Reset()
				inAdmonition = false
			}
		}
	}
	if inAdmonition {
		results = append(results, strings.TrimRight(current.String(), "\n \t\r"))
	}
	return results, nil
}

