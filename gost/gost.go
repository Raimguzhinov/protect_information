package gost

import (
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"github.com/Raimguzhinov/protect-information/common"
	"io"
	"math/big"
	"os"
)

// gostSignature содержит параметры и результаты подписи
type gostSignature struct {
	P, Q, A      *big.Int // Параметры
	PrivateKey   *big.Int // Приватный ключ
	PublicKey    *big.Int // Публичный ключ
	SignatureR   *big.Int // Часть подписи R
	SignatureS   *big.Int // Часть подписи S
	Input        io.Reader
	OutputSigned io.Writer
	Message      []byte // Сообщение для подписи
}

// Генерация случайного числа в диапазоне [min, max)
func genRandomInRange(min, max *big.Int) (*big.Int, error) {
	diff := new(big.Int).Sub(max, min)
	num, err := rand.Int(rand.Reader, diff)
	if err != nil {
		return nil, fmt.Errorf("ошибка генерации случайного числа: %v", err)
	}
	return num.Add(num, min), nil
}

// Генерация параметров p, q, a для ГОСТ с использованием функций из common
func generateGOSTParams() (*big.Int, *big.Int, *big.Int, error) {
	// Шаг 1: Генерируем 256-битное простое число q с использованием common.GenPrimeBig
	qMin := new(big.Int).Lsh(big.NewInt(1), 255)                                  // Минимальное значение для 256 бит
	qMax := new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 256), big.NewInt(1)) // Максимальное значение для 256 бит
	q := common.GenPrimeBig(qMin, qMax)
	// Шаг 2: Генерируем 1024-битное простое число p = b*q + 1
	var p, b *big.Int
	pMin := new(big.Int).Lsh(big.NewInt(1), 1023)                                  // Минимальное значение для 1024 бит
	pMax := new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 1024), big.NewInt(1)) // Максимальное значение для 1024 бит
	for {
		// Генерируем случайное значение для b в диапазоне [ceil(pMin/q), floor(pMax/q)]
		bMin := new(big.Int).Add(new(big.Int).Div(pMin, q), big.NewInt(1))
		bMax := new(big.Int).Div(pMax, q)
		// Выбираем случайное b в диапазоне [bMin, bMax]
		var err error
		b, err = genRandomInRange(bMin, bMax)
		if err != nil {
			return nil, nil, nil, fmt.Errorf("ошибка генерации b: %v", err)
		}
		// Вычисляем p = b*q + 1
		p = new(big.Int).Mul(b, q)
		p.Add(p, big.NewInt(1))
		// Проверяем, что p является простым
		if p.ProbablyPrime(20) {
			break
		}
	}
	// Шаг 3: Находим a такое, что a = g^b mod p и a > 1
	var a *big.Int
	for {
		// Генерируем случайное значение g в диапазоне [2, p-2]
		gMin := big.NewInt(2)
		gMax := new(big.Int).Sub(p, big.NewInt(2))
		g, err := genRandomInRange(gMin, gMax)
		if err != nil {
			return nil, nil, nil, fmt.Errorf("ошибка генерации g: %v", err)
		}
		// Вычисляем a = g^b mod p
		a = common.ModularExponentiationBig(g, b, p)
		// Проверяем, что a > 1
		if a.Cmp(big.NewInt(1)) > 0 {
			break
		}
	}
	return p, q, a, nil
}

func NewSignature(input io.Reader, output io.ReadWriter) (common.Signer, error) {
	p, q, a, err := generateGOSTParams()
	if err != nil {
		fmt.Printf("Ошибка генерации параметров: %v\n", err)
		return nil, err
	}
	gs := &gostSignature{P: p, Q: q, A: a, Input: input, OutputSigned: output}
	gs.GenerateKeys()
	return gs, nil
}

func (gs *gostSignature) GenerateKeys() {
	// Приватный ключ x — случайное число в диапазоне [1, q-1)
	gs.PrivateKey = common.GenCoprimeBig(gs.Q, big.NewInt(1), new(big.Int).Sub(gs.Q, big.NewInt(1)))
	// Публичный ключ y = a^x mod p
	gs.PublicKey = common.ModularExponentiationBig(gs.A, gs.PrivateKey, gs.P)
}

// Подпись сообщения
func (gs *gostSignature) Sign() error {
	message, err := io.ReadAll(gs.Input)
	gs.Message = message
	if err != nil {
		return err
	}
	// Хешируем сообщение
	hash := sha256.Sum256(message)
	hashInt := new(big.Int).SetBytes(hash[:])
	// Убедимся, что hashInt лежит в диапазоне [0, q)
	if hashInt.Cmp(gs.Q) >= 0 {
		hashInt.Mod(hashInt, gs.Q)
	}
	fmt.Printf("Message hash (as int): %s\n", hashInt.String())
	// Шаги 2–4: Генерация случайного числа k и вычисление R и S
	for {
		k := common.GenCoprimeBig(gs.Q, big.NewInt(1), new(big.Int).Sub(gs.Q, big.NewInt(1)))
		r := common.ModularExponentiationBig(gs.A, k, gs.P)
		r.Mod(r, gs.Q)
		if r.Cmp(big.NewInt(0)) == 0 {
			continue // Если R = 0, снова выбираем k
		}
		// Вычисляем s = (k*h + x*r) mod q
		s := new(big.Int).Mul(k, hashInt)
		s.Add(s, new(big.Int).Mul(gs.PrivateKey, r))
		s.Mod(s, gs.Q)
		if s.Cmp(big.NewInt(0)) == 0 {
			continue // Если S = 0, снова выбираем k
		}
		// Успешная подпись, сохраняем R и S
		gs.SignatureR = r
		gs.SignatureS = s
		break
	}
	fmt.Printf("Подпись создана: R = %s, S = %s", gs.SignatureR.String(), gs.SignatureS.String())
	return nil
}

// Проверка подписи
func (gs *gostSignature) Verify() (bool, error) {
	// Хешируем сообщение
	hash := sha256.Sum256(gs.Message)
	hashInt := new(big.Int).SetBytes(hash[:])
	fmt.Printf("Message hash (as int): %s\n", hashInt.String())
	// Проверка неравенств для R и S
	if gs.SignatureR.Cmp(big.NewInt(0)) <= 0 || gs.SignatureR.Cmp(gs.Q) >= 0 {
		return false, nil
	}
	if gs.SignatureS.Cmp(big.NewInt(0)) <= 0 || gs.SignatureS.Cmp(gs.Q) >= 0 {
		return false, nil
	}
	// Вычисляем h^(-1) mod q с помощью common.GCDExtendedBig
	gcd, hInv, _ := common.GCDExtendedBig(hashInt, gs.Q)
	if gcd.Cmp(big.NewInt(1)) != 0 {
		return false, nil // Обратного элемента не существует
	}
	hInv.Mod(hInv, gs.Q)
	// Вычисляем u1 и u2
	u1 := new(big.Int).Mul(gs.SignatureS, hInv)
	u1.Mod(u1, gs.Q)
	u2 := new(big.Int).Mul(new(big.Int).Neg(gs.SignatureR), hInv)
	u2.Mod(u2, gs.Q)
	// Вычисляем v = (a^u1 * y^u2 mod p) mod q
	v1 := common.ModularExponentiationBig(gs.A, u1, gs.P)
	v2 := common.ModularExponentiationBig(gs.PublicKey, u2, gs.P)
	v := new(big.Int).Mul(v1, v2)
	v.Mod(v, gs.P)
	v.Mod(v, gs.Q)
	// Сравниваем v и R
	return v.Cmp(gs.SignatureR) == 0, nil
}

func (gs *gostSignature) SignAndVerify() error {
	pwd, err := os.Getwd()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	fmt.Print("Signed: ")
	if err = gs.Sign(); err != nil {
		return err
	}
	if _, ok := gs.Input.(*os.File); ok {
		fmt.Print(pwd + "/" + gs.OutputSigned.(*os.File).Name())
	}
	if out, ok := gs.OutputSigned.(*os.File); ok {
		_ = out.Sync()
		defer out.Close()
		gs.OutputSigned, err = os.OpenFile(out.Name(), os.O_RDONLY, 0600)
		if err != nil {
			return err
		}
	}
	fmt.Print("\nVerified: ")
	ok, err := gs.Verify()
	if err != nil {
		return err
	}
	if ok {
		fmt.Println("true")
	} else {
		fmt.Println("false")
	}
	return nil
}
