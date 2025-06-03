package lib

import (
	"fmt"
	"strings"
)

func (v *ViewDefinition) IsValidGroupBy(keys []string) bool {
	if len(keys) == 0 {
		return false
	}

	// Build map of key -> index
	indexMap := make(map[string]int)
	indexMap[ShardLabel] = 0
	for i, k := range v.KeyOrder {
		indexMap[k] = i + 1
	}

	prevIndex := -1
	for _, k := range keys {
		idx, ok := indexMap[k]
		if !ok {
			return false
		}
		if idx <= prevIndex {
			return false
		}
		prevIndex = idx
	}

	return true
}

func (v *ViewInstance) GroupBy(keys []string) (map[string]int, error) {
	return v.GroupByWithShards(keys, v.Shards)
}

func (v *ViewInstance) GroupByWithShards(keys []string, shardsToGroup []*Shard) (map[string]int, error) {
	if !v.Mapper.d.IsValidGroupBy(keys) {
		keyOrders := append([]string{ShardLabel}, v.Mapper.d.KeyOrder...)
		return nil, fmt.Errorf("invalid group by key order: got %v expected %v", keys, keyOrders)
	}

	results := make(map[string]int)

	for _, shard := range shardsToGroup {
		prefix := ""
		groupKeys := keys
		parts := []string{}

		if len(keys) > 0 && keys[0] == ShardLabel {
			prefix = shard.Key.Format(v.Mapper.layout)
			groupKeys = keys[1:]
			parts = append(parts, prefix)
		}

		// Grouping by shard only
		if len(groupKeys) == 0 {
			results[prefix] = shard.Count
			continue
		}

		for _, root := range shard.Roots {
			traverseForGroupBy(root, groupKeys, 0, parts, results)
		}
	}

	return results, nil
}

func traverseForGroupBy(node *Node, groupKeys []string, level int, parts []string, result map[string]int) {
	if level >= len(groupKeys) {
		return // done collecting parts, but we should’ve already added result earlier
	}

	targetKey := groupKeys[level]

	if node.Key == targetKey {
		parts = append(parts, node.Value)

		if level == len(groupKeys)-1 {
			key := strings.Join(parts, ".")
			result[key] += node.Count
			return
		}

		// Recurse to next level only after consuming this groupKey
		for _, child := range node.Nodes {
			traverseForGroupBy(child, groupKeys, level+1, parts, result)
		}
	} else {
		// Keep looking deeper to match current groupKey
		for _, child := range node.Nodes {
			traverseForGroupBy(child, groupKeys, level, parts, result)
		}
	}
}
