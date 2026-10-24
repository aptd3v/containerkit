package types

import (
	"fmt"
	"strconv"
	"strings"
)

// Cpuset is a Docker CpusetCpus / CpusetMems list, such as "0-3,5".
type Cpuset string

// ParseCpuset checks a cpuset list and returns it.
func ParseCpuset(list string) (Cpuset, error) {
	if list == "" {
		return "", fmt.Errorf("cpuset is empty")
	}
	if err := validateCpuset(list); err != nil {
		return "", err
	}
	return Cpuset(list), nil
}

// String returns the cpuset list.
func (s Cpuset) String() string { return string(s) }

func validateCpuset(mems string) error {
	for r := range strings.SplitSeq(mems, ",") {
		r = strings.TrimSpace(r)
		if r == "" {
			return fmt.Errorf("empty cpuset range")
		}
		parts := strings.Split(r, "-")
		if len(parts) == 1 {
			if _, err := strconv.Atoi(parts[0]); err != nil {
				return fmt.Errorf("invalid cpuset value: %s", r)
			}
		} else if len(parts) == 2 {
			start, err1 := strconv.Atoi(parts[0])
			end, err2 := strconv.Atoi(parts[1])
			if err1 != nil || err2 != nil || start > end {
				return fmt.Errorf("invalid cpuset range: %s", r)
			}
		} else {
			return fmt.Errorf("invalid cpuset format: %s", r)
		}
	}
	return nil
}
