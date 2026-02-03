package lib

import (
	"testing"
	"time"
)

func TestNewShard(t *testing.T) {
	key := time.Date(2024, time.January, 1, 0, 0, 0, 0, time.UTC)

	shard, err := newShard(key)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if shard == nil {
		t.Fatal("expected non-nil shard")
	}

	if !shard.Key.Equal(key) {
		t.Fatalf("expected key %v, got %v", key, shard.Key)
	}

	if shard.Count != 0 {
		t.Fatalf("expected count 0, got %d", shard.Count)
	}

	if shard.Roots == nil || len(shard.Roots) != 0 {
		t.Fatalf("expected non-nil empty roots map, got %#v", shard.Roots)
	}
}

func TestCompareShard(t *testing.T) {
	baseTime := time.Date(2024, time.March, 10, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name string
		l    *Shard
		r    *Shard
		want bool
	}{
		{
			name: "equal shards without roots",
			l: &Shard{
				Key:   baseTime,
				Roots: map[string]*Node{},
				Count: 0,
			},
			r: &Shard{
				Key:   baseTime,
				Roots: map[string]*Node{},
				Count: 0,
			},
			want: true,
		},
		{
			name: "equal shards with roots and children",
			l: &Shard{
				Key: baseTime,
				Roots: map[string]*Node{
					"user": {
						Key:   "user",
						Value: "alice",
						Count: 2,
						Nodes: map[string]*Node{
							"country": {Key: "country", Value: "us", Count: 2},
						},
					},
				},
				Count: 2,
			},
			r: &Shard{
				Key: baseTime,
				Roots: map[string]*Node{
					"user": {
						Key:   "user",
						Value: "alice",
						Count: 2,
						Nodes: map[string]*Node{
							"country": {Key: "country", Value: "us", Count: 2},
						},
					},
				},
				Count: 2,
			},
			want: true,
		},
		{
			name: "key mismatch",
			l: &Shard{
				Key:   baseTime,
				Roots: map[string]*Node{},
				Count: 0,
			},
			r: &Shard{
				Key:   baseTime.Add(time.Hour),
				Roots: map[string]*Node{},
				Count: 0,
			},
			want: false,
		},
		{
			name: "count mismatch",
			l: &Shard{
				Key:   baseTime,
				Roots: map[string]*Node{},
				Count: 1,
			},
			r: &Shard{
				Key:   baseTime,
				Roots: map[string]*Node{},
				Count: 2,
			},
			want: false,
		},
		{
			name: "root map length mismatch",
			l: &Shard{
				Key: baseTime,
				Roots: map[string]*Node{
					"user": {Key: "user", Value: "alice", Count: 1},
				},
				Count: 1,
			},
			r: &Shard{
				Key: baseTime,
				Roots: map[string]*Node{
					"user": {Key: "user", Value: "alice", Count: 1},
					"team": {Key: "team", Value: "blue", Count: 1},
				},
				Count: 1,
			},
			want: false,
		},
		{
			name: "missing root node in expected tree",
			l: &Shard{
				Key: baseTime,
				Roots: map[string]*Node{
					"user": {Key: "user", Value: "alice", Count: 1},
				},
				Count: 1,
			},
			r: &Shard{
				Key: baseTime,
				Roots: map[string]*Node{
					"team": {Key: "team", Value: "blue", Count: 1},
				},
				Count: 1,
			},
			want: false,
		},
		{
			name: "node mismatch in root tree",
			l: &Shard{
				Key: baseTime,
				Roots: map[string]*Node{
					"user": {
						Key:   "user",
						Value: "alice",
						Count: 1,
						Nodes: map[string]*Node{
							"country": {Key: "country", Value: "us", Count: 1},
						},
					},
				},
				Count: 1,
			},
			r: &Shard{
				Key: baseTime,
				Roots: map[string]*Node{
					"user": {
						Key:   "user",
						Value: "alice",
						Count: 1,
						Nodes: map[string]*Node{
							"country": {Key: "country", Value: "us", Count: 2},
						},
					},
				},
				Count: 1,
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := compareShard(tt.l, tt.r)
			if got != tt.want {
				t.Fatalf("compareShard() = %v, want %v", got, tt.want)
			}
		})
	}
}
