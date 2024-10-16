package common

import (
	"fmt"
	"log"
)

// DiffieHellman - функция вычисления ключа шифрования Diffie-Hellman
func DiffieHellman() (int64, error) {
	minV, maxV := 1_000_000, 1_000_000_000
	q := GenPrime(int64(minV), int64(maxV))
	P := 2*q + 1
	g := int64(0)
	for i := int64(2); i < P-1; i++ {
		g = i
		if ModularExponentiation(g, q, P) != 1 {
			break
		}
	}
	log.Printf("P = %d, g = %d", P, g)
	Xa := Seed().Int63n(P-1) + 1 // private Alice key
	Xb := Seed().Int63n(P-1) + 1 // private Bob key
	log.Printf("Xa = %d, Xb = %d", Xa, Xb)
	Ya := ModularExponentiation(g, Xa, P) // public Alice key
	Yb := ModularExponentiation(g, Xb, P) // public Bob key
	log.Printf("Ya = %d, Yb = %d", Ya, Yb)
	Zab := ModularExponentiation(Yb, Xa, P)
	Zba := ModularExponentiation(Ya, Xb, P)
	if Zab != Zba {
		return -1, fmt.Errorf("z_ab must be equal to z_ba")
	}
	return Zab, nil
}
