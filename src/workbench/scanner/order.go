package scanner

import (
	"sort"

	"hatch/core/config"
)

// ApplyOrder reorders scripts according to a saved path list; scripts missing
// from order keep their scan order and are appended at the end.
func ApplyOrder(scripts []config.ScriptMeta, order []string) []config.ScriptMeta {
	rank := make(map[string]int, len(order))
	for i, p := range order {
		rank[p] = i
	}
	result := make([]config.ScriptMeta, len(scripts))
	copy(result, scripts)
	sort.SliceStable(result, func(i, j int) bool {
		ri, iok := rank[result[i].Path]
		rj, jok := rank[result[j].Path]
		if iok && jok {
			return ri < rj
		}
		return iok
	})
	return result
}
