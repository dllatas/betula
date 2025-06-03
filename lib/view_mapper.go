package lib

import (
	"fmt"
	"time"
)

type ViewMapper struct {
	d      *ViewDefinition
	layout string
}

func NewViewMapper(v *ViewDefinition) *ViewMapper {
	return &ViewMapper{
		d:      v,
		layout: layoutForUnit(v.Granularity),
	}
}

func (v *ViewMapper) Layout() string {
	return v.layout
}

func (v *ViewMapper) Granularity() TimeUnit {
	return v.d.Granularity
}

func (v *ViewMapper) PartitionKey(e Event) (time.Time, error) {
	return v.alignToPartition(e.Timestamp)
}

func (v *ViewMapper) alignToPartition(t time.Time) (time.Time, error) {
	if t.IsZero() {
		return time.Time{}, fmt.Errorf("align: zero time passed")
	}

	label := t.Format(v.layout)
	aligned, err := time.ParseInLocation(v.layout, label, time.UTC)
	if err != nil {
		return time.Time{}, fmt.Errorf(
			"align: failed to parse formatted time. Value: %s, Layout: %s, Err: %w",
			label, v.layout, err,
		)
	}

	return aligned, nil
}

func (v *ViewMapper) TreePath(e Event) ([]string, error) {
	path := make([]string, 0, len(v.d.KeyOrder))

	for _, label := range v.d.KeyOrder {
		val, ok := e.Labels[label]

		if val == "" || !ok {
			return nil, fmt.Errorf("tree path: label has no value in event. label: %s. event: %+v", label, e)
		}

		path = append(path, val)
	}

	return path, nil
}
