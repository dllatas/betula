package lib

import (
	"log"
	"log/slog"
	"sync"
	"time"
)

type Event struct {
	Timestamp time.Time
	Labels    map[string]string
}

type ViewInstance struct {
	Mapper     *ViewMapper
	Mu         sync.RWMutex
	Shards     []*Shard       // sorted by partition key
	ShardIndex map[string]int // partition key → index in shards
}

func NewViewInstance(m *ViewMapper) *ViewInstance {
	return &ViewInstance{
		Mapper:     m,
		Shards:     []*Shard{},
		ShardIndex: map[string]int{},
	}
}

func (v *ViewInstance) Print(title string) {
	log.Printf("ViewInstance: %s", title)
	space := "  "

	for _, shard := range v.Shards {
		log.Printf("%sShard %s (count: %d)", space, shard.Key, shard.Count)
		printNodeMap(shard.Roots, space+"  ")
	}
}

func printNodeMap(nodes map[string]*Node, space string) {
	if len(nodes) == 0 {
		return
	}

	for _, node := range nodes {
		log.Printf("%sNode key=%s value=%s count=%d", space, node.Key, node.Value, node.Count)
		printNodeMap(node.Nodes, space+"  ")
	}
}

func compareViewInstances(l, r *ViewInstance) bool {
	if len(l.Shards) != len(r.Shards) {
		slog.Warn("instance shards have different length", "l", len(l.Shards), "r", len(r.Shards))
		return false
	}

	if l.Mapper.layout != r.Mapper.layout {
		slog.Warn("instance time format are different", "l", l.Mapper.layout, "r", r.Mapper.layout)
		return false
	}

	if len(l.ShardIndex) != len(r.ShardIndex) {
		slog.Warn("instance shard index have different length", "l", len(l.ShardIndex), "r", len(r.ShardIndex))
		return false
	}

	for key := range l.ShardIndex {
		if _, ok := r.ShardIndex[key]; !ok {
			slog.Warn("key missing in right index", "key", key)
			return false
		}
	}

	for key := range r.ShardIndex {
		if _, ok := l.ShardIndex[key]; !ok {
			slog.Warn("key missing in left index", "key", key)
			return false
		}
	}

	for key, value := range l.ShardIndex {
		if value != r.ShardIndex[key] {
			slog.Warn("shard index have different value", "key", key, "l", value, "r", r.ShardIndex[key])
			return false
		}
	}

	for i := 0; i < len(l.Shards); i++ {
		if !compareShard(l.Shards[i], r.Shards[i]) {
			return false
		}
	}

	return true
}
