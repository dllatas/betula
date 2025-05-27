package lib

import (
	"fmt"
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

func (v *ViewMapper) PartitionKey(e Event) (string, error) {
	if e.Timestamp.IsZero() {
		return "", fmt.Errorf("partition key: timestamp has zero value. View: %s. TimeFormat: %s", v.d.Name, v.layout)
	}

	r := e.Timestamp.Format(v.layout)

	return r, nil
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
