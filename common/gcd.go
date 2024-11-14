package common

import (
	"fmt"
	"math/big"
)

// GCD - функция нахождения НОД по алгоритму Евклида
func GCD(a, b int64) int64 {
	for b != 0 {
		a, b = b, a%b
	}
	return a
}

// GCDExtended - функция нахождения НОД и коэффициентов x и y по расширенному алгоритму Евклида
func GCDExtended(a, b int64) (gcd int64, x int64, y int64) {
	if a < b {
		a, b = b, a
	}
	U := [3]int64{a, 1, 0}
	V := [3]int64{b, 0, 1}
	for V[0] != 0 {
		q := U[0] / V[0]
		T := [3]int64{U[0] % V[0], U[1] - q*V[1], U[2] - q*V[2]}
		U, V = V, T
	}
	return U[0], U[1], U[2]
}

// ModInverse - находит модульную обратную a по модулю p (a^-1 mod p)
func ModInverse(a, p int64) (int64, error) {
	if a > p {
		a %= p
	}
	gcd, _, y := GCDExtended(a, p)
	if gcd != 1 {
		return 0, fmt.Errorf("no modular inverse exists for %d mod %d", a, p)
	}
	if y < 0 {
		y += p
	}
	return y, nil
}

// GCD - вычисляет наибольший общий делитель для big.Int
func GCDBig(a, b *big.Int) *big.Int {
	gcd := new(big.Int)
	return gcd.GCD(nil, nil, a, b) // Использует встроенный метод GCD
}

// GCDExtendedBig вычисляет GCD и коэффициенты x и y для уравнения a*x + b*y = GCD(a, b)
func GCDExtendedBig(a, b *big.Int) (*big.Int, *big.Int, *big.Int) {
	// Инициализируем переменные для алгоритма
	x := big.NewInt(1)
	y := big.NewInt(0)
	x1 := big.NewInt(0)
	y1 := big.NewInt(1)
	aCopy := new(big.Int).Set(a)
	bCopy := new(big.Int).Set(b)
	// Выполняем расширенный алгоритм Евклида
	for bCopy.Cmp(big.NewInt(0)) != 0 {
		q := new(big.Int).Div(aCopy, bCopy)
		r := new(big.Int).Mod(aCopy, bCopy)
		aCopy.Set(bCopy)
		bCopy.Set(r)
		// Обновляем x и y
		x, x1 = x1, new(big.Int).Sub(x, new(big.Int).Mul(q, x1))
		y, y1 = y1, new(big.Int).Sub(y, new(big.Int).Mul(q, y1))
	}
	return aCopy, x, y
}

// ModInverseBig находит обратное значение a по модулю p, используя расширенный алгоритм Евклида
func ModInverseBig(a, p *big.Int) (*big.Int, error) {
	// Проверка, что a < p, и приведение к нужному диапазону
	aMod := new(big.Int).Mod(a, p)
	// Вычисляем GCD и коэффициенты с помощью GCDExtendedBig
	gcd, _, y := GCDExtendedBig(aMod, p)
	if gcd.Cmp(big.NewInt(1)) != 0 {
		return nil, fmt.Errorf("no modular inverse exists for %s mod %s", a.String(), p.String())
	}
	// Убедимся, что y положительное
	if y.Cmp(big.NewInt(0)) < 0 {
		y.Add(y, p) // Приведение к положительному значению, если y < 0
	}
	return y, nil
}

//// GCDExtended реализует расширенный алгоритм Евклида для нахождения GCD
//func GCDExtendedBig(a, b *big.Int) (*big.Int, *big.Int, *big.Int) {
//	x0, x1 := big.NewInt(1), big.NewInt(0)
//	y0, y1 := big.NewInt(0), big.NewInt(1)
//	origB := new(big.Int).Set(b)
//
//	for a.Cmp(big.NewInt(0)) != 0 {
//		q := new(big.Int).Div(b, a)
//		b, a = a, new(big.Int).Mod(b, a)
//		x0, x1 = x1, new(big.Int).Sub(x0, new(big.Int).Mul(q, x1))
//		y0, y1 = y1, new(big.Int).Sub(y0, new(big.Int).Mul(q, y1))
//	}
//
//	if b.Cmp(big.NewInt(1)) != 0 {
//		return big.NewInt(0), big.NewInt(0), big.NewInt(0) // Нет решения
//	}
//
//	return b, x0.Mod(x0, origB), y0
//}
