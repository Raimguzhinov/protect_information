package main

import (
	"log"
	"protect_information/common"
)

func main() {
	minV, maxV := int64(1_000_000), int64(1_000_000_00)
	p := common.GenPrime(minV, maxV) //common.Seed().Int63n(maxV-minV) + minV
	a := common.Seed().Int63n(maxV-minV) + minV
	x := common.Seed().Int63n(p-2-1) + 1 // 1, p-2
	y := common.ModularExponentiation(a, x, p)
	log.Println(y)
	var xVzlom int64
	var err error
	for {
		xVzlom, err = common.GiantBabyStep(a, p, y)
		if err != nil {
			log.Fatal(err)
		}
		if xVzlom == x {
			break
		}
	}
	log.Println(xVzlom == x)
}
