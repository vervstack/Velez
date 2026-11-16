package velez_api_impl

import (
	"sort"
	"strconv"
)

func portsToRanges(ports []int) []string {
	if len(ports) == 0 {
		return []string{}
	}

	sortedPorts := append([]int(nil), ports...)
	sort.Ints(sortedPorts)

	ranges := make([]string, 0, len(sortedPorts))
	rangeStart := sortedPorts[0]
	rangeEnd := sortedPorts[0]

	for _, port := range sortedPorts[1:] {
		if port == rangeEnd || port == rangeEnd+1 {
			rangeEnd = port

			continue
		}

		ranges = append(ranges, portRange(rangeStart, rangeEnd))
		rangeStart = port
		rangeEnd = port
	}

	ranges = append(ranges, portRange(rangeStart, rangeEnd))

	return ranges
}

func portRange(start int, end int) string {
	if start == end {
		return strconv.Itoa(start)
	}

	return strconv.Itoa(start) + ":" + strconv.Itoa(end)
}
