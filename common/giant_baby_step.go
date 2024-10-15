package common

import (
	"fmt"
	"math"
)

func GiantBabyStep(a, p, y int64) (int64, error) {
	m := int64(math.Sqrt(float64(p)) + 1)
	k := m
	mp := make(map[int64]int64)
	num := y % p
	mp[y] = 0
	for i := int64(1); i < m; i++ {
		num = (num * a) % p
		mp[num] = i
	}
	num = ModularExponentiation(a, m, p)
	if mm, ok := mp[num]; ok {
		return m - mm, nil
	}
	step := num
	for i := int64(2); i <= k; i++ {
		num = (num * step) % p
		if mm, ok := mp[num]; ok {
			return i*m - mm, nil
		}
	}
	return -1, fmt.Errorf("not found")
}
