package main

import (
	"bufio"
	"fmt"
	"github.com/Raimguzhinov/protect-information/common"
	"github.com/samber/lo"
	"log"
	"math/big"
	"math/rand"
	"os"
	"strings"
)

type Edge struct {
	From  int
	To    int
	Index int
}

// Чтение графа из файла
func readGraph(filename string) ([]Edge, []string, int, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, nil, 0, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	// Считываем количество вершин и рёбер
	scanner.Scan()
	var vertexNum, edgeNum int
	_, err = fmt.Sscanf(scanner.Text(), "%d %d", &vertexNum, &edgeNum)
	if err != nil {
		return nil, nil, 0, err
	}

	edges := make([]Edge, 0, edgeNum)

	// Считываем рёбра
	for i := 0; i < edgeNum; i++ {
		scanner.Scan()
		var from, to int
		_, err = fmt.Sscanf(scanner.Text(), "%d %d", &from, &to)
		if err != nil {
			return nil, nil, 0, err
		}
		edges = append(edges, Edge{From: from, To: to, Index: i + 1})
	}

	// Считываем цвета вершин
	scanner.Scan()
	colors := strings.Fields(scanner.Text())

	return edges, colors, vertexNum, nil
}

// shuffleUntilDifferent гарантирует, что перемешанный срез отличается от оригинального
func shuffleUntilDifferent(colors []string) []string {
	shuffled := make([]string, len(colors))
	copy(shuffled, colors)

	for {
		rand.Shuffle(len(shuffled), func(i, j int) {
			shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
		})

		if !equalSlices(colors, shuffled) {
			break
		}
	}
	return shuffled
}

// equalSlices сравнивает два среза на равенство
func equalSlices(slice1, slice2 []string) bool {
	if len(slice1) != len(slice2) {
		return false
	}
	for i := range slice1 {
		if slice1[i] != slice2[i] {
			return false
		}
	}
	return true
}

func main() {
	filename := "correct_graph.txt"

	// Чтение графа из файла
	edges, colors, vertexNum, err := readGraph(filename)
	if err != nil {
		log.Fatalf("Ошибка чтения графа: %v", err)
	}

	fmt.Printf("Граф содержит %d вершин и %d рёбер:\n", vertexNum, len(edges))
	for _, edge := range edges {
		fmt.Printf("%d %d (ребро %d)\n", edge.From, edge.To, edge.Index)
	}

	fmt.Println("Исходная раскраска:", strings.Join(colors, " "))

	// Перекрашивание
	everyColorIsValid := lo.EveryBy(colors, func(item string) bool {
		return lo.Contains([]string{"R", "B", "Y"}, item)
	})
	if !everyColorIsValid {
		log.Fatalln("Неизвестный цвет в раскраске графа")
	}
	uniqueColors := lo.Uniq(colors)
	shuffledColors := shuffleUntilDifferent(uniqueColors)
	colorMapping := make(map[string]string)
	for i, color := range uniqueColors {
		colorMapping[color] = shuffledColors[i]
	}
	recoloredColors := lo.Map(colors, func(color string, _ int) string {
		return colorMapping[color]
	})
	fmt.Println("Перекрашенная раскраска:", strings.Join(recoloredColors, " "))

	// Генерация криптографических параметров
	r := make([]*big.Int, vertexNum)
	p := make([]*big.Int, vertexNum)
	q := make([]*big.Int, vertexNum)
	n := make([]*big.Int, vertexNum)
	phi := make([]*big.Int, vertexNum)
	d := make([]*big.Int, vertexNum)
	c := make([]*big.Int, vertexNum)
	Z := make([]*big.Int, vertexNum)

	for i := 0; i < vertexNum; i++ {
		// Генерируем простые числа p и q
		p[i] = common.GenPrimeBig(big.NewInt(32500), big.NewInt(45000))
		q[i] = common.GenPrimeBig(big.NewInt(32500), big.NewInt(45000))

		// Вычисляем n = p * q и φ(n) = (p - 1) * (q - 1)
		n[i] = new(big.Int).Mul(p[i], q[i])
		phi[i] = new(big.Int).Mul(new(big.Int).Sub(p[i], big.NewInt(1)), new(big.Int).Sub(q[i], big.NewInt(1)))

		// Генерируем взаимно простое число d
		d[i] = common.GenCoprimeBig(phi[i], big.NewInt(2), phi[i])

		// Вычисляем обратное число c = d^-1 mod φ(n)
		c[i], _ = common.ModInverseBig(d[i], phi[i])

		// Генерируем случайное число r
		r[i] = common.GenCoprimeBig(n[i], big.NewInt(1), n[i])

		// Вычисляем Z = r^d mod n
		Z[i] = common.ModularExponentiationBig(r[i], d[i], n[i])
	}

	// Проверка корректности раскраски
	flag := false
	for _, edge := range edges {
		u := edge.From - 1
		v := edge.To - 1

		// Вычисляем Z1 и Z2 для вершин u и v
		Z1 := common.ModularExponentiationBig(Z[u], c[u], n[u])
		Z2 := common.ModularExponentiationBig(Z[v], c[v], n[v])

		// Сравниваем младшие 2 бита
		if Z1.Bit(0) != Z2.Bit(0) || Z1.Bit(1) != Z2.Bit(1) {
			fmt.Printf("Для ребра %d два младших бита различны.\n", edge.Index)
		} else {
			flag = true
			fmt.Printf("Ошибка! Два последних бита совпадают у ребра %d.\n", edge.Index)
		}
	}

	if flag {
		fmt.Println("Ошибка в раскраске графа!")
	} else {
		fmt.Println("Граф раскрашен корректно!")
	}
}
