package lib

import "fmt"

type TimeUnit string

const (
	UnitYear   TimeUnit = "year"
	UnitMonth  TimeUnit = "month"
	UnitDay    TimeUnit = "day"
	UnitHour   TimeUnit = "hour"
	UnitMinute TimeUnit = "minute"
	UnitSecond TimeUnit = "second"
	UnitMilli  TimeUnit = "millisecond"
	UnitMicro  TimeUnit = "microsecond"
	UnitNano   TimeUnit = "nanosecond"
	ShardLabel string   = "shard"
)

var unitOrder = []TimeUnit{
	UnitYear,
	UnitMonth,
	UnitDay,
	UnitHour,
	UnitMinute,
	UnitSecond,
	UnitMilli,
	UnitMicro,
	UnitNano,
}

func indexOf(units []TimeUnit, element TimeUnit) int {
	for i, v := range units {
		if v == element {
			return i
		}
	}
	return -1
}

/*
* If view is defined with granularity nanoseconds, then it is possible to span back in time with any other time unit.
* Just be careful that using a higher unit might make the range much more larger and pull much more data.
* If a view has granularity year, then it must query only years since it cant really use another time unit to perform a query.
 */

func IsUnitCompatible(viewUnit, queryUnit TimeUnit) bool {
	vIdx := indexOf(unitOrder, viewUnit)
	qIdx := indexOf(unitOrder, queryUnit)

	return qIdx <= vIdx // query must be coarser or equal granularity
}

func layoutForUnit(u TimeUnit) string {
	switch u {
	case UnitYear:
		return "2006"
	case UnitMonth:
		return "2006-01"
	case UnitDay:
		return "2006-01-02"
	case UnitHour:
		return "2006-01-02 15"
	case UnitMinute:
		return "2006-01-02 15:04"
	case UnitSecond:
		return "2006-01-02 15:04:05"
	case UnitMilli:
		return "2006-01-02 15:04:05.000"
	case UnitMicro:
		return "2006-01-02 15:04:05.000000"
	case UnitNano:
		return "2006-01-02 15:04:05.000000000"
	default:
		return "2006-01-02" // conservative fallback
	}
}

type ViewDefinition struct {
	Name        string
	KeyOrder    []string
	Granularity TimeUnit
}

func NewViewDefinition(name string, keyOrder []string, unit TimeUnit) *ViewDefinition {
	return &ViewDefinition{
		Name:        name,
		KeyOrder:    keyOrder,
		Granularity: unit,
	}
}

func (v *ViewDefinition) String() string {
	return fmt.Sprintf("View[%s] Keys: %v, Granularity: %s", v.Name, v.KeyOrder, v.Granularity)
}
