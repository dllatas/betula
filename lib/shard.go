package lib

import (
	"log/slog"
	"time"
)

type Shard struct {
	Key       string           // e.g. "2024-05-20"
	ParsedKey time.Time        // for internal comparisons if needed
	Roots     map[string]*Node // The actual tree root for this partition
	Count     int              // Optional: total count at shard level
}

func newShard(key string, ts time.Time, fmt string) (*Shard, error) {
	t, err := time.Parse(fmt, ts.Format(fmt))
	if err != nil {
		return nil, err
	}

	r := make(map[string]*Node)

	return &Shard{
		Key:       key,
		ParsedKey: t,
		Roots:     r,
		Count:     0,
	}, nil
}

func compareShard(l, r *Shard) bool {
	if l.Key != r.Key {
		slog.Warn("shard key mismatch", "l", l.Key, "r", r.Key)
		return false
	}

	if !l.ParsedKey.Equal(r.ParsedKey) {
		slog.Warn("shard parsed key mismatch", "l", l.ParsedKey, "r", r.ParsedKey)
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
