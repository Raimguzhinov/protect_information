package vernam

import (
	"bytes"
	"fmt"
	"github.com/Raimguzhinov/protect-information/common"
	"github.com/Raimguzhinov/protect-information/elgamal"
	"github.com/erikgeiser/promptkit/confirmation"
	"github.com/erikgeiser/promptkit/textinput"
	"io"
	"os"
)

type vernamCipher struct {
	Key             []byte
	Input           io.Reader
	OutputEncrypted io.Writer
	OutputDecrypted io.Writer
	buffer          []byte
	cipher          common.Cipher
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
	outputEncrypted, err := os.OpenFile("enkey.dat", os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0600)
	outputDecrypted, err := os.OpenFile("deckey.dat", os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0600)

	p, g, err := promptForPrimeWithRoot()
	vc.cipher, err = elgamal.NewCipher(p, g, bytes.NewReader(vc.Key), outputEncrypted, outputDecrypted)
	if err != nil {
		return err
	}
	err = vc.cipher.Encrypt()
	if err != nil {
		return err
	}

	return common.WriteData(vc.OutputEncrypted, encryptedMessage)
}

func (vc *vernamCipher) Decrypt() error {
	err := vc.cipher.Decrypt()
	if err != nil {
		return err
	}
	vc.Key = vc.cipher.(*elgamal.ElgamalCipher).GetDecryptMsg()
	// Дешифрование с помощью побитовой операции XOR (обратное шифрование)
	decryptedMessage := make([]byte, len(vc.buffer))
	for i := range vc.buffer {
		decryptedMessage[i] = vc.buffer[i] ^ vc.Key[i]
	}
	return common.WriteData(vc.OutputDecrypted, decryptedMessage)
}

// Do - объединяет шифрование и дешифрование
func (vc *vernamCipher) EncryptAndDecrypt() error {
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

func promptForPrimeWithRoot() (int64, int64, error) {
	confirm := confirmation.New("Generate a random prime number?", confirmation.Yes)
	confirmed, err := confirm.RunPrompt()
	if err != nil {
		return 0, 0, err
	}
	if confirmed {
		var P, q int64
		minV, maxV := 1_000_000, 1_000_000_000
		for {
			q = common.GenPrime(int64(minV), int64(maxV))
			P = 2*q + 1
			if common.IsPrime(P) {
				break
			}
		}
		g := int64(0)
		for i := int64(2); i < P-1; i++ {
			g = i
			if common.ModularExponentiation(g, q, P) != 1 {
				break
			}
		}
		fmt.Printf("New prime number is: %d root: %d\n", P, g)
		return P, g, nil
	}
	// Ввод значения p
	prompt := textinput.New("Enter prime number p:")
	prompt.Placeholder = "Example: 7, 11, 53, 131, 997"
	response, err := prompt.RunPrompt()
	if err != nil {
		return 0, 0, err
	}
	var p int64
	_, err = fmt.Sscanf(response, "%d", &p)
	if err != nil {
		return 0, 0, err
	}
	// Ввод значения g
	prompt = textinput.New("Enter primitive root g:")
	prompt.Placeholder = "Example: 2"
	response, err = prompt.RunPrompt()
	if err != nil {
		return 0, 0, err
	}
	var g int64
	_, err = fmt.Sscanf(response, "%d", &g)
	if err != nil {
		return 0, 0, err
	}
	return p, g, nil
}
