package common

import "fmt"

// GCD - функция нахождения НОД по алгоритму Евклида
func GCD(a, b int64) int64 {
	for b != 0 {
		a, b = b, a%b
	}
	return a
}

// GCDExtended - функция нахождения НОД и коэффициентов x и y по расширенному алгоритму Евклида
func GCDExtended(a, b int64) (gcd int64, x int64, y int64, err error) {
	if a < b {
		err = fmt.Errorf("a must be greater than b or equal")
		return
	}
	U := [3]int64{a, 1, 0}
	V := [3]int64{b, 0, 1}
	for V[0] != 0 {
		q := U[0] / V[0]
		T := [3]int64{U[0] % V[0], U[1] - q*V[1], U[2] - q*V[2]}
		U, V = V, T
	}
	return U[0], U[1], U[2], nil
}

// ModInverse - находит модульную обратную a по модулю p (a^-1 mod p)
func ModInverse(a, p int64) (int64, error) {
	gcd, x, _, err := GCDExtended(a, p)
	if err != nil {
		return 0, err
	}
	if gcd != 1 {
		return 0, fmt.Errorf("no modular inverse exists for %d mod %d", a, p)
	}
	// x может быть отрицательным, приводим его к положительному значению
	return (x%p + p) % p, nil
}
