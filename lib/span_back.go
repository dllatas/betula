package lib

import (
	"fmt"
	"log/slog"
	"time"
)

// SpanBack returns all shards covering a duration of `count` units
// ending at `ref`, inclusive. The number of returned shards depends on
// the view's granularity. The query unit must be coarser or equal to the view’s granularity.
func (v *ViewInstance) SpanBack(ref time.Time, unit TimeUnit, duration int, verbose bool) ([]*Shard, error) {
	if !IsUnitCompatible(v.Mapper.d.Granularity, unit) {
		return nil, fmt.Errorf("sub-shared resolution not supported. Def %s Query %s", v.Mapper.d.Granularity, unit)
	}

	if duration == 0 {
		return nil, fmt.Errorf("duration must be higher than zero")
	}

	var from time.Time
	if duration == 1 {
		from = truncateToUnit(ref, unit)
	} else {
		from = TruncateBack(ref, unit, duration)
	}

	if verbose {
		slog.Info("TruncateBack", "from", from, "ref", ref, "unit", unit, "duration", duration)
	}

	shards, err := v.Range(from, ref, verbose)
	if err != nil {
		return nil, err
	}

	return shards, nil
}

func TruncateBack(ref time.Time, unit TimeUnit, count int) time.Time {
	switch unit {
	case UnitYear:
		return ref.AddDate(-count, 0, 0)
	case UnitMonth:
		return ref.AddDate(0, -count, 0)
	case UnitDay:
		return ref.AddDate(0, 0, -count)
	case UnitHour:
		return ref.Add(-time.Duration(count) * time.Hour)
	case UnitMinute:
		return ref.Add(-time.Duration(count) * time.Minute)
	case UnitSecond:
		return ref.Add(-time.Duration(count) * time.Second)
	case UnitMilli:
		return ref.Add(-time.Duration(count) * time.Millisecond)
	case UnitMicro:
		return ref.Add(-time.Duration(count) * time.Microsecond)
	case UnitNano:
		return ref.Add(-time.Duration(count) * time.Nanosecond)
	default:
		return ref // fallback: no change
	}
}

func truncateToUnit(t time.Time, unit TimeUnit) time.Time {
	switch unit {
	case UnitYear:
		return time.Date(t.Year(), 1, 1, 0, 0, 0, 0, t.Location())
	case UnitMonth:
		return time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, t.Location())
	case UnitDay:
		return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
	case UnitHour:
		return time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), 0, 0, 0, t.Location())
	case UnitMinute:
		return time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), 0, 0, t.Location())
	case UnitSecond:
		return time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second(), 0, t.Location())
	case UnitMilli:
		// strip micro and nano
		ns := t.Nanosecond()
		return t.Add(time.Duration(-ns % int(time.Millisecond)))
	case UnitMicro:
		// strip nano
		ns := t.Nanosecond()
		return t.Add(time.Duration(-ns % int(time.Microsecond)))
	case UnitNano:
		return t // already most granular
	default:
		panic("truncateToUnit: unsupported unit " + string(unit))
	}
}
