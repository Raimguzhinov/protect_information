package shamir

import (
	"fmt"
	"io"
	"os"
	"protect_information/common"
)

type shamirCipher struct {
	P               int64
	CA, DA          int64
	CB, DB          int64
	Input           io.Reader
	OutputEncrypted io.Writer
	OutputDecrypted io.Writer
	buffer          []int64
}

func NewCipher(p int64, input io.Reader, encOut, decOut io.Writer) (common.Cipher, error) {
	c := &shamirCipher{
		P:               p,
		Input:           input,
		OutputEncrypted: encOut,
		OutputDecrypted: decOut,
	}
	// Генерация ключей для Alice и Bob
	var err error
	c.CA, c.DA, err = generateKeyPair(c.P)
	if err != nil {
		return nil, err
	}
	c.CB, c.DB, err = generateKeyPair(c.P)
	if err != nil {
		return nil, err
	}
	return c, nil
}

func generateKeyPair(p int64) (int64, int64, error) {
	var c, d int64
	for {
		c = common.Seed().Int63n(p-1) + 1
		var err error
		d, err = generateSecretKey(c, p)
		if err == nil {
			break
		}
	}
	return c, d, nil
}

// Генерация секретного ключа с использованием расширенного алгоритма Евклида
func generateSecretKey(cA, p int64) (int64, error) {
	gcd, _, y := common.GCDExtended(p-1, cA)
	if gcd != 1 {
		return -1, fmt.Errorf("error: cA and p-1 are not relatively prime")
	}
	d := y
	if d < 0 {
		d += p - 1
	}
	return d, nil
}

func (sc *shamirCipher) Encrypt() error {
	message, err := io.ReadAll(sc.Input) //common.ReadNumbers(sc.Input)
	if err != nil {
		return err
	}
	encryptedMessage := make([]int64, len(message))
	for i, byteVal := range message {
		if int64(byteVal) >= sc.P {
			return fmt.Errorf("byte %d is greater than or equal to p", byteVal)
		}
		x1 := common.ModularExponentiation(int64(byteVal), sc.CA, sc.P)
		x2 := common.ModularExponentiation(x1, sc.CB, sc.P)
		encryptedMessage[i] = x2
	}
	sc.buffer = encryptedMessage
	return common.WriteNumbers(sc.OutputEncrypted, encryptedMessage)
}

func (sc *shamirCipher) Decrypt() error {
	decryptedMessage := make([]int64, len(sc.buffer))
	for i, byteVal := range sc.buffer {
		x3 := common.ModularExponentiation(byteVal, sc.DA, sc.P)
		x4 := common.ModularExponentiation(x3, sc.DB, sc.P)
		decryptedMessage[i] = x4
	}
	return common.WriteNumbers(sc.OutputDecrypted, decryptedMessage)
}

// Do - объединяет шифрование и дешифрование
func (sc *shamirCipher) Do() error {
	pwd, err := os.Getwd()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	fmt.Print("Encrypted: ")
	if err = sc.Encrypt(); err != nil {
		return err
	}
	if _, ok := sc.Input.(*os.File); ok {
		fmt.Print(pwd + "/" + sc.OutputEncrypted.(*os.File).Name())
	}
	fmt.Print("\nDecrypted: ")
	defer func() {
		if _, ok := sc.Input.(*os.File); ok {
			fmt.Print(pwd + "/" + sc.OutputDecrypted.(*os.File).Name())
		}
		fmt.Print("\n")
	}()
	return sc.Decrypt()
}
