package main

import (
	"fmt"
	"github.com/Raimguzhinov/protect-information/common"
	"github.com/Raimguzhinov/protect-information/elgamal"
	"github.com/Raimguzhinov/protect-information/gost"
	"github.com/Raimguzhinov/protect-information/rsa"
	"github.com/Raimguzhinov/protect-information/shamir"
	"github.com/Raimguzhinov/protect-information/vernam"
	"github.com/erikgeiser/promptkit/confirmation"
	"github.com/erikgeiser/promptkit/selection"
	"github.com/erikgeiser/promptkit/textinput"
	"github.com/ktr0731/go-fuzzyfinder"
	"io"
	"log"
	"math/big"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

func scanFilesInDirectory(dirName string, mu *sync.RWMutex, files *[]string) {
	_ = filepath.Walk(os.Getenv("HOME")+"/"+dirName, func(filepath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && !strings.Contains(filepath, "/.") {
			mu.Lock()
			*files = append(*files, filepath)
			mu.Unlock()
			time.Sleep(10 * time.Millisecond)
		}
		return nil
	})
}

// Функция для нечеткого поиска файлов с помощью fzf
func fuzzyFileSearch() (string, error) {
	var files []string
	var mu sync.RWMutex
	go func(files *[]string) {
		//dirs := []string{"Univer"}
		dirs := []string{"Downloads", "Documents", "Pictures", "Videos", "Univer"}
		for _, dir := range dirs {
			scanFilesInDirectory(dir, &mu, files)
		}
	}(&files)
	idx, err := fuzzyfinder.Find(
		&files,
		func(i int) string {
			return files[i]
		},
		fuzzyfinder.WithHotReloadLock(mu.RLocker()),
	)
	if err != nil {
		return "", err
	}
	return files[idx], nil
}

// Функция для запроса простого числа p
func promptForPrime() (int64, error) {
	confirm := confirmation.New("Generate a random prime number?", confirmation.Yes)
	confirmed, err := confirm.RunPrompt()
	if err != nil {
		return 0, err
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
		fmt.Printf("New prime number is: %d\n", P)
		return P, nil
	}
	// Ввод значения p
	prompt := textinput.New("Enter prime number p:")
	prompt.Placeholder = "Example: 7, 11, 53, 131, 997"
	response, err := prompt.RunPrompt()
	if err != nil {
		return 0, err
	}
	var p int64
	_, err = fmt.Sscanf(response, "%d", &p)
	if err != nil {
		return 0, err
	}
	return p, nil
}

func promptForPrimeWithRoot() (int64, int64, error) {
	confirm := confirmation.New("Generate a random prime number?", confirmation.Yes)
	confirmed, err := confirm.RunPrompt()
	if err != nil {
		return 0, 0, err
	}
	if confirmed {
		var P, q int64
		minV, maxV := 1_000_000_000, 1_000_000_0000
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

// Функция для выбора режима работы
func promptForMode() (string, error) {
	selectionPrompt := selection.New[string]("Select mode:", []string{"file", "text"})
	mode, err := selectionPrompt.RunPrompt()
	if err != nil {
		return "", err
	}
	return mode, nil
}

// Функция для ввода имени файла
func promptForFileName() (string, error) {
	file, err := fuzzyFileSearch()
	if err != nil {
		return "", err
	}
	return file, nil
}

// Функция для выбора шифра
func promptForCipher() (string, error) {
	selectionPrompt := selection.New[string]("Select cipher:", []string{"shamir", "vernam", "elgamal", "rsa"})
	cipher, err := selectionPrompt.RunPrompt()
	if err != nil {
		return "", err
	}
	return cipher, nil
}

// Функция для выбора подписи
func promptForSignature() (string, error) {
	selectionPrompt := selection.New[string]("Select signature:", []string{"elgamal", "rsa", "ГОСТ"})
	signature, err := selectionPrompt.RunPrompt()
	if err != nil {
		return "", err
	}
	return signature, nil
}

func InteractiveEncryptAndDecrypt() error {
	var (
		input           io.ReadCloser
		outputEncrypted io.WriteCloser
		outputDecrypted io.WriteCloser
		wg              sync.WaitGroup
		cipher          common.Cipher
	)
	mode, err := promptForMode()
	if err != nil {
		return fmt.Errorf("error selecting mode: %v", err)
	}
	wg.Add(1)
	switch mode {
	case "text":
		prompt := textinput.New("Enter message m (text): ")
		prompt.Placeholder = "Example: hello, world"
		m, err := prompt.RunPrompt()
		if err != nil {
			return err
		}
		input = io.NopCloser(strings.NewReader(m))
		outputEncrypted = os.Stdout
		outputDecrypted = os.Stdout
		wg.Done()
	case "file":
		wg.Add(2)
		var inputFile, outputEncFile, outputDecFile string
		inputFile, err = promptForFileName()
		if err != nil {
			return fmt.Errorf("error selecting input file: %v", err)
		}
		go func() {
			defer wg.Done()
			input, err = os.Open(inputFile)
			if err != nil {
				log.Fatalf("Error: %v", err)
			}
		}()
		defer func() {
			_ = input.Close()
		}()
		encPrompt := textinput.New("Enter name for the encrypted output file:")
		encPrompt.Placeholder = "Example: encrypted_output.dat"
		outputEncFile, err = encPrompt.RunPrompt()
		if err != nil {
			return fmt.Errorf("error entering encrypted file name: %v", err)
		}
		go func() {
			defer wg.Done()
			outputEncrypted, err = os.OpenFile(outputEncFile, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0600)
			if err != nil {
				log.Fatalf("Error: %v", err)
			}
		}()
		defer func() {
			_ = outputEncrypted.Close()
		}()
		decPrompt := textinput.New("Enter name for the decrypted output file:")
		decPrompt.Placeholder = "Example: decrypted_output.dat"
		outputDecFile, err = decPrompt.RunPrompt()
		if err != nil {
			return fmt.Errorf("error entering decrypted file name: %v", err)
		}
		go func() {
			defer wg.Done()
			outputDecrypted, err = os.OpenFile(outputDecFile, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)
			if err != nil {
				log.Fatalf("Error: %v", err)
			}
		}()
		defer func() {
			_ = outputDecrypted.Close()
		}()
	default:
		return fmt.Errorf("invalid action: %s", mode)
	}
	wg.Wait()
	cipherName, err := promptForCipher()
	if err != nil {
		return err
	}
	switch cipherName {
	case "shamir":
		p, err := promptForPrime() // Выбор простого числа
		if err != nil {
			return fmt.Errorf("error selecting prime number: %v", err)
		}
		cipher, err = shamir.NewCipher(p, input, outputEncrypted, outputDecrypted)
		if err != nil {
			return err
		}
	case "vernam":
		cipher, err = vernam.NewCipher(input, outputEncrypted, outputDecrypted)
		if err != nil {
			return err
		}
	case "elgamal":
		p, g, err := promptForPrimeWithRoot()
		if err != nil {
			return fmt.Errorf("error selecting prime number: %v", err)
		}
		cipher, err = elgamal.NewCipher(p, g, input, outputEncrypted, outputDecrypted)
		if err != nil {
			return err
		}
	case "rsa":
		cipher, err = rsa.NewCipher(input, outputEncrypted, outputDecrypted)
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("invalid cipher: %s", cipherName)
	}
	return cipher.EncryptAndDecrypt()
}

func InteractiveSignature() error {
	var (
		input     io.ReadCloser
		output    io.ReadWriteCloser
		wg        sync.WaitGroup
		signature common.Signer
	)
	mode, err := promptForMode()
	if err != nil {
		return fmt.Errorf("error selecting mode: %v", err)
	}
	wg.Add(1)
	switch mode {
	case "text":
		prompt := textinput.New("Enter message m (text): ")
		prompt.Placeholder = "Example: hello, world"
		m, err := prompt.RunPrompt()
		if err != nil {
			return err
		}
		input = io.NopCloser(strings.NewReader(m))
		output, err = os.OpenFile("/Users/yanakosteva/Study/protect_information/signature.dat", os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0600)
		if err != nil {
			log.Fatalf("Error: %v", err)
		}
		wg.Done()
	case "file":
		wg.Add(1)
		var inputFile, outputSignFile string
		inputFile, err = promptForFileName()
		if err != nil {
			return fmt.Errorf("error selecting input file: %v", err)
		}
		go func() {
			defer wg.Done()
			input, err = os.Open(inputFile)
			if err != nil {
				log.Fatalf("Error: %v", err)
			}
		}()
		defer func() {
			_ = input.Close()
		}()
		signPrompt := textinput.New("Enter name for the signed output file:")
		signPrompt.Placeholder = "Example: signed_output.dat"
		outputSignFile, err = signPrompt.RunPrompt()
		if err != nil {
			return fmt.Errorf("error entering signed file name: %v", err)
		}
		go func() {
			defer wg.Done()
			output, err = os.OpenFile(outputSignFile, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0600)
			if err != nil {
				log.Fatalf("Error: %v", err)
			}
		}()
		defer func() {
			_ = output.Close()
		}()
	default:
		return fmt.Errorf("invalid action: %s", mode)
	}
	wg.Wait()
	signatureName, err := promptForSignature()
	if err != nil {
		return err
	}
	switch signatureName {
	case "elgamal":
		var P, q *big.Int
		minV := big.NewInt(1_000_000_000)
		maxV := big.NewInt(10_000_000_000)
		// Генерация простого числа q и проверка, что P = 2q + 1 также простое
		for {
			q = common.GenPrimeBig(minV, maxV)
			P = new(big.Int).Mul(q, big.NewInt(2))
			P.Add(P, big.NewInt(1)) // P = 2 * q + 1

			if common.IsPrimeBig(P) {
				break
			}
		}
		// Поиск примитивного корня g
		g := big.NewInt(0)
		for i := big.NewInt(2); i.Cmp(new(big.Int).Sub(P, big.NewInt(1))) < 0; i.Add(i, big.NewInt(1)) {
			g.Set(i)
			if common.ModularExponentiationBig(g, q, P).Cmp(big.NewInt(1)) != 0 {
				break
			}
		}
		signature, err = elgamal.NewSignature(P, g, input, output)
		if err != nil {
			return err
		}
	case "rsa":
		signature, err = rsa.NewSignature(input, output)
		if err != nil {
			return err
		}
	case "ГОСТ":
		signature, err = gost.NewSignature(input, output)
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("invalid cipher: %s", signatureName)
	}
	return signature.SignAndVerify()
}

func main() {
	if err := InteractiveSignature(); err != nil {
		log.Fatalf("Error: %v", err)
	}
}
