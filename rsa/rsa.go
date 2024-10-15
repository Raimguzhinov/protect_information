package rsa

import (
	"fmt"
	"io"
	"os"
	"protect_information/common"
)

type rsaCipher struct {
	N, E, D         int64 // Параметры RSA: N = p*q, E - экспонента, D - приватный ключ
	Input           io.Reader
	OutputEncrypted io.Writer
	OutputDecrypted io.Writer
	buffer          []int64
}

// NewCipher - конструктор структуры для шифра RSA
func NewCipher(input io.Reader, encOut, decOut io.Writer) (common.Cipher, error) {
	c := &rsaCipher{
		Input:           input,
		OutputEncrypted: encOut,
		OutputDecrypted: decOut,
	}
	// Генерация публичного и приватного ключей RSA
	var err error
	c.N, c.E, c.D, err = generateKeyPair()
	if err != nil {
		return nil, err
	}
	return c, nil
}

// Генерация ключевой пары для RSA
func generateKeyPair() (n, e, d int64, err error) {
	// Генерация двух случайных простых чисел p и q
	p := common.GenPrime(1000, 5000)
	q := common.GenPrime(1000, 5000)
	n = p * q // N = p * q
	// Вычисление φ(n) = (p-1) * (q-1)
	phi := (p - 1) * (q - 1)
	// Выбор e, такое что gcd(e, φ(n)) = 1
	e = 65537 // Чаще всего используется 65537
	gcd, _, _, _ := common.GCDExtended(e, phi)
	if gcd != 1 {
		return 0, 0, 0, fmt.Errorf("e and phi(n) are not coprime")
	}
	// Вычисление d, где d = e^(-1) mod φ(n)
	d, err = common.ModInverse(e, phi)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("failed to compute modular inverse for d: %v", err)
	}
	return n, e, d, nil
}

func (rc *rsaCipher) Encrypt() error {
	message, err := io.ReadAll(rc.Input)
	if err != nil {
		return err
	}
	encryptedMessage := make([]int64, len(message))
	for i, byteVal := range message {
		m := int64(byteVal)
		// Шифрование: c = m^e mod n
		encryptedMessage[i] = common.ModularExponentiation(m, rc.E, rc.N)
	}
	rc.buffer = encryptedMessage
	return common.WriteNumbers(rc.OutputEncrypted, encryptedMessage)
}

func (rc *rsaCipher) Decrypt() error {
	decryptedMessage := make([]byte, len(rc.buffer))
	for i, c := range rc.buffer {
		// Дешифрование: m = c^d mod n
		m := common.ModularExponentiation(c, rc.D, rc.N)
		decryptedMessage[i] = byte(m)
	}
	return common.WriteData(rc.OutputDecrypted, decryptedMessage)
}

// Do - объединяет шифрование и дешифрование
func (rc *rsaCipher) Do() error {
	pwd, err := os.Getwd()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	fmt.Print("Encrypted: ")
	if err = rc.Encrypt(); err != nil {
		return err
	}
	if _, ok := rc.Input.(*os.File); ok {
		fmt.Print(pwd + "/" + rc.OutputEncrypted.(*os.File).Name())
	}
	fmt.Print("\nDecrypted: ")
	defer func() {
		if _, ok := rc.Input.(*os.File); ok {
			fmt.Print(pwd + "/" + rc.OutputDecrypted.(*os.File).Name())
		}
		fmt.Print("\n")
	}()
	return rc.Decrypt()
}
