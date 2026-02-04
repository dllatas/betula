package lib

import "slices"

func (v *ViewInstance) Filter(match map[string][]string) ([]*Shard, error) {
	return v.FilterWithShards(match, v.Shards)
}

func (v *ViewInstance) FilterWithShards(match map[string][]string, shardsToFilter []*Shard) ([]*Shard, error) {
	if len(match) == 0 {
		return shardsToFilter, nil
	}

	resulting := shardsToFilter

	for _, orderKey := range v.Mapper.D.KeyOrder {
		matches, found := match[orderKey]
		if !found {
			continue
		}

		var next []*Shard

		for _, shard := range resulting {
			newRoots := map[string]*Node{}
			totalCount := 0

			for _, root := range shard.Roots {
				n := filterNodeRecursive(root, orderKey, matches)
				if n != nil {
					newRoots[shardMapKey(n.Key, n.Value)] = n
					totalCount += n.Count
				}
			}

			if len(newRoots) > 0 {
				next = append(next, &Shard{
					Key:   shard.Key,
					Roots: newRoots,
					Count: totalCount,
				})
			}
		}

		resulting = next
	}

	return resulting, nil
}

func filterNodeRecursive(n *Node, filterKey string, allowed []string) *Node {
	isMatchKey := n.Key == filterKey
	isAllowed := slices.Contains(allowed, n.Value)

	if isMatchKey && !isAllowed {
		return nil
	}

	if isMatchKey && isAllowed {
		return n
	}

	newNode := newNode(n.Key, n.Value)

	for _, child := range n.Nodes {
		fNode := filterNodeRecursive(child, filterKey, allowed)
		if fNode != nil {
			newNode.Nodes[shardMapKey(fNode.Key, fNode.Value)] = fNode
			newNode.Count += fNode.Count
		}
	}

	if len(newNode.Nodes) > 0 {
		return newNode
	}

	return nil
}
