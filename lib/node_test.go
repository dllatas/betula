package lib

import "testing"

func TestNewNode(t *testing.T) {
	n := newNode("user", "alice")

	if n.Key != "user" {
		t.Fatalf("expected key %q, got %q", "user", n.Key)
	}

	if n.Value != "alice" {
		t.Fatalf("expected value %q, got %q", "alice", n.Value)
	}

	if n.Count != 0 {
		t.Fatalf("expected count 0, got %d", n.Count)
	}

	if n.Nodes == nil || len(n.Nodes) != 0 {
		t.Fatalf("expected non-nil empty child map, got %#v", n.Nodes)
	}
}

func TestCompareNode(t *testing.T) {
	tests := []struct {
		name string
		l    *Node
		r    *Node
		want bool
	}{
		{
			name: "equal nodes without children",
			l:    &Node{Key: "user", Value: "alice", Count: 1},
			r:    &Node{Key: "user", Value: "alice", Count: 1},
			want: true,
		},
		{
			name: "equal nodes with nested children",
			l: &Node{
				Key:   "user",
				Value: "alice",
				Count: 2,
				Nodes: map[string]*Node{
					"country": {Key: "country", Value: "us", Count: 2},
				},
			},
			r: &Node{
				Key:   "user",
				Value: "alice",
				Count: 2,
				Nodes: map[string]*Node{
					"country": {Key: "country", Value: "us", Count: 2},
				},
			},
			want: true,
		},
		{
			name: "key mismatch",
			l:    &Node{Key: "user", Value: "alice", Count: 1},
			r:    &Node{Key: "account", Value: "alice", Count: 1},
			want: false,
		},
		{
			name: "value mismatch",
			l:    &Node{Key: "user", Value: "alice", Count: 1},
			r:    &Node{Key: "user", Value: "bob", Count: 1},
			want: false,
		},
		{
			name: "count mismatch",
			l:    &Node{Key: "user", Value: "alice", Count: 1},
			r:    &Node{Key: "user", Value: "alice", Count: 2},
			want: false,
		},
		{
			name: "child map length mismatch",
			l: &Node{
				Key:   "user",
				Value: "alice",
				Count: 1,
				Nodes: map[string]*Node{
					"country": {Key: "country", Value: "us", Count: 1},
				},
			},
			r: &Node{
				Key:   "user",
				Value: "alice",
				Count: 1,
				Nodes: map[string]*Node{
					"country": {Key: "country", Value: "us", Count: 1},
					"city":    {Key: "city", Value: "nyc", Count: 1},
				},
			},
			want: false,
		},
		{
			name: "missing child node in expected tree",
			l: &Node{
				Key:   "user",
				Value: "alice",
				Count: 1,
				Nodes: map[string]*Node{
					"country": {Key: "country", Value: "us", Count: 1},
				},
			},
			r: &Node{
				Key:   "user",
				Value: "alice",
				Count: 1,
				Nodes: map[string]*Node{
					"city": {Key: "city", Value: "nyc", Count: 1},
				},
			},
			want: false,
		},
		{
			name: "mismatch in child node tree",
			l: &Node{
				Key:   "user",
				Value: "alice",
				Count: 1,
				Nodes: map[string]*Node{
					"country": {Key: "country", Value: "us", Count: 1},
				},
			},
			r: &Node{
				Key:   "user",
				Value: "alice",
				Count: 1,
				Nodes: map[string]*Node{
					"country": {Key: "country", Value: "us", Count: 2},
				},
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := compareNode(tt.l, tt.r)
			if got != tt.want {
				t.Fatalf("compareNode() = %v, want %v", got, tt.want)
			}
		})
	}
}
