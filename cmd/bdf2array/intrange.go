package main

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
)

// intrange contains one or more integers
type intrange map[int]struct{}

// Set adds values to the existing codepoints list.
// Valid value formats are:
// decimal numbers: 1
// hexadecimal numbers: 0x02
// comma-separated lists: 0x01,2,0x04
// comma-separated lists with spaces (must be quoted on the command line): 0x01, 0x02, 0x04
// ranges: 0x00-0xFF
// combinations of the above
func (ir *intrange) Set(value string) error {
	if ir == nil {
		return fmt.Errorf("nil receiver")
	}
	if *ir == nil {
		*ir = make(intrange)
	}
	commasegments := strings.Split(value, ",")

	for _, commasegment := range commasegments {
		dashsegments := strings.Split(commasegment, "-")
		dashsegments[0] = strings.TrimSpace(dashsegments[0])
		first, err := strconv.ParseUint(dashsegments[0], 0, 8)
		if err != nil {
			return err
		}

		switch len(dashsegments) {
		case 1:
			(*ir)[int(first)] = struct{}{}

		case 2: // nolint:gomnd
			dashsegments[1] = strings.TrimSpace(dashsegments[1])
			second, err := strconv.ParseUint(dashsegments[1], 0, 8)
			if err != nil {
				return err
			}
			if first >= second {
				return fmt.Errorf("invalid range specified ('%s')", commasegment)
			}
			for i := int(first); i <= int(second); i++ {
				(*ir)[i] = struct{}{}
			}

		default:
			return fmt.Errorf("'%s' contains too many '-'", commasegment)
		}
	}

	return nil
}

func (ir *intrange) GetSlice() []int {
	if ir == nil || *ir == nil {
		return []int{}
	}

	numvalues := len(*ir)
	values := make([]int, numvalues)
	i := 0
	for val := range *ir {
		values[i] = val
		i++
	}
	sort.Ints(values)
	return values
}

func (ir *intrange) String() string {
	if ir == nil || *ir == nil {
		return "nil"
	}

	values := ir.GetSlice()

	if len(values) == 0 {
		return ""
	}

	parts := make([]string, 0)
	// always print the first value:
	low := values[0]
	high := low
	isRange := false

	// start at the 2nd value:
	numvalues := len(values)
	for i := 1; i < numvalues; i++ {
		if values[i] == high+1 {
			high = values[i] // start a range or extend the existing range
			isRange = true
			continue
		}
		switch {
		case !isRange:
			// the previous value was not a range, output it as single value:
			parts = append(parts, fmt.Sprintf("0x%02X", low))
		case high-low == 1:
			// low-high was a range of only 2 values, output them individually:
			parts = append(parts, fmt.Sprintf("0x%02X", low))
			parts = append(parts, fmt.Sprintf("0x%02X", high))
		default:
			// output the previous range:
			parts = append(parts, fmt.Sprintf("0x%02X-0x%02X", low, high))
		}
		isRange = false
		// and start a new one:
		low = values[i]
		high = low
	}

	switch {
	case !isRange:
		// output the last individual value:
		parts = append(parts, fmt.Sprintf("0x%02X", low))
	case high-low == 1:
		// low-high was a range of only 2 values, output them individually:
		parts = append(parts, fmt.Sprintf("0x%02X", low))
		parts = append(parts, fmt.Sprintf("0x%02X", high))
	default:
		parts = append(parts, fmt.Sprintf("0x%02X-0x%02X", low, high))
	}

	return strings.Join(parts, ",")
}
