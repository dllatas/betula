package lib

import (
	"fmt"
	"time"
)

type ViewDefinition struct {
	Name       string
	KeyOrder   []string
	TimeFormat string
}

func NewViewDefinition(name string, keyOrder []string, timeFormat string) *ViewDefinition {
	return &ViewDefinition{
		Name:       name,
		KeyOrder:   keyOrder,
		TimeFormat: timeFormat,
	}
}

func (v *ViewDefinition) String() string {
	return fmt.Sprintf("View[%s] Keys: %v, TimeFormat: %s", v.Name, v.KeyOrder, v.TimeFormat)
}

type ViewMapper struct {
	d *ViewDefinition
}

func NewViewMapper(v *ViewDefinition) *ViewMapper {
	return &ViewMapper{
		d: v,
	}
}

func (v *ViewMapper) PartitionKey(e Event) (string, error) {
	if e.Timestamp.IsZero() {
		return "", fmt.Errorf("partition key: timestamp has zero value. View: %s. TimeFormat: %s", v.d.Name, v.d.TimeFormat)
	}

	r := e.Timestamp.Format(v.d.TimeFormat)

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

type Event struct {
	Timestamp time.Time
	Labels    map[string]string
}
