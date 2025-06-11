package lib

import "log/slog"

func (v *ViewInstance) Delete(e Event) error {
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

	idx, found := v.ShardIndex[key]
	if !found {
		slog.Warn("delete: no shard to delete from", "event", e)
		return nil
	}

	shard := v.Shards[idx]

	children := &shard.Roots

	for idx, treeValue := range treeValues {
		var next *Node

		orderKey := v.Mapper.D.KeyOrder[idx]
		mapKey := shardMapKey(orderKey, treeValue)

		child, found := (*children)[mapKey]
		if found {
			if child.Count > 0 {
				child.Count--
			}
			next = child
		} else {
			break
		}

		children = &next.Nodes
	}

	if shard.Count > 0 {
		shard.Count--
	}

	return nil
}
