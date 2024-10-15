package common

import (
	"math"
	"strconv"
)

// ModularExponentiation - функция быстрого возведения a^x mod p
func ModularExponentiation(a, x, p int64) int64 {
	y := int64(1)
	s := a
	t := int(math.Floor(math.Log2(float64(x))))
	binaryMask := strconv.FormatInt(x, 2)
	for i := t; i >= 0; i-- {
		if binaryMask[i] == '1' {
			y = (y * s) % p
		}
		s = (s * s) % p
	}
	return y
}
