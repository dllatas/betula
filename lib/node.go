package lib

import "log/slog"

type Node struct {
	Key   string           // e.g. "user"
	Value string           // e.g. "alice"
	Count int              // amount of ocurrences
	Nodes map[string]*Node // list of dependant nodes
}

func newNode(key string, value string) *Node {
	return &Node{
		Key:   key,
		Value: value,
		Count: 0,
		Nodes: map[string]*Node{},
	}
}

func compareNode(l, r *Node) bool {
	if l.Key != r.Key {
		slog.Warn("node key mismatch", "l", l.Key, "r", r.Key)
		return false
	}

	if l.Value != r.Value {
		slog.Warn("node value mismatch", "key", l.Key, "l", l.Value, "r", r.Value)
		return false
	}

	if l.Count != r.Count {
		slog.Warn("node count mismatch", "key", l.Key+":"+l.Value, "l", l.Count, "r", r.Count)
		return false
	}

	if len(l.Nodes) != len(r.Nodes) {
		slog.Warn("child node map length mismatch", "key", l.Key+":"+l.Value, "l", len(l.Nodes), "r", len(r.Nodes))
		return false
	}

	for k, lChild := range l.Nodes {
		rChild, ok := r.Nodes[k]
		if !ok {
			slog.Warn("missing child node in expected tree", "key", k)
			return false
		}

		if !compareNode(lChild, rChild) {
			slog.Warn("mismatch in child node tree", "key", k)
			return false
		}
	}

	return true
}
