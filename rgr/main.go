package main

import (
	"bufio"
	"fmt"
	"image/color"
	"log"
	"math"
	"math/big"
	"math/rand"
	"os"
	"strconv"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"github.com/Raimguzhinov/protect-information/common"
	"github.com/samber/lo"
)

type Edge struct {
	From, To int
}

type Graph struct {
	Vertices int
	Edges    []Edge
	Colors   []string
}

func readGraphFromFile(filename string) (*Graph, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			log.Fatalln(err)
		}
	}(file)

	scanner := bufio.NewScanner(file)

	scanner.Scan()
	line := scanner.Text()
	parts := strings.Fields(line)
	vertices, err := strconv.Atoi(parts[0])
	if err != nil {
		return nil, err
	}
	edgesCount, err := strconv.Atoi(parts[1])
	if err != nil {
		return nil, err
	}

	edges := make([]Edge, edgesCount)
	for i := 0; i < edgesCount; i++ {
		scanner.Scan()
		line := scanner.Text()
		parts := strings.Fields(line)
		from, err := strconv.Atoi(parts[0])
		if err != nil {
			return nil, err
		}
		to, err := strconv.Atoi(parts[1])
		if err != nil {
			return nil, err
		}
		edges[i] = Edge{From: from, To: to}
	}

	scanner.Scan()
	colors := strings.Fields(scanner.Text())

	return &Graph{
		Vertices: vertices,
		Edges:    edges,
		Colors:   colors,
	}, nil
}

func visualizeGraph(graph *Graph) *fyne.Container {
	width, height := 800.0, 600.0
	centerX, centerY := width/2, height/2
	radius := 200.0

	vertexCoords := make(map[int]fyne.Position)
	angleStep := 2 * math.Pi / float64(graph.Vertices)
	for i := 0; i < graph.Vertices; i++ {
		angle := angleStep * float64(i)
		x := centerX + radius*math.Cos(angle)
		y := centerY + radius*math.Sin(angle)
		vertexCoords[i+1] = fyne.NewPos(float32(x), float32(y))
	}

	lines := []fyne.CanvasObject{}
	for _, edge := range graph.Edges {
		from, to := vertexCoords[edge.From], vertexCoords[edge.To]
		line := canvas.NewLine(color.White)
		line.Position1, line.Position2 = from, to
		line.StrokeWidth = 2
		lines = append(lines, line)
	}

	circles := []fyne.CanvasObject{}
	for i := 0; i < graph.Vertices; i++ {
		pos := vertexCoords[i+1]
		col := parseColor(graph.Colors[i])
		circle := canvas.NewCircle(col)
		circle.Move(fyne.NewPos(pos.X-15, pos.Y-15))
		circle.Resize(fyne.NewSize(30, 30))
		circles = append(circles, circle)
	}

	labels := []fyne.CanvasObject{}
	for i := 0; i < graph.Vertices; i++ {
		pos := vertexCoords[i+1]
		label := canvas.NewText(fmt.Sprintf("%d", i+1), color.White)
		label.Alignment = fyne.TextAlignCenter
		label.Move(fyne.NewPos(pos.X-10, pos.Y-30))
		labels = append(labels, label)
	}

	allObjects := append(lines, circles...)
	allObjects = append(allObjects, labels...)
	return container.NewWithoutLayout(allObjects...)
}

func checkGraphColoring(graph *Graph) string {
	fmt.Printf("Граф содержит %d вершин и %d рёбер:\n", graph.Vertices, len(graph.Edges))
	for i, edge := range graph.Edges {
		fmt.Printf("%d %d (ребро %d)\n", edge.From, edge.To, i+1)
	}
	fmt.Println("Исходная раскраска:", strings.Join(graph.Colors, " "))

	// Перекрашивание
	everyColorIsValid := lo.EveryBy(graph.Colors, func(item string) bool {
		return lo.Contains([]string{"R", "B", "Y"}, item)
	})
	if !everyColorIsValid {
		return fmt.Sprintln("Неизвестный цвет в раскраске графа")
	}
	uniqueColors := lo.Uniq(graph.Colors)
	shuffledColors := shuffleUntilDifferent(uniqueColors)
	colorMapping := make(map[string]string)
	for i, iColor := range uniqueColors {
		colorMapping[iColor] = shuffledColors[i]
	}
	recoloredColors := lo.Map(graph.Colors, func(iColor string, _ int) string {
		return colorMapping[iColor]
	})
	fmt.Println("Перекрашенная раскраска:", strings.Join(recoloredColors, " "))

	// Генерация криптографических параметров
	r := make([]*big.Int, graph.Vertices)
	p := make([]*big.Int, graph.Vertices)
	q := make([]*big.Int, graph.Vertices)
	n := make([]*big.Int, graph.Vertices)
	phi := make([]*big.Int, graph.Vertices)
	d := make([]*big.Int, graph.Vertices)
	c := make([]*big.Int, graph.Vertices)
	Z := make([]*big.Int, graph.Vertices)

	for i := 0; i < graph.Vertices; i++ {
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
		// Модифицируем r по цвету
		r[i] = modifyRByColor(r[i], recoloredColors[i])
		// Вычисляем Z = r^d mod n
		Z[i] = common.ModularExponentiationBig(r[i], d[i], n[i])
	}

	for _, edge := range graph.Edges {
		u, v := edge.From-1, edge.To-1
		Z1 := common.ModularExponentiationBig(Z[u], c[u], n[u])
		Z2 := common.ModularExponentiationBig(Z[v], c[v], n[v])
		mask := big.NewInt(3)
		Z1LowerBits := new(big.Int).And(Z1, mask)
		Z2LowerBits := new(big.Int).And(Z2, mask)
		if Z1LowerBits.Cmp(Z2LowerBits) == 0 {
			return fmt.Sprintf("Ошибка: вершины %d и %d имеют одинаковые младшие биты!", edge.From, edge.To)
		}
	}
	return "Граф раскрашен корректно!"
}

func parseColor(iColor string) color.Color {
	switch strings.ToUpper(iColor) {
	case "R":
		return color.NRGBA{R: 255, G: 0, B: 0, A: 255}
	case "B":
		return color.NRGBA{R: 0, G: 0, B: 255, A: 255}
	case "Y":
		return color.NRGBA{R: 255, G: 255, B: 0, A: 255}
	default:
		return color.Black
	}
}

func modifyRByColor(r *big.Int, color string) *big.Int {
	r = new(big.Int).And(r, new(big.Int).Not(big.NewInt(3)))
	switch color {
	case "R":
		fmt.Printf("[R]: ")
	case "B":
		fmt.Printf("[B]: ")
		r = new(big.Int).Or(r, big.NewInt(1))
	case "Y":
		fmt.Printf("[Y]: ")
		r = new(big.Int).Or(r, big.NewInt(2))
	default:
		log.Fatalf("Неизвестный цвет: %s", color)
	}
	fmt.Printf("r (after modification) = %d%d\n", r.Bit(0), r.Bit(1))
	return r
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
	a := app.New()
	w := a.NewWindow("Graph Coloring Visualizer")
	w.Resize(fyne.NewSize(900, 700))

	graphContainer := container.NewVBox()
	var currentGraph *Graph

	chooseFileButton := widget.NewButton("Выбрать файл", func() {
		dialog.NewFileOpen(func(file fyne.URIReadCloser, err error) {
			if file == nil {
				return
			}
			graph, err := readGraphFromFile(file.URI().Path())
			if err != nil {
				dialog.ShowError(err, w)
				return
			}
			currentGraph = graph
			graphContainer.Objects = []fyne.CanvasObject{visualizeGraph(graph)}
			graphContainer.Refresh()
		}, w).Show()
	})

	checkButton := widget.NewButton("Проверить раскраску", func() {
		if currentGraph == nil {
			dialog.ShowInformation("Ошибка", "Граф не загружен.", w)
			return
		}
		result := checkGraphColoring(currentGraph)
		dialog.ShowInformation("Результат проверки", result, w)
	})

	w.SetContent(container.NewBorder(
		container.NewHBox(chooseFileButton, checkButton), nil, nil, nil, graphContainer,
	))

	w.ShowAndRun()
}
