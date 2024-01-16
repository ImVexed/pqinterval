package pqinterval

import (
	"fmt"
	"strconv"
	"strings"
)

// ParseErr is returned on a failure to parse a
// postgres result into an Interval or Duration.
type ParseErr struct {
	String string
	Cause  error
}

func Parse(s string) (Interval, error) {
	chunks := strings.Split(s, " ")

	ival := Interval{}
	var negTime bool

	if len(chunks)%2 == 1 {
		// Parse the time component
		t := chunks[len(chunks)-1]
		chunks = chunks[:len(chunks)-1]

		if t[0] == '-' {
			negTime = true
			t = t[1:]
		} else if t[0] == '+' {
			t = t[1:]
		}

		parts := strings.Split(t, ":")
		if len(parts) < 2 || len(parts) > 3 {
			return ival, fmt.Errorf("invalid time format")
		}

		hrs, err := strconv.Atoi(parts[0])
		if err != nil {
			return ival, err
		}
		if negTime {
			hrs = -hrs
		}
		mins, err := strconv.Atoi(parts[1])
		if err != nil {
			return ival, err
		}
		var secs int
		if len(parts) > 2 {
			secs, err = strconv.Atoi(parts[2])
			if err != nil {
				return ival, err
			}
		}

		ival.hrs = int32(hrs)
		ival.us = uint32(mins*usPerMin + secs*usPerSec)
	}

	for len(chunks) > 0 {
		quantity, err := strconv.Atoi(chunks[0])
		if err != nil {
			return Interval{}, err
		}
		unit := chunks[1]
		chunks = chunks[2:]

		switch unit {
		case "year", "years":
			ival.yrs = uint32(quantity)
			if negTime {
				ival.yrs |= yrSignBit
			}
		case "mon", "mons":
			ival.hrs += int32(24 * daysPerMon * quantity)
		case "day", "days":
			ival.hrs += int32(24 * quantity)
		default:
			return Interval{}, fmt.Errorf("invalid unit: %s", unit)
		}
	}

	return ival, nil
}

// Error implements the error interface.
func (pe ParseErr) Error() string {
	return fmt.Sprintf("pqinterval: Error parsing %q: %s", pe.String, pe.Cause)
}
