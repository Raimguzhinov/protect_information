package elgamal

import (
	"fmt"
	"io"
	"os"
	"protect_information/common"
)

type elgamalCipher struct {
	P, G, Y         int64 // Публичные параметры: простое число p, основание g, публичный ключ Y = g^X mod p
	X               int64 // Приватный ключ X
	Input           io.Reader
	OutputEncrypted io.Writer
	OutputDecrypted io.Writer
	buffer          [][2]int64 // Буфер для хранения зашифрованных пар (r, e)
}

// NewCipher - конструктор структуры для шифра Эль-Гамаля
func NewCipher(p, g int64, input io.Reader, encOut, decOut io.Writer) (common.Cipher, error) {
	c := &elgamalCipher{
		P:               p,
		G:               g,
		Input:           input,
		OutputEncrypted: encOut,
		OutputDecrypted: decOut,
	}

	// Генерация приватного и публичного ключей
	var err error
	c.X, c.Y, err = generateKeyPair(c.P, c.G)
	if err != nil {
		return nil, err
	}
	return c, nil
}

// Генерация ключевой пары (X, Y), где Y = G^X mod P
func generateKeyPair(p, g int64) (int64, int64, error) {
	x := common.Seed().Int63n(p-1) + 1         // Приватный ключ X
	y := common.ModularExponentiation(g, x, p) // Публичный ключ Y = G^X mod P
	return x, y, nil
}

func (ec *elgamalCipher) Encrypt() error {
	message, err := io.ReadAll(ec.Input)
	if err != nil {
		return err
	}
	encryptedMessage := make([][2]int64, len(message))
	for i, byteVal := range message {
		if int64(byteVal) >= ec.P {
			return fmt.Errorf("byte %d is greater than or equal to p", byteVal)
		}
		// Случайное значение k
		k := common.Seed().Int63n(ec.P-1) + 1
		// Шифрование: r = G^k mod P, e = M * Y^k mod P
		r := common.ModularExponentiation(ec.G, k, ec.P)
		e := (int64(byteVal) * common.ModularExponentiation(ec.Y, k, ec.P)) % ec.P
		encryptedMessage[i] = [2]int64{r, e}
	}
	ec.buffer = encryptedMessage
	return common.WritePair(ec.OutputEncrypted, encryptedMessage)
}

func (ec *elgamalCipher) Decrypt() error {
	decryptedMessage := make([]byte, len(ec.buffer))
	for i, pair := range ec.buffer {
		r, e := pair[0], pair[1]
		// Дешифрование: M = e * (r^(P-1-X) mod P)
		s := common.ModularExponentiation(r, ec.P-1-ec.X, ec.P)
		m := (e * s) % ec.P
		decryptedMessage[i] = byte(m)
	}
	return common.WriteData(ec.OutputDecrypted, decryptedMessage)
}

// Do - объединяет шифрование и дешифрование
func (ec *elgamalCipher) Do() error {
	pwd, err := os.Getwd()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	fmt.Print("Encrypted: ")
	if err = ec.Encrypt(); err != nil {
		return err
	}
	if _, ok := ec.Input.(*os.File); ok {
		fmt.Print(pwd + "/" + ec.OutputEncrypted.(*os.File).Name())
	}
	fmt.Print("\nDecrypted: ")
	defer func() {
		if _, ok := ec.Input.(*os.File); ok {
			fmt.Print(pwd + "/" + ec.OutputDecrypted.(*os.File).Name())
		}
		fmt.Print("\n")
	}()
	return ec.Decrypt()
}
