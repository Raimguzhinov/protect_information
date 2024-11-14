package elgamal

import (
	"crypto/md5"
	"fmt"
	"github.com/Raimguzhinov/protect-information/common"
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

// EncryptAndDecrypt - объединяет шифрование и дешифрование
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

// elgamalSignature содержит параметры и результаты подписи
type elgamalSignature struct {
	P            *big.Int // Простое число (модуль)
	G            *big.Int // Основание (генератор группы)
	X            *big.Int // Секретный ключ
	Y            *big.Int // Публичный ключ Y = G^X mod P
	R            *big.Int // Часть подписи
	Signature    *big.Int // Подпись
	Input        io.Reader
	OutputSigned io.Writer
	Message      []byte // Сообщение для подписи
}

func NewSignature(p, g *big.Int, input io.Reader, output io.ReadWriter) (common.Signer, error) {
	es := elgamalSignature{
		P: p,
		G: g,
		X: GenerateX(p),
	}
	es.Y = common.ModularExponentiationBig(g, es.X, p)
	es.Input = input
	es.OutputSigned = output
	return &es, nil
}

func GenerateX(p *big.Int) *big.Int {
	one := big.NewInt(1)
	maximum := new(big.Int).Sub(p, one) // P - 1
	for {
		x := new(big.Int).Rand(common.SeedBig(), maximum)
		if x.Cmp(one) > 0 {
			return x
		}
	}
}

func (es *elgamalSignature) Sign() error {
	message, err := io.ReadAll(es.Input)
	if err != nil {
		return err
	}
	hash := md5.Sum(message)
	hashInt := new(big.Int).SetBytes(hash[:])
	// Приводим хеш к модулю (P - 1), чтобы h < P - 1
	hashInt.Mod(hashInt, new(big.Int).Sub(es.P, big.NewInt(1)))
	fmt.Printf("Message hash (as int): %s\n", hashInt.String())
	// k ∈ [2, P-2], gcd(k, P - 1) = 1
	k := common.GenCoprimeBig(new(big.Int).Sub(es.P, big.NewInt(1)), big.NewInt(2), new(big.Int).Sub(es.P, big.NewInt(2)))
	// R = G^k mod P
	es.R = common.ModularExponentiationBig(es.G, k, es.P)
	// u = (h - x*R) mod (P - 1)
	u := new(big.Int).Sub(hashInt, new(big.Int).Mul(es.X, es.R))
	u.Mod(u, new(big.Int).Sub(es.P, big.NewInt(1)))
	if u.Cmp(big.NewInt(0)) < 0 {
		u.Add(u, new(big.Int).Sub(es.P, big.NewInt(1)))
	}
	// gcd(k, P-1) = 1, находим k1 = k^(-1) mod (P - 1)
	gcd, k1, _ := common.GCDExtendedBig(k, new(big.Int).Sub(es.P, big.NewInt(1)))
	if gcd.Cmp(big.NewInt(1)) != 0 {
		return fmt.Errorf("gcd(%s, %s) != 1", k.String(), new(big.Int).Sub(es.P, big.NewInt(1)).String())
	}
	// Приводим k1 к модулю (P - 1)
	//k1.Mod(k1, new(big.Int).Sub(es.P, big.NewInt(1)))
	//if k1.Cmp(big.NewInt(0)) < 0 {
	//	k1.Add(k1, new(big.Int).Sub(es.P, big.NewInt(1)))
	//}
	// S = (u * k^(-1)) mod (P - 1)
	es.Signature = new(big.Int).Mul(u, k1)
	es.Signature.Mod(es.Signature, new(big.Int).Sub(es.P, big.NewInt(1)))
	fmt.Printf("Подпись создана: S = %s", es.Signature.String())
	es.Message = message
	return common.WriteNumbers(es.OutputSigned, []int64{es.Signature.Int64()})
}

func (es *elgamalSignature) Verify() (bool, error) {
	// Шаг 1: Вычисление хеша от сохраненного сообщения
	hash := md5.Sum(es.Message)
	hashInt := new(big.Int).SetBytes(hash[:])
	// Приводим хеш к модулю (P - 1), чтобы h < P - 1
	// hashInt = h mod (P - 1)
	hashInt.Mod(hashInt, new(big.Int).Sub(es.P, big.NewInt(1)))
	fmt.Printf("Message hash (as int): %s\n", hashInt.String())
	// Шаг 2: Вычисление значения yr = Y^R * R^S mod P
	// yr = Y^R * R^S mod P
	yr := new(big.Int).Mul(
		common.ModularExponentiationBig(es.Y, es.R, es.P),         // Y^R mod P
		common.ModularExponentiationBig(es.R, es.Signature, es.P), // R^S mod P
	)
	yr.Mod(yr, es.P) // Приводим к модулю P
	// Шаг 3: Вычисление g = G^h mod P
	// g = G^h mod P
	g := common.ModularExponentiationBig(es.G, hashInt, es.P)
	// Шаг 4: Сравнение yr и g
	// Подпись верна, если yr == g
	return yr.Cmp(g) == 0, nil
}

func (es *elgamalSignature) SignAndVerify() error {
	pwd, err := os.Getwd()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	fmt.Print("Signed: ")
	if err = es.Sign(); err != nil {
		return err
	}
	if _, ok := es.Input.(*os.File); ok {
		fmt.Print(pwd + "/" + es.OutputSigned.(*os.File).Name())
	}
	if out, ok := es.OutputSigned.(*os.File); ok {
		_ = out.Sync()
		defer out.Close()
		es.OutputSigned, err = os.OpenFile(out.Name(), os.O_RDONLY, 0600)
		if err != nil {
			return err
		}
	}
	fmt.Print("\nVerified: ")
	ok, err := es.Verify()
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
