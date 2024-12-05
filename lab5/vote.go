package main

import (
	"crypto/sha512"
	"fmt"
	"github.com/Raimguzhinov/protect-information/common"
	"log"
	"math/big"
	"sync"
)

// Тип для представления варианта голоса
type Vote int

const (
	YES Vote = iota + 1
	NO
	ABSTAIN
)

// Структура сервера
type Server struct {
	n     *big.Int        // Модуль RSA
	d     *big.Int        // Публичный экспонент
	c     *big.Int        // Приватный экспонент
	voted map[string]bool // Отслеживание проголосовавших пользователей
	votes map[Vote]int    // Результаты голосования
	mu    sync.Mutex      // Мьютекс для синхронизации
}

// Создание нового сервера
func NewServer() *Server {
	// Генерируем простые числа p и q
	bitSize := 1024
	minV := new(big.Int).Lsh(big.NewInt(1), uint(bitSize-1))
	maxV := new(big.Int).Lsh(big.NewInt(1), uint(bitSize))

	p := common.GenPrimeBig(minV, maxV)
	q := common.GenPrimeBig(minV, maxV)
	for p.Cmp(q) == 0 {
		q = common.GenPrimeBig(minV, maxV)
	}

	// Вычисляем n = p * q и φ(n) = (p - 1) * (q - 1)
	n := new(big.Int).Mul(p, q)
	phi := new(big.Int).Mul(new(big.Int).Sub(p, big.NewInt(1)), new(big.Int).Sub(q, big.NewInt(1)))

	// Выбираем d, взаимно простое с φ(n)
	d := common.GenCoprimeBig(phi, big.NewInt(3), phi)

	// Вычисляем c, обратное к d по модулю φ(n)
	c, err := common.ModInverseBig(d, phi)
	if err != nil {
		log.Fatalf("Ошибка при вычислении c: %v", err)
	}

	return &Server{
		n:     n,
		d:     d,
		c:     c,
		voted: make(map[string]bool),
		votes: map[Vote]int{YES: 0, NO: 0, ABSTAIN: 0},
	}
}

// Метод сервера для подписи слепого сообщения
func (s *Server) GetBlindSignature(username string, blindedHash *big.Int) *big.Int {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.voted[username] {
		fmt.Printf("Сервер: Пользователь %s уже голосовал\n", username)
		return nil
	}

	// Отмечаем, что пользователь проголосовал
	s.voted[username] = true

	// Подписываем слепое сообщение
	signature := common.ModularExponentiationBig(blindedHash, s.c, s.n)
	return signature
}

// Метод сервера для проверки и учёта голоса
func (s *Server) SubmitVote(voteValue *big.Int, signature *big.Int) bool {
	// Вычисляем хэш от голоса
	hash := sha512.Sum512(voteValue.Bytes())
	hashInt := new(big.Int).SetBytes(hash[:])

	// Проверяем подпись
	expectedHash := common.ModularExponentiationBig(signature, s.d, s.n)

	if hashInt.Cmp(expectedHash) == 0 {
		// Извлекаем значение голоса (1 - YES, 2 - NO, 3 - ABSTAIN)
		voteInt := new(big.Int).And(voteValue, big.NewInt(3)).Int64()
		s.mu.Lock()
		s.votes[Vote(voteInt)]++
		s.mu.Unlock()
		fmt.Println("Сервер: Голос принят")
		return true
	}

	fmt.Println("Сервер: Голос отклонён: некорректная подпись")
	return false
}

// Метод для отображения результатов голосования
func (s *Server) ShowResults() {
	fmt.Println("\nСервер: Результаты голосования:")
	for vote, count := range s.votes {
		var voteStr string
		switch vote {
		case YES:
			voteStr = "Да"
		case NO:
			voteStr = "Нет"
		case ABSTAIN:
			voteStr = "Воздержался"
		}
		fmt.Printf("%s: %d\n", voteStr, count)
	}
}

// Структура клиента
type Client struct {
	server   *Server
	username string
}

// Создание нового клиента
func NewClient(server *Server, username string) *Client {
	return &Client{
		server:   server,
		username: username,
	}
}

// Метод клиента для голосования
func (c *Client) Vote(vote Vote) {
	// Генерируем случайное число r, взаимно простое с n
	r := common.GenCoprimeBig(c.server.n, big.NewInt(2), c.server.n)

	// Формируем сообщение m (голос)
	minV := big.NewInt(1_000_000_000_000_000)
	maxV := big.NewInt(1_000_000_000_000_000_000)
	randomPadding := common.GenPrimeBig(minV, maxV)
	m := new(big.Int).Lsh(randomPadding, 2)
	m = new(big.Int).Or(m, big.NewInt(int64(vote)))

	// Вычисляем хэш от m
	hash := sha512.Sum512(m.Bytes())
	hashInt := new(big.Int).SetBytes(hash[:])

	// Вычисляем слепое сообщение
	rExpE := common.ModularExponentiationBig(r, c.server.d, c.server.n)
	blindedHash := new(big.Int).Mul(hashInt, rExpE)
	//blindedHash.Mod(blindedHash, c.server.n)

	// Получаем слепую подпись от сервера
	blindSignature := c.server.GetBlindSignature(c.username, blindedHash)
	if blindSignature == nil {
		return
	}

	// Снимаем слепоту с подписи
	rInv, err := common.ModInverseBig(r, c.server.n)
	if err != nil {
		log.Fatalf("Ошибка при вычислении обратного к r: %v", err)
	}
	signature := new(big.Int).Mul(blindSignature, rInv)
	//signature.Mod(signature, c.server.n)

	// Отправляем голос и подпись на сервер
	if c.server.SubmitVote(m, signature) {
		fmt.Printf("Клиент: Голос пользователя %s принят\n", c.username)
	} else {
		fmt.Printf("Клиент: Голос пользователя %s отклонён\n", c.username)
	}
}

// Главная функция
func main() {
	// Создаём сервер
	server := NewServer()
	fmt.Printf("Параметры сервера:\n n = %s\n d = %s\n c = %s\n", server.n.String(), server.d.String(), server.c.String())

	// Создаём клиентов
	alice := NewClient(server, "Alice")
	bob := NewClient(server, "Bob")
	charlie := NewClient(server, "Charlie")

	// Клиенты голосуют
	alice.Vote(YES)
	bob.Vote(NO)
	charlie.Vote(ABSTAIN)

	// Попытка повторного голосования
	alice.Vote(NO)

	// Отображаем результаты
	server.ShowResults()
}
