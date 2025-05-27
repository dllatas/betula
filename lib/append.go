package lib

import "sort"

func (v *ViewInstance) Append(e Event) error {
	v.Mu.Lock()
	defer v.Mu.Unlock()

	key, err := v.Mapper.PartitionKey(e)
	if err != nil {
		return err
	}

	treeValues, err := v.Mapper.TreePath(e)
	if err != nil {
		return err
	}

	shard := &Shard{}

	// does the shard for this key exist?
	idx, found := v.ShardIndex[key]
	if !found {
		// new shard
		var errShard error
		shard, errShard = newShard(key, e.Timestamp, v.Mapper.layout)
		if errShard != nil {
			return errShard
		}

		// add shard to shards in view instance
		v.addShard(shard)
	} else {
		shard = v.Shards[idx]
	}

	// update shard counter
	shard.Count++

	// get a ref to shard nodes
	children := &shard.Roots

	for idx, treeValue := range treeValues {
		var next *Node

		orderKey := v.Mapper.d.KeyOrder[idx]
		mapKey := orderKey + ":" + treeValue

		child, found := (*children)[mapKey]
		if found {
			child.Count++
			next = child
		} else {
			next = newNode(orderKey, treeValue)
			next.Count++
			(*children)[mapKey] = next
		}

		children = &next.Nodes
	}

	return nil
}

func (v *ViewInstance) addShard(s *Shard) {
	if len(v.Shards) == 0 {
		v.Shards = append(v.Shards, s)
		v.ShardIndex[s.Key] = len(v.Shards) - 1
		return
	}

	lastShard := v.Shards[len(v.Shards)-1]
	if lastShard.ParsedKey.Before(s.ParsedKey) {
		v.Shards = append(v.Shards, s)
		v.ShardIndex[s.Key] = len(v.Shards) - 1
		return
	}

	v.Shards = append(v.Shards, s)
	sort.Slice(v.Shards, func(i, j int) bool {
		return v.Shards[i].ParsedKey.Before(v.Shards[j].ParsedKey)
	})

	// Rebuild the index map from scratch
	for i, shard := range v.Shards {
		v.ShardIndex[shard.Key] = i
	}
}
