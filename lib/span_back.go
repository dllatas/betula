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
	var err error
	if duration == 1 {
		from, err = truncateToUnit(ref, unit)
		if err != nil {
			return nil, err
		}
	} else {
		from, err = truncateBack(ref, unit, duration)
		if err != nil {
			return nil, err
		}
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

func truncateBack(ref time.Time, unit TimeUnit, count int) (time.Time, error) {
	switch unit {
	case UnitYear:
		return ref.AddDate(-count, 0, 0), nil
	case UnitMonth:
		return ref.AddDate(0, -count, 0), nil
	case UnitDay:
		return ref.AddDate(0, 0, -count), nil
	case UnitHour:
		return ref.Add(-time.Duration(count) * time.Hour), nil
	case UnitMinute:
		return ref.Add(-time.Duration(count) * time.Minute), nil
	case UnitSecond:
		return ref.Add(-time.Duration(count) * time.Second), nil
	case UnitMilli:
		return ref.Add(-time.Duration(count) * time.Millisecond), nil
	case UnitMicro:
		return ref.Add(-time.Duration(count) * time.Microsecond), nil
	case UnitNano:
		return ref.Add(-time.Duration(count) * time.Nanosecond), nil
	default:
		return time.Time{}, fmt.Errorf("truncate back: unit %s not supported", unit)
	}
}

func truncateToUnit(t time.Time, unit TimeUnit) (time.Time, error) {
	switch unit {
	case UnitYear:
		return time.Date(t.Year(), 1, 1, 0, 0, 0, 0, time.UTC), nil
	case UnitMonth:
		return time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, time.UTC), nil
	case UnitDay:
		return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC), nil
	case UnitHour:
		return time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), 0, 0, 0, time.UTC), nil
	case UnitMinute:
		return time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), 0, 0, time.UTC), nil
	case UnitSecond:
		return time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second(), 0, time.UTC), nil
	case UnitMilli:
		// strip micro and nano
		ns := t.Nanosecond()
		return t.Add(time.Duration(-ns % int(time.Millisecond))), nil
	case UnitMicro:
		// strip nano
		ns := t.Nanosecond()
		return t.Add(time.Duration(-ns % int(time.Microsecond))), nil
	case UnitNano:
		return t, nil // already most granular
	default:
		return time.Time{}, fmt.Errorf("truncate to unit: unit %s not supported", unit)
	}
}
