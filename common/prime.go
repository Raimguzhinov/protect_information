package common

import (
	"math/big"
	"math/rand"
	"runtime"
	"sync"
	"time"
)

func GenPrime(minV int64, maxV int64) int64 {
	for {
		num := Seed().Int63n(maxV-minV) + minV
		if IsPrime(num) {
			return num
		}
	}
}

// IsPrime - функция проверки числа на простоту (алгоритм Ферма)
func IsPrime(p int64) bool {
	if p <= 1 {
		return false
	} else if p <= 3 {
		return true
	} else if p%2 == 0 || p%3 == 0 {
		return false
	}
	var wg sync.WaitGroup
	wg.Add(runtime.NumCPU())
	doneCh := make(chan bool, runtime.NumCPU())
	for i := 0; i < runtime.NumCPU(); i++ {
		go func() {
			defer wg.Done()
			a := Seed().Int63n(p-2) + 2 // [2, p-2]
			if ModularExponentiation(a, p-1, p) == 1 {
				doneCh <- true
			}
		}()
	}
	wg.Wait()
	close(doneCh)
	for v := range doneCh {
		if v {
			return true
		}
	}
	return false
}

func Seed() *rand.Rand {
	return rand.New(rand.NewSource(time.Now().UnixNano()))
}

// GenCoprime - генерирует число, взаимно простое с n
func GenCoprime(n int64, minV int64, maxV int64) int64 {
	for {
		num := Seed().Int63n(maxV-minV) + minV
		if gcd, _, _ := GCDExtended(n, num); gcd == 1 {
			return num
		}
	}
}

// PrimitiveRoot - находит примитивный корень для простого числа p
func PrimitiveRoot(p int64) int64 {
	if !IsPrime(p) {
		return -1 // p должно быть простым числом
	}
	var factors []int64
	phi := p - 1 // φ(p) = p-1 для простых чисел
	// Разложение φ(p) на простые множители
	n := phi
	for i := int64(2); i*i <= n; i++ {
		if n%i == 0 {
			factors = append(factors, i)
			for n%i == 0 {
				n /= i
			}
		}
	}
	if n > 1 {
		factors = append(factors, n)
	}
	// Ищем минимальное g, такое что g^((p-1)/f) mod p != 1 для каждого простого делителя f
	for g := int64(2); g < p; g++ {
		isPrimitive := true
		for _, factor := range factors {
			// Проверяем, что g^((p-1)/factor) % p != 1
			if ModularExponentiation(g, phi/factor, p) == 1 {
				isPrimitive = false
				break
			}
		}
		if isPrimitive {
			return g
		}
	}
	return -1 // Если не найдено (теоретически не должно происходить)
}

func GenPrimeBig(minV, maxV *big.Int) *big.Int {
	for {
		num := new(big.Int).Rand(SeedBig(), new(big.Int).Sub(maxV, minV))
		num.Add(num, minV) // num = num + minV
		if IsPrimeBig(num) {
			return num
		}
	}
}

// IsPrimeBig проверяет, является ли число простым (используется алгоритм Ферма)
func IsPrimeBig(p *big.Int) bool {
	one := big.NewInt(1)
	two := big.NewInt(2)
	three := big.NewInt(3)
	if p.Cmp(one) <= 0 {
		return false
	} else if p.Cmp(three) <= 0 {
		return true
	}
	if new(big.Int).Mod(p, two).Cmp(big.NewInt(0)) == 0 {
		return false
	}
	var wg sync.WaitGroup
	wg.Add(runtime.NumCPU())
	doneCh := make(chan bool, runtime.NumCPU())
	for i := 0; i < runtime.NumCPU(); i++ {
		go func() {
			defer wg.Done()
			a := new(big.Int).Rand(SeedBig(), new(big.Int).Sub(p, two)) // [2, p-2]
			a.Add(a, two)

			if ModularExponentiationBig(a, new(big.Int).Sub(p, one), p).Cmp(one) == 0 {
				doneCh <- true
			}
		}()
	}
	wg.Wait()
	close(doneCh)
	for v := range doneCh {
		if v {
			return true
		}
	}
	return false
}

// SeedBig возвращает новый источник случайных чисел
func SeedBig() *rand.Rand {
	return rand.New(rand.NewSource(time.Now().UnixNano()))
}

// GenCoprimeBig генерирует случайное число num в диапазоне [minV, maxV] такое, что GCD(n, num) = 1
func GenCoprimeBig(n, minV, maxV *big.Int) *big.Int {
	// Проверяем, что minV < maxV
	if minV.Cmp(maxV) >= 0 {
		return nil // Или можно выбросить ошибку
	}
	// Разница maxV - minV для генерации значения в диапазоне
	rangeSize := new(big.Int).Sub(maxV, minV)
	for {
		// Генерируем случайное значение num в диапазоне [minV, maxV]
		num := new(big.Int).Rand(SeedBig(), rangeSize)
		num.Add(num, minV) // num = num + minV
		// Проверяем, что num взаимно просто с n (GCD(n, num) == 1)
		if new(big.Int).GCD(nil, nil, n, num).Cmp(big.NewInt(1)) == 0 {
			return num
		}
	}
}
