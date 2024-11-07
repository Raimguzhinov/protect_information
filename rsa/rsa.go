package rsa

import (
	"crypto/md5"
	"fmt"
	"github.com/Raimguzhinov/protect_information/common"
	"io"
	"os"
)

type rsaCipher struct {
	P, Q, N, Phi      int64
	PrivateC, PublicD int64
	Input             io.Reader
	OutputSigned      io.ReadWriter
	OutputEncrypted   io.Writer
	OutputDecrypted   io.Writer
	buffer            []int64
	msgBuf            []byte
	signature         int64
}

func newRsaAlgorithm() (*rsaCipher, error) {
	P := common.GenPrime(1000, 50000)
	Q := common.GenPrime(1000, 50000)
	c := &rsaCipher{
		P: P,
		Q: Q,
	}
	// Вычисляем N = P * Q и Phi = (P-1)*(Q-1)
	c.N = c.P * c.Q
	c.Phi = (c.P - 1) * (c.Q - 1)
	var err error
	c.PublicD = common.GenCoprime(c.Phi, 2, c.Phi-1)
	c.PrivateC, err = common.ModInverse(c.PublicD, c.Phi)
	if err != nil {
		return nil, fmt.Errorf("не удалось найти инверсию: %v", err)
	}
	return c, nil
}

func NewCipher(input io.Reader, encOut, decOut io.Writer) (common.Cipher, error) {
	c, err := newRsaAlgorithm()
	if err != nil {
		return nil, err
	}
	c.Input = input
	c.OutputEncrypted = encOut
	c.OutputDecrypted = decOut
	return c, nil
}

func (rc *rsaCipher) Encrypt() error {
	message, err := io.ReadAll(rc.Input)
	if err != nil {
		return err
	}
	encryptedMessage := make([]int64, len(message))
	for i, byteVal := range message {
		encryptedMessage[i] = common.ModularExponentiation(int64(byteVal), rc.PublicD, rc.N)
	}
	rc.buffer = encryptedMessage
	return common.WriteNumbers(rc.OutputEncrypted, encryptedMessage)
}

func (rc *rsaCipher) Decrypt() error {
	decryptedMessage := make([]int64, len(rc.buffer))
	for i, encVal := range rc.buffer {
		decryptedMessage[i] = common.ModularExponentiation(encVal, rc.PrivateC, rc.N)
	}
	return common.WriteNumbers(rc.OutputDecrypted, decryptedMessage)
}

// Do - объединяет шифрование и дешифрование
func (rc *rsaCipher) EncryptAndDecrypt() error {
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

func NewSignature(input io.Reader, output io.ReadWriter) (common.Signer, error) {
	c, err := newRsaAlgorithm()
	if err != nil {
		return nil, err
	}
	c.Input = input
	c.OutputSigned = output
	return c, nil
}

func (rc *rsaCipher) Sign() error {
	message, err := io.ReadAll(rc.Input)
	if err != nil {
		return err
	}
	hash := md5.Sum(message)
	hashInt := int64(hash[0])
	rc.signature = common.ModularExponentiation(hashInt, rc.PrivateC, rc.N)
	rc.msgBuf = message
	return common.WriteNumbers(rc.OutputSigned, []int64{rc.signature})
}

func (rc *rsaCipher) Verify() (bool, error) {
	hash := md5.Sum(rc.msgBuf)
	hashInt := int64(hash[0])
	w := common.ModularExponentiation(rc.signature, rc.PublicD, rc.N)
	return hashInt == w, nil
}

func (rc *rsaCipher) SignAndVerify() error {
	pwd, err := os.Getwd()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	fmt.Print("Signed: ")
	if err = rc.Sign(); err != nil {
		return err
	}
	if _, ok := rc.Input.(*os.File); ok {
		fmt.Print(pwd + "/" + rc.OutputSigned.(*os.File).Name())
	}
	if out, ok := rc.OutputSigned.(*os.File); ok {
		_ = out.Sync()
		defer out.Close()
		rc.OutputSigned, err = os.OpenFile(out.Name(), os.O_RDONLY, 0600)
		if err != nil {
			return err
		}
	}
	fmt.Print("\nVerified: ")
	ok, err := rc.Verify()
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
