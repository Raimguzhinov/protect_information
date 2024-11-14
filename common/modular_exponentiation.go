package common

import (
	"math"
	"math/big"
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

// ModularExponentiationBig выполняет возведения a^x mod p для больших чисел
func ModularExponentiationBig(a, x, p *big.Int) *big.Int {
	// Инициализируем результат как 1
	result := big.NewInt(1)
	base := new(big.Int).Mod(a, p) // Приводим `a` к модулю `p`
	exp := new(big.Int).Set(x)     // Копия `x` для модификации в цикле
	// Используем цикл для быстрого возведения в степень
	zero := big.NewInt(0)
	one := big.NewInt(1)
	for exp.Cmp(zero) > 0 {
		// Если текущий бит экспоненты равен 1, умножаем результат на базу
		if new(big.Int).And(exp, one).Cmp(one) == 0 {
			result.Mul(result, base)
			result.Mod(result, p)
		}
		// Возводим базу в квадрат и берем по модулю `p`
		base.Mul(base, base)
		base.Mod(base, p)
		// Сдвигаем экспоненту вправо на 1 бит (деление на 2)
		exp.Rsh(exp, 1)
	}
	return result
}
