package lib

import (
	"fmt"
	"log/slog"
	"time"
)

func (v *ViewInstance) Range(from, to time.Time) ([]*Shard, error) {
	if from.After(to) {
		return nil, fmt.Errorf("invalid range: from %v is after to %v", from, to)
	}

	fromStr := from.Format(v.Def.TimeFormat)
	toStr := to.Format(v.Def.TimeFormat)

	fromIdx, found := v.ShardIndex[fromStr]
	if !found {
		slog.Warn("no index shard found", "from", fromStr, "formatted from", from)
		return []*Shard{}, nil
	}

	toIdx, found := v.ShardIndex[toStr]
	if !found {
		slog.Warn("no index shard found", "to", toStr, "formatted from", to)
		return []*Shard{}, nil
	}

	shards, err := v.getShardsByIndex(fromIdx, toIdx)
	if err != nil {
		slog.Warn("getShardsByIndex failed", "err", err.Error())
		return []*Shard{}, nil
	}

	return shards, nil
}

// isNodeRangeValid checks that head and tail input makes sense
// when getting a list of nodes.
// Basically, both params need to be larger than zero since
// they will be used as array indexes.
// Also, tail should not be larger than the current nodes
// array size.
// Finally head must be lower than tail
func validateIndexRange(size, head, tail int) error {
	if head < 0 || tail < 0 {
		return fmt.Errorf("isIndexPairValid failed: head (%d) and tail (%d) are lower than zero", head, tail)
	}

	if size < tail {
		return fmt.Errorf("isIndexPairValid failed: shard size (%d) is lower than tail (%d)", size, tail)
	}

	if head > tail {
		return fmt.Errorf("isIndexPairValid failed: head (%d) is higher than tail (%d)", head, tail)
	}

	return nil
}

func (v *ViewInstance) getShardsByIndex(head, tail int) ([]*Shard, error) {
	tail = tail + 1
	if err := validateIndexRange(len(v.Shards), head, tail); err != nil {
		return nil, err
	}

	return v.Shards[head:tail], nil
}
