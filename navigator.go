package dros

import (
	"sort"
)

type DrosNavigator struct {
	weaver *DrosWeaver
}

func NewDrosNavigator(weaver *DrosWeaver) *DrosNavigator {
	return &DrosNavigator{weaver: weaver}
}

// Navigate 拓撲鄰域擴展導航與共鳴度累加演算法
func (dn *DrosNavigator) Navigate(matches []DrosMatch, decayFactor float64) []ActiveNeighbor {
	coreNodeIDs := make(map[string]bool)
	for _, m := range matches {
		coreNodeIDs[m.NodeID] = true
	}

	neighborMap := make(map[string]ActiveNeighbor)

	for _, m := range matches {
		coreNode, exists := dn.weaver.GetNode(m.NodeID)
		if !exists {
			continue
		}

		for _, synapse := range coreNode.Synapses {
			// 如果鄰居節點本身已經被直接匹配，過濾以防冗餘
			if coreNodeIDs[synapse.Target] {
				continue
			}

			targetNode, exists := dn.weaver.GetNode(synapse.Target)
			if !exists {
				continue
			}

			decayedWeight := synapse.Weight * decayFactor

			existing, found := neighborMap[synapse.Target]
			if found {
				// 【共鳴累加】
				existing.Weight += decayedWeight
				neighborMap[synapse.Target] = existing
			} else {
				neighborMap[synapse.Target] = ActiveNeighbor{
					NodeID:       synapse.Target,
					Canonical:    targetNode.Canonical,
					Relation:     synapse.Relation,
					SourceNodeID: m.NodeID,
					Weight:       decayedWeight,
				}
			}
		}
	}

	// 轉化為切片 (Slice)
	var result []ActiveNeighbor
	for _, neighbor := range neighborMap {
		result = append(result, neighbor)
	}

	// 依照衰減共鳴權重進行降序排序
	sort.Slice(result, func(i, j int) bool {
		return result[i].Weight > result[j].Weight
	})

	return result
}
