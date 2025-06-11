package lib

import (
	"fmt"
	"time"
)

type ViewMapper struct {
	D      *ViewDefinition
	Layout string
}

func NewViewMapper(v *ViewDefinition) *ViewMapper {
	return &ViewMapper{
		D:      v,
		Layout: layoutForUnit(v.Granularity),
	}
}

func (v *ViewMapper) ViewLayout() string {
	return v.Layout
}

func (v *ViewMapper) Granularity() TimeUnit {
	return v.D.Granularity
}

func (v *ViewMapper) PartitionKey(e Event) (time.Time, error) {
	return v.alignToPartition(e.Timestamp)
}

func (v *ViewMapper) alignToPartition(t time.Time) (time.Time, error) {
	if t.IsZero() {
		return time.Time{}, fmt.Errorf("align: zero time passed")
	}

	label := t.Format(v.Layout)
	aligned, err := time.ParseInLocation(v.Layout, label, time.UTC)
	if err != nil {
		return time.Time{}, fmt.Errorf(
			"align: failed to parse formatted time. Value: %s, Layout: %s, Err: %w",
			label, v.Layout, err,
		)
	}

	return aligned, nil
}

func (v *ViewMapper) TreePath(e Event) ([]string, error) {
	path := make([]string, 0, len(v.D.KeyOrder))

	for _, label := range v.D.KeyOrder {
		val, ok := e.Labels[label]

		if val == "" || !ok {
			return nil, fmt.Errorf("tree path: label has no value in event. label: %s. event: %+v", label, e)
		}

		path = append(path, val)
	}

	return path, nil
}
