package elgamal

import (
	"crypto/md5"
	"encoding/binary"
	"fmt"
	"github.com/Raimguzhinov/protect_information/common"
	"io"
	"math/big"
	"os"
)

type ElgamalCipher struct {
	P, G, Y         int64 // Публичные параметры: простое число p, основание g, публичный ключ Y = g^X mod p
	X               int64 // Приватный ключ X
	R               int64
	Input           io.Reader
	OutputSigned    io.ReadWriter
	OutputEncrypted io.Writer
	OutputDecrypted io.Writer
	buffer          [][2]int64 // Буфер для хранения зашифрованных пар (r, e)
	msg             []byte
	signature       int64
	msgBuf          []byte
}

func newElgamalAlgorithm(p, g int64) (*ElgamalCipher, error) {
	c := &ElgamalCipher{
		P: p,
		G: g,
	}

	// Генерация приватного и публичного ключей
	var err error
	c.X, c.Y, err = generateKeyPair(c.P, c.G)
	if err != nil {
		return nil, err
	}
	return c, nil
}

// NewCipher - конструктор структуры для шифра Эль-Гамаля
func NewCipher(p, g int64, input io.Reader, encOut, decOut io.Writer) (common.Cipher, error) {
	c, err := newElgamalAlgorithm(p, g)
	if err != nil {
		return nil, err
	}
	c.Input = input
	c.OutputEncrypted = encOut
	c.OutputDecrypted = decOut
	return c, nil
}

// Генерация ключевой пары (X, Y), где Y = G^X mod P
func generateKeyPair(p, g int64) (int64, int64, error) {
	x := common.Seed().Int63n(p-1) + 1         // Приватный ключ X
	y := common.ModularExponentiation(g, x, p) // Публичный ключ Y = G^X mod P
	return x, y, nil
}

func (ec *ElgamalCipher) Encrypt() error {
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

func (ec *ElgamalCipher) Decrypt() error {
	decryptedMessage := make([]byte, len(ec.buffer))
	for i, pair := range ec.buffer {
		r, e := pair[0], pair[1]
		// Дешифрование: M = e * (r^(P-1-X) mod P)
		s := common.ModularExponentiation(r, ec.P-1-ec.X, ec.P)
		m := (e * s) % ec.P
		decryptedMessage[i] = byte(m)
	}
	ec.msg = decryptedMessage
	return common.WriteData(ec.OutputDecrypted, decryptedMessage)
}

// Do - объединяет шифрование и дешифрование
func (ec *ElgamalCipher) EncryptAndDecrypt() error {
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

func (ec *ElgamalCipher) GetDecryptMsg() []byte {
	return ec.msg
}

func NewSignature(p, g int64, input io.Reader, output io.ReadWriter) (common.Signer, error) {
	c, err := newElgamalAlgorithm(p, g)
	if err != nil {
		return nil, err
	}
	c.Input = input
	c.OutputSigned = output
	return c, nil
}

func (c *ElgamalCipher) Sign() error {
	message, err := io.ReadAll(c.Input)
	if err != nil {
		return err
	}
	hash := md5.Sum(message)
	hashInt0 := new(big.Int).SetUint64(binary.BigEndian.Uint64(hash[:8]))
	pBigInt := big.NewInt(c.P)
	hashInt := hashInt0.Mod(hashInt0, pBigInt).Int64()

	if hashInt >= c.P || hashInt <= 0 {
		return fmt.Errorf("hashInt >= c.P || hashInt <= 0")
	}
	k := common.GenCoprime(c.P-1, 2, c.P-2)
	c.R = common.ModularExponentiation(c.G, k, c.P)
	u := (hashInt - c.X*c.R) % (c.P - 1)
	u = (u + (c.P - 1)) % (c.P - 1)
	gcd, k1, _ := common.GCDExtended(k, c.P-1)
	if gcd != 1 {
		return fmt.Errorf("gcd(%d, %d) != 1", k, c.P-1)
	}
	c.signature = (u * k1) % (c.P - 1)
	c.msgBuf = message
	return common.WriteNumbers(c.OutputSigned, []int64{c.signature})
}

func (c *ElgamalCipher) Verify() (bool, error) {
	hash := md5.Sum(c.msgBuf)
	hashInt0 := new(big.Int).SetUint64(binary.BigEndian.Uint64(hash[:8]))
	pBigInt := big.NewInt(c.P)
	hashInt := hashInt0.Mod(hashInt0, pBigInt).Int64()
	yr := common.ModularExponentiation(c.Y, c.R, c.P) * common.ModularExponentiation(c.R, c.signature, c.P) % c.P
	g := common.ModularExponentiation(c.G, hashInt, c.P)
	return yr == g, nil
}

func (c *ElgamalCipher) SignAndVerify() error {
	pwd, err := os.Getwd()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	fmt.Print("Signed: ")
	if err = c.Sign(); err != nil {
		return err
	}
	if _, ok := c.Input.(*os.File); ok {
		fmt.Print(pwd + "/" + c.OutputSigned.(*os.File).Name())
	}
	if out, ok := c.OutputSigned.(*os.File); ok {
		_ = out.Sync()
		defer out.Close()
		c.OutputSigned, err = os.OpenFile(out.Name(), os.O_RDONLY, 0600)
		if err != nil {
			return err
		}
	}
	fmt.Print("\nVerified: ")
	ok, err := c.Verify()
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
