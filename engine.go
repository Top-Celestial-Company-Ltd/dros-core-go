package dros

type DrosProcessResult struct {
	Matches         []DrosMatch
	ActiveNeighbors []ActiveNeighbor
	ContextPrompt   string
}

type DrosEngine struct {
	Weaver *DrosWeaver
}

func NewDrosEngine(manifest DrosManifest) *DrosEngine {
	return &DrosEngine{
		Weaver: NewDrosWeaver(manifest),
	}
}

// Process 核心集成管道 (Weave -> Navigate -> Compile)
func (de *DrosEngine) Process(text string, mode string, decayFactor float64) DrosProcessResult {
	matches := de.Weaver.Weave(text)

	navigator := NewDrosNavigator(de.Weaver)
	activeNeighbors := navigator.Navigate(matches, decayFactor)

	guardVM := NewDrosGuardVM(de.Weaver)
	contextPrompt := guardVM.Compile(matches, activeNeighbors, mode)

	return DrosProcessResult{
		Matches:         matches,
		ActiveNeighbors: activeNeighbors,
		ContextPrompt:   contextPrompt,
	}
}
