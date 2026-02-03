# Betula

A high-performance, in-memory time-partitioned tree engine for aggregating and querying hierarchical event data across user-defined views.

## Installation

```bash
go get github.com/dllatas/betula
```

## Library Usage

Betula stores events in **views**. A view defines:

- `KeyOrder`: the ordered label keys that form the tree path (e.g. `country`, `user`, `event`).
- `Granularity`: the time partitioning unit (e.g. `minute`, `hour`, `day`), which determines shard boundaries.

### Quickstart

```go
package main

import (
	"fmt"
	"time"

	"github.com/dllatas/betula/lib"
)

func main() {
	// 1) Define a view.
	def := lib.NewViewDefinition("sessions", []string{"country", "user"}, lib.UnitMinute)
	mapper := lib.NewViewMapper(def)
	view := lib.NewViewInstance(mapper)

	// (Optional) Keep multiple views in a store.
	store := lib.NewStore()
	_ = store.Register(view)

	// 2) Append events.
	_ = view.Append(lib.Event{
		Timestamp: time.Date(2026, 2, 3, 21, 0, 0, 0, time.UTC),
		Labels: map[string]string{
			"country": "fr",
			"user":    "u1",
		},
	})
	_ = view.Append(lib.Event{
		Timestamp: time.Date(2026, 2, 3, 21, 5, 0, 0, time.UTC),
		Labels: map[string]string{
			"country": "fr",
			"user":    "u2",
		},
	})

	// 3) Query a time window (SpanBack is built on Range).
	ref := time.Date(2026, 2, 3, 21, 5, 0, 0, time.UTC)
	shards, _ := view.SpanBack(ref, lib.UnitMinute, 10, false)

	// 4) Filter and group results.
	filtered, _ := view.FilterWithShards(map[string][]string{
		"country": []string{"fr"},
	}, shards)

	grouped, _ := view.GroupByWithShards([]string{"country"}, filtered)
	fmt.Println(grouped) // map["fr"]=2
}
```

### Notes

- Grouping order must follow the view key order (you can also include `lib.ShardLabel` first to group by time shard).
- `Granularity` controls sharding; queries can use coarser (or equal) units, e.g. an hourly view can be queried by day.

## Test Coverage

- Run coverage across the repo: `go test ./... -coverprofile=coverage.out`
- Inspect the HTML report locally: `go tool cover -html=coverage.out`
- Coverage result files (`coverage.out`, `cover.out`, etc.) are build artifacts; keep them out of version control.
