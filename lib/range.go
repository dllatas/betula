package lib

import (
	"fmt"
	"log/slog"
	"time"
)

func (v *ViewInstance) Range(from, to time.Time, verbose bool) ([]*Shard, error) {
	if from.After(to) {
		return nil, fmt.Errorf("invalid range: from %v is after to %v", from, to)
	}

	fromAligned, err := v.Mapper.alignToPartition(from)
	if err != nil {
		return []*Shard{}, nil
	}

	fromIdx, found := v.ShardIndex[fromAligned]
	if !found {
		if verbose {
			slog.Warn("no index shard found", "from", fromAligned, "formatted from", from)
		}
		fromIdx = v.findCloserIndex(from, ">", verbose)
		if fromIdx == -1 {
			return []*Shard{}, nil
		}
	}

	toAligned, err := v.Mapper.alignToPartition(to)
	if err != nil {
		return []*Shard{}, nil
	}

	toIdx, found := v.ShardIndex[toAligned]
	if !found {
		if verbose {
			slog.Warn("no index shard found", "to", toAligned, "formatted from", to)
		}
		toIdx = v.findCloserIndex(to, "<", verbose)
		if toIdx == -1 {
			return []*Shard{}, nil
		}
	}

	if verbose {
		slog.Info("get shards for", "from", from, "fromStr", fromAligned, "fromIdx", fromIdx, "to", to, "toStr", toAligned, "toIdx", toIdx)
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

// we know that timeRef does not exist
func (v *ViewInstance) findCloserIndex(timeRef time.Time, sign string, verbose bool) int {
	if len(v.Shards) == 0 {
		return -1
	}

	head := 0
	tail := len(v.Shards) - 1
	middle := 0

	for head <= tail {
		middle = (head + tail) / 2

		root := v.Shards[middle]

		if verbose {
			slog.Info("findClosestLaterIndex", "head", head, "tail", tail, "middle", middle, "root", root)
		}

		if root.Key.Before(timeRef) {
			head = middle + 1
			continue
		}

		if root.Key.After(timeRef) {
			tail = middle - 1
			continue
		}
	}

	found := v.Shards[middle]

	if sign == "<" && found.Key.After(timeRef) {
		return -1
	}

	if sign == ">" && found.Key.Before(timeRef) {
		return -1
	}

	return middle
}
