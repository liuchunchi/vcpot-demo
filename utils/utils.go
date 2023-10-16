package utils

import "time"

func IsUnique(nums []uint32) bool {
	freq := make(map[uint32]int)

	for _, num := range nums {
		freq[num]++
	}

	for _, count := range freq {
		if count > 1 {
			return false
		}
	}

	return true
}

func DurationDivideBy(duration time.Duration, divisor int) time.Duration {
	return time.Duration(duration.Nanoseconds() / int64(divisor))
}
