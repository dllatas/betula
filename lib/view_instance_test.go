package lib

import (
	"bytes"
	"log"
	"testing"
	"time"
)

func TestNewViewInstance(t *testing.T) {
	d := NewViewDefinition("test-view", []string{"club", "player"}, UnitDay)
	m := NewViewMapper(d)

	v := NewViewInstance(m)

	if v.Mapper != m {
		t.Fatalf("expected mapper %v, got %v", m, v.Mapper)
	}
	if v.Shards == nil || len(v.Shards) != 0 {
		t.Fatalf("expected empty shards slice, got %#v", v.Shards)
	}
	if v.ShardIndex == nil || len(v.ShardIndex) != 0 {
		t.Fatalf("expected empty shard index map, got %#v", v.ShardIndex)
	}
}

func TestViewInstancePrint(t *testing.T) {
	var buf bytes.Buffer
	origWriter := log.Writer()
	log.SetOutput(&buf)
	defer log.SetOutput(origWriter)

	d := NewViewDefinition("test-view", []string{"club", "player"}, UnitDay)
	m := NewViewMapper(d)
	v := NewViewInstance(m)

	baseTime := time.Date(2024, time.January, 1, 0, 0, 0, 0, time.UTC)
	v.Shards = []*Shard{
		{
			Key: baseTime,
			Roots: map[string]*Node{
				"user": {
					Key:   "user",
					Value: "alice",
					Count: 1,
				},
			},
			Count: 1,
		},
	}

	v.Print("test-instance")

	if buf.Len() == 0 {
		t.Fatal("expected Print to write log output, got empty buffer")
	}
}

func TestCompareViewInstances(t *testing.T) {
	baseTime := time.Date(2024, time.January, 1, 0, 0, 0, 0, time.UTC)

	makeInstance := func(layout string, shardKeys []time.Time, shardCounts []int) *ViewInstance {
		if len(shardKeys) != len(shardCounts) {
			t.Fatalf("test setup error: shardKeys and shardCounts length mismatch")
		}

		d := &ViewDefinition{
			Name:        "test-view",
			KeyOrder:    []string{"club", "player"},
			Granularity: UnitDay,
		}
		m := &ViewMapper{
			D:      d,
			Layout: layout,
		}

		shards := make([]*Shard, len(shardKeys))
		shardIndex := make(map[time.Time]int, len(shardKeys))

		for i, k := range shardKeys {
			shards[i] = &Shard{
				Key:   k,
				Roots: map[string]*Node{},
				Count: shardCounts[i],
			}
			shardIndex[k] = i
		}

		return &ViewInstance{
			Mapper:     m,
			Shards:     shards,
			ShardIndex: shardIndex,
		}
	}

	tests := []struct {
		name string
		l    *ViewInstance
		r    *ViewInstance
		want bool
	}{
		{
			name: "equal instances with same shards and mapper layout",
			l: func() *ViewInstance {
				keys := []time.Time{baseTime, baseTime.Add(24 * time.Hour)}
				counts := []int{1, 2}
				inst := makeInstance("2006-01-02", keys, counts)

				inst.Shards[0].Roots["user"] = &Node{
					Key:   "user",
					Value: "alice",
					Count: 1,
				}
				inst.Shards[1].Roots["user"] = &Node{
					Key:   "user",
					Value: "bob",
					Count: 2,
				}
				return inst
			}(),
			r: func() *ViewInstance {
				keys := []time.Time{baseTime, baseTime.Add(24 * time.Hour)}
				counts := []int{1, 2}
				inst := makeInstance("2006-01-02", keys, counts)

				inst.Shards[0].Roots["user"] = &Node{
					Key:   "user",
					Value: "alice",
					Count: 1,
				}
				inst.Shards[1].Roots["user"] = &Node{
					Key:   "user",
					Value: "bob",
					Count: 2,
				}
				return inst
			}(),
			want: true,
		},
		{
			name: "different number of shards",
			l:    makeInstance("2006-01-02", []time.Time{baseTime}, []int{1}),
			r:    makeInstance("2006-01-02", []time.Time{baseTime, baseTime.Add(24 * time.Hour)}, []int{1, 2}),
			want: false,
		},
		{
			name: "different mapper layouts",
			l:    makeInstance("2006-01-02", []time.Time{baseTime}, []int{1}),
			r:    makeInstance("2006-01-02 15:04", []time.Time{baseTime}, []int{1}),
			want: false,
		},
		{
			name: "different shard index lengths",
			l:    makeInstance("2006-01-02", []time.Time{baseTime}, []int{1}),
			r: func() *ViewInstance {
				inst := makeInstance("2006-01-02", []time.Time{baseTime}, []int{1})
				inst.ShardIndex[baseTime.Add(24*time.Hour)] = 1
				return inst
			}(),
			want: false,
		},
		{
			name: "missing key in right shard index",
			l: func() *ViewInstance {
				inst := makeInstance("2006-01-02", []time.Time{baseTime}, []int{1})
				return inst
			}(),
			r: func() *ViewInstance {
				inst := makeInstance("2006-01-02", []time.Time{baseTime}, []int{1})
				delete(inst.ShardIndex, baseTime)
				return inst
			}(),
			want: false,
		},
		{
			name: "missing key in left shard index",
			l: func() *ViewInstance {
				inst := makeInstance("2006-01-02", []time.Time{baseTime}, []int{1})
				delete(inst.ShardIndex, baseTime)
				return inst
			}(),
			r: func() *ViewInstance {
				inst := makeInstance("2006-01-02", []time.Time{baseTime}, []int{1})
				return inst
			}(),
			want: false,
		},
		{
			name: "different shard index values for same key",
			l: func() *ViewInstance {
				inst := makeInstance("2006-01-02", []time.Time{baseTime, baseTime.Add(24 * time.Hour)}, []int{1, 2})
				return inst
			}(),
			r: func() *ViewInstance {
				inst := makeInstance("2006-01-02", []time.Time{baseTime, baseTime.Add(24 * time.Hour)}, []int{1, 2})
				inst.ShardIndex[baseTime] = 1
				inst.ShardIndex[baseTime.Add(24*time.Hour)] = 0
				return inst
			}(),
			want: false,
		},
		{
			name: "shard comparison fails",
			l: func() *ViewInstance {
				keys := []time.Time{baseTime}
				counts := []int{1}
				inst := makeInstance("2006-01-02", keys, counts)
				inst.Shards[0].Roots["user"] = &Node{
					Key:   "user",
					Value: "alice",
					Count: 1,
				}
				return inst
			}(),
			r: func() *ViewInstance {
				keys := []time.Time{baseTime}
				counts := []int{1}
				inst := makeInstance("2006-01-02", keys, counts)
				inst.Shards[0].Roots["user"] = &Node{
					Key:   "user",
					Value: "alice",
					Count: 2,
				}
				return inst
			}(),
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := compareViewInstances(tt.l, tt.r)
			if got != tt.want {
				t.Fatalf("compareViewInstances() = %v, want %v", got, tt.want)
			}
		})
	}
}
