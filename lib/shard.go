package lib

import (
	"log/slog"
	"time"
)

type Shard struct {
	Key   time.Time        // for internal comparisons if needed
	Roots map[string]*Node // The actual tree root for this partition
	Count int              // Optional: total count at shard level
}

func newShard(key time.Time) (*Shard, error) {
	return &Shard{
		Key:   key,
		Roots: map[string]*Node{},
		Count: 0,
	}, nil
}

func compareShard(l, r *Shard) bool {
	if !l.Key.Equal(r.Key) {
		slog.Warn("shard parsed key mismatch", "l", l.Key, "r", r.Key)
		return false
	}

	if l.Count != r.Count {
		slog.Warn("shard count mismatch", "shard", l.Key, "l", l.Count, "r", r.Count)
		return false
	}

	if len(l.Roots) != len(r.Roots) {
		slog.Warn("root map length mismatch", "shard", l.Key, "l", len(l.Roots), "r", len(r.Roots))
		return false
	}

	for k, lNode := range l.Roots {
		rNode, ok := r.Roots[k]
		if !ok {
			slog.Warn("missing root node in expected tree", "key", k)
			return false
		}
		if !compareNode(lNode, rNode) {
			slog.Warn("node mismatch in root", "key", k)
			return false
		}
	}

	return true
}
