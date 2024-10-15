package common

import (
	"fmt"
	"log"
	"math/rand"
	"runtime"
	"sync"
	"time"
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

func GenPrime(minV int64, maxV int64) int64 {
	for {
		num := Seed().Int63n(maxV-minV) + minV
		if IsPrime(num) {
			return num
		}
	}
}

// IsPrime - функция проверки числа на простоту (алгоритм Ферма)
func IsPrime(p int64) bool {
	if p <= 1 {
		return false
	} else if p <= 3 {
		return true
	} else if p%2 == 0 || p%3 == 0 {
		return false
	}
	var wg sync.WaitGroup
	wg.Add(runtime.NumCPU())
	doneCh := make(chan bool, runtime.NumCPU())
	for i := 0; i < runtime.NumCPU(); i++ {
		go func() {
			defer wg.Done()
			a := Seed().Int63n(p-2) + 2 // [2, p-2]
			if ModularExponentiation(a, p-1, p) == 1 {
				doneCh <- true
			}
		}()
	}
	wg.Wait()
	close(doneCh)
	for v := range doneCh {
		if v {
			return true
		}
	}
	return false
}

func Seed() *rand.Rand {
	return rand.New(rand.NewSource(time.Now().UnixNano()))
}
