package vernam

import (
	"fmt"
	"io"
	"os"
	"protect_information/common"
)

type vernamCipher struct {
	Key             []byte
	Input           io.Reader
	OutputEncrypted io.Writer
	OutputDecrypted io.Writer
	buffer          []byte
}

func NewCipher(input io.Reader, encOut, decOut io.Writer) (common.Cipher, error) {
	c := &vernamCipher{
		Input:           input,
		OutputEncrypted: encOut,
		OutputDecrypted: decOut,
	}
	return c, nil
}

func (vc *vernamCipher) Encrypt() error {
	message, err := io.ReadAll(vc.Input)
	if err != nil {
		return err
	}
	// Генерация ключа той же длины, что и сообщение
	vc.Key = vc.generateKey(len(message))
	// Шифрование с помощью побитовой операции XOR
	encryptedMessage := make([]byte, len(message))
	for i := range message {
		encryptedMessage[i] = message[i] ^ vc.Key[i]
	}
	vc.buffer = encryptedMessage
	return common.WriteData(vc.OutputEncrypted, encryptedMessage)
}

func (vc *vernamCipher) Decrypt() error {
	// Дешифрование с помощью побитовой операции XOR (обратное шифрование)
	decryptedMessage := make([]byte, len(vc.buffer))
	for i := range vc.buffer {
		decryptedMessage[i] = vc.buffer[i] ^ vc.Key[i]
	}
	return common.WriteData(vc.OutputDecrypted, decryptedMessage)
}

// Do - объединяет шифрование и дешифрование
func (vc *vernamCipher) Do() error {
	pwd, err := os.Getwd()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	fmt.Print("Encrypted: ")
	if err = vc.Encrypt(); err != nil {
		return err
	}
	if _, ok := vc.Input.(*os.File); ok {
		fmt.Print(pwd + "/" + vc.OutputEncrypted.(*os.File).Name())
	}
	fmt.Print("\nDecrypted: ")
	defer func() {
		if _, ok := vc.Input.(*os.File); ok {
			fmt.Print(pwd + "/" + vc.OutputDecrypted.(*os.File).Name())
		}
		fmt.Print("\n")
	}()
	return vc.Decrypt()
}

func (vc *vernamCipher) generateKey(length int) []byte {
	key := make([]byte, length)
	_, err := common.Seed().Read(key)
	if err != nil {
		panic(fmt.Sprintf("failed to generate key: %v", err))
	}
	return key
}
