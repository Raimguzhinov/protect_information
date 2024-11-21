package main

import (
	"fmt"
	"github.com/Raimguzhinov/protect-information/common"
	"github.com/samber/lo"
	"log"
	"math/big"
)

// Карта представляется числом от 2 до p-1
type Card struct {
	ID   *big.Int // Уникальный идентификатор карты
	Name string   // Название карты (например, "2♥")
}

func (c Card) String() string {
	return c.ID.String()
}

// Игрок
type Player struct {
	ID            int
	EncryptionKey *big.Int // c_i (взаимно простое с p)
	DecryptionKey *big.Int // d_i (обратное к c_i по модулю p)
}

// Генерация простого числа p
func generatePrime() *big.Int {
	minV := big.NewInt(1_000_000_000_000_000)
	maxV := big.NewInt(1_000_000_000_000_000_000)
	var p *big.Int
	//p := common.GenPrimeBig(minV, maxV)
	for {
		q := common.GenPrimeBig(minV, maxV)
		p = new(big.Int).Mul(q, big.NewInt(2))
		p.Add(p, big.NewInt(1)) // P = 2 * q + 1

		if common.IsPrimeBig(p) {
			break
		}
	}
	return p
}

// Генерация взаимно простого числа с p и его обратного элемента
func generateEncryptionKeys(p *big.Int) (*big.Int, *big.Int) {
	phiP := new(big.Int).Sub(p, big.NewInt(1)) // φ(p) = p - 1
	c := common.GenCoprimeBig(phiP, big.NewInt(2), phiP)
	d, err := common.ModInverseBig(c, phiP)
	if err != nil {
		log.Fatal(err)
	}
	if d.Cmp(phiP) >= 0 || c.Cmp(big.NewInt(1)) <= 0 {
		return nil, nil
	}
	return c, d
}

// Генерация колоды карт с уникальными случайными идентификаторами
func generateDeck() []Card {
	suits := []string{"♥", "♠", "♣", "♦"}
	values := []string{"2", "3", "4", "5", "6", "7", "8", "9", "10", "J", "Q", "K", "A"}
	deck := make([]Card, 0, 52)

	i := 2
	for _, value := range values {
		for _, suit := range suits {
			cardName := value + suit
			deck = append(deck, Card{
				ID:   big.NewInt(int64(i)),
				Name: cardName,
			})
			i++
		}
	}
	return deck
}

func encryptDeck(deck []Card, key *big.Int, p *big.Int) []Card {
	encryptedDeck := make([]Card, len(deck))
	for i, card := range deck {
		encryptedID := common.ModularExponentiationBig(card.ID, key, p)
		encryptedDeck[i] = Card{
			ID:   encryptedID,
			Name: card.Name,
		}
	}
	return encryptedDeck
}

func decryptDeck(deck []Card, key *big.Int, p *big.Int) []Card {
	decryptedDeck := make([]Card, len(deck))
	for i, card := range deck {
		decryptedID := common.ModularExponentiationBig(card.ID, key, p)
		decryptedDeck[i] = Card{
			ID:   decryptedID,
			Name: card.Name,
		}
	}
	return decryptedDeck
}

// Поиск названия карты по её идентификатору
func findCardByID(deck []Card, id *big.Int) string {
	for _, card := range deck {
		if card.ID.Cmp(id) == 0 {
			return card.Name
		}
	}
	return "Unknown"
}

func main() {
	numPlayers := 5
	p := generatePrime()
	fmt.Printf("p: %s\n", p.String())
	players := make([]Player, numPlayers)
	for i := 0; i < numPlayers; i++ {
		c, d := generateEncryptionKeys(p)
		players[i] = Player{
			ID:            i + 1,
			EncryptionKey: c,
			DecryptionKey: d,
		}
		fmt.Printf("Player %d: c = %s, d = %s\n", players[i].ID, c.String(), d.String())
	}

	originalDeck := generateDeck()
	fmt.Printf("Original Deck:\n%s\n\n", originalDeck)

	deck := make([]Card, len(originalDeck))
	copy(deck, originalDeck)

	for _, player := range players {
		deck = encryptDeck(deck, player.EncryptionKey, p)
		deck = lo.Shuffle(deck)
		fmt.Printf("Shuffled deck after encryption player %d:\n%s\n\n", player.ID, deck)
	}

	playerHands := make([][]Card, numPlayers)
	for i := 0; i < numPlayers; i++ {
		playerHands[i] = []Card{deck[0], deck[1]}
		deck = deck[2:]
	}
	tableCards := deck[:5]

	fmt.Println("\nКарты на столе:")
	fmt.Println(tableCards)
	for _, player := range players {
		tableCards = decryptDeck(tableCards, player.DecryptionKey, p)
	}
	for _, card := range tableCards {
		cardName := findCardByID(originalDeck, card.ID)
		fmt.Printf("%s ", cardName)
	}
	fmt.Println()

	for i := 0; i < numPlayers; i++ {
		fmt.Printf("\nКарты игрока %d:\n", players[i].ID)
		fmt.Println(playerHands[i])
		hand := playerHands[i]
		for j := len(players) - 1; j >= 0; j-- {
			if players[j].ID != players[i].ID {
				hand = decryptDeck(hand, players[j].DecryptionKey, p)
			}
		}
		hand = decryptDeck(hand, players[i].DecryptionKey, p)
		for _, card := range hand {
			cardName := findCardByID(originalDeck, card.ID)
			fmt.Printf("%s ", cardName)
		}
		fmt.Println()
	}
}
