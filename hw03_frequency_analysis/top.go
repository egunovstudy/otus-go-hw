package hw03frequencyanalysis

import (
	"cmp"
	"slices"
	"strings"
)

func Top10(input string) []string {
	dict := make(map[string]Node)

	analyzeStrFrequence(input, dict)

	nodes := transformFrequencyMapToSortedSlice(dict)

	result := fulfillResult(nodes)

	return checkResultLengthAndReturnFinal(result)
}

func checkResultLengthAndReturnFinal(result []string) []string {
	if len(result) < 10 {
		return result
	}
	return result[0:10]
}

func fulfillResult(nodes []Node) []string {
	result := []string{}
	for idx, node := range nodes {
		if idx == 10 {
			break
		}
		result = append(result, node.name)
	}
	return result
}

func transformFrequencyMapToSortedSlice(dict map[string]Node) []Node {
	nodes := make([]Node, 0, len(dict))
	for _, node := range dict {
		nodes = append(nodes, node)
	}
	slices.SortFunc(nodes, func(a, b Node) int {
		if a.count == b.count {
			return strings.Compare(a.name, b.name)
		}
		return cmp.Compare(b.count, a.count)
	})
	return nodes
}

func analyzeStrFrequence(input string, dict map[string]Node) {
	split := strings.Fields(input)
	for _, str := range split {
		val, exists := dict[str]
		if exists {
			val.count++
			dict[str] = val
		} else {
			dict[str] = Node{name: str, count: 1}
		}
	}
}

type Node struct {
	name  string
	count int
}
