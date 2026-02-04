package lib

import "sort"

func (v *ViewInstance) Append(e Event) error {
	v.mu.Lock()
	defer v.mu.Unlock()

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
		shard, errShard = newShard(key)
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

		orderKey := v.Mapper.D.KeyOrder[idx]
		mapKey := shardMapKey(orderKey, treeValue)

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

func shardMapKey(orderKey, treeValue string) string {
	return orderKey + ":" + treeValue
}

func (v *ViewInstance) addShard(s *Shard) {
	if len(v.Shards) == 0 {
		v.Shards = append(v.Shards, s)
		v.ShardIndex[s.Key] = len(v.Shards) - 1
		return
	}

	lastShard := v.Shards[len(v.Shards)-1]
	if lastShard.Key.Before(s.Key) {
		v.Shards = append(v.Shards, s)
		v.ShardIndex[s.Key] = len(v.Shards) - 1
		return
	}

	v.Shards = append(v.Shards, s)
	sort.Slice(v.Shards, func(i, j int) bool {
		return v.Shards[i].Key.Before(v.Shards[j].Key)
	})

	// Rebuild the index map from scratch
	for i, shard := range v.Shards {
		v.ShardIndex[shard.Key] = i
	}
}
