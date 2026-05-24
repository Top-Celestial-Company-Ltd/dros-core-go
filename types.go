package dros

/**
 * 📿 DROS Standard Type Definitions - Go Edition
 * RFC 001: Unified Multi-Language µDROS Core Specification
 */

import "sync"

// DrosSynapse 義理突觸連線
type DrosSynapse struct {
	Target   string  `json:"target"`   // 目標節點 T-Number
	Relation string  `json:"relation"` // 關係種類
	Weight   float64 `json:"weight"`   // 連線權重
}

// DrosNode 義理原子節點
type DrosNode struct {
	ID         string             `json:"id"`
	Canonical  string             `json:"canonical"`
	Aliases    []string           `json:"aliases"`
	Weights    map[string]float64 `json:"weights"`
	Definition string             `json:"definition"`
	Synapses   []DrosSynapse      `json:"synapses"`
}

// DrosManifest DROS 黃金清單 Manifest
type DrosManifest struct {
	Version  string              `json:"version"`
	Metadata map[string]interface{} `json:"metadata"`
	Nodes    map[string]DrosNode `json:"nodes"`
}

// DrosMatch 掃描匹配結構體
type DrosMatch struct {
	NodeID      string
	StartIndex  int // 在字元陣列中的起點索引 (0-based)
	EndIndex    int // 在字元陣列中的終點索引 (0-based, exclusive)
	MatchedText string
	Score       float64 // 匹配權重/溢價分 (如 T-座標的 +40.0)
}

// ActiveNeighbor 拓撲活化鄰居
type ActiveNeighbor struct {
	NodeID       string
	Canonical    string
	Relation     string
	SourceNodeID string
	Weight       float64
}

// SafeCoordinateMap 全域安全 T-Coordinate 快取結構 (防競態)
type SafeCoordinateMap struct {
	Mu    sync.RWMutex
	Cache map[string]float64
}

func NewSafeCoordinateMap() *SafeCoordinateMap {
	return &SafeCoordinateMap{
		Cache: make(map[string]float64),
	}
}
