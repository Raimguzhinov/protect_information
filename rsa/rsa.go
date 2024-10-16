package rsa

import (
	"fmt"
	"github.com/Raimguzhinov/protect_information/common"
	"io"
	"os"
)

type rsaCipher struct {
	P, Q, N, Phi      int64
	PublicE, PrivateD int64
	Input             io.Reader
	OutputEncrypted   io.Writer
	OutputDecrypted   io.Writer
	buffer            []int64
}

func NewCipher(input io.Reader, encOut, decOut io.Writer) (common.Cipher, error) {
	P := common.GenPrime(1000, 50000)
	Q := common.GenPrime(1000, 50000)
	c := &rsaCipher{
		P:               P,
		Q:               Q,
		Input:           input,
		OutputEncrypted: encOut,
		OutputDecrypted: decOut,
	}
	// Вычисляем N = P * Q и Phi = (P-1)*(Q-1)
	c.N = c.P * c.Q
	c.Phi = (c.P - 1) * (c.Q - 1)
	var err error
	c.PrivateD = common.GenCoprime(c.Phi, 2, c.Phi)
	c.PublicE, err = common.ModInverse(c.PrivateD, c.Phi)
	if err != nil {
		return nil, fmt.Errorf("не удалось найти инверсию: %v", err)
	}
	return c, nil
}

func (rc *rsaCipher) Encrypt() error {
	message, err := io.ReadAll(rc.Input)
	if err != nil {
		return err
	}
	encryptedMessage := make([]int64, len(message))
	for i, byteVal := range message {
		encryptedMessage[i] = common.ModularExponentiation(int64(byteVal), rc.PublicE, rc.N)
	}
	rc.buffer = encryptedMessage
	return common.WriteNumbers(rc.OutputEncrypted, encryptedMessage)
}

func (rc *rsaCipher) Decrypt() error {
	decryptedMessage := make([]int64, len(rc.buffer))
	for i, encVal := range rc.buffer {
		decryptedMessage[i] = common.ModularExponentiation(encVal, rc.PrivateD, rc.N)
	}
	return common.WriteNumbers(rc.OutputDecrypted, decryptedMessage)
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
