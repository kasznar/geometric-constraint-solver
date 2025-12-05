package main

import (
	"fmt"
	"image/color"
	"log"
	"strconv"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

type Distance struct {
	P1, P2 *Point
	Value  float64
}

var distances []*Distance

type Game struct {
	inputRunes []rune
	inputText  string
	counter    int
	onSolve    func([]*Distance, []*Point, func([]*Point))
}

type Point struct {
	X, Y     int
	Selected bool
	index    int
}

var points []*Point

var lastMousePressed bool

var index int

func deselect() {
	for i := range points {
		points[i].Selected = false
	}
}

func (g *Game) Update() error {
	if ebiten.IsKeyPressed(ebiten.KeyR) {
		if len(points) > 2 {
			points = points[:2]
		}
		distances = distances[:0]
	}
	// Deselect all points if Escape is pressed
	if ebiten.IsKeyPressed(ebiten.KeyEscape) {
		deselect()
	}
	if ebiten.IsKeyPressed(ebiten.KeyQ) {
		return ebiten.Termination
	}
	// Delete selected points if D is pressed
	if ebiten.IsKeyPressed(ebiten.KeyD) {
		newPoints := make([]*Point, 0, len(points))
		for _, pt := range points {
			if !pt.Selected {
				newPoints = append(newPoints, pt)
			}
		}
		points = newPoints
	}
	// Detect mouse click
	mousePressed := ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft)
	if mousePressed && !lastMousePressed {
		x, y := ebiten.CursorPosition()
		screenWidth, screenHeight := ebiten.WindowSize()
		x0 := screenWidth / 2
		y0 := screenHeight / 2
		// Convert screen coordinates to centered coordinates
		coordX := x - x0
		coordY := y0 - y // y axis is inverted in screen coordinates
		log.Printf("Clicked at screen: (%d, %d), coordinate system: (%d, %d)", x, y, coordX, coordY)

		// Check if click is on 'SOLVE' text (bottom right corner)
		solveTextWidth := 50  // rough width in pixels
		solveTextHeight := 16 // rough height in pixels
		solveX := screenWidth - 60
		solveY := screenHeight - 20
		if x >= solveX && x <= solveX+solveTextWidth && y >= solveY && y <= solveY+solveTextHeight {
			if !lastMousePressed {
				g.onSolve(distances, points, func(p []*Point) {
					log.Printf("Solved points: %+v", p)
					points = p
				})
				log.Println("SOLVE button clicked!")
			}
			lastMousePressed = mousePressed

			// Do not add a new point if SOLVE is clicked
			return nil
		}

		// Check if click is near any point (in centered coordinates)
		selected := 0
		found := false
		for i, pt := range points {
			dx := pt.X - coordX
			dy := pt.Y - coordY
			if dx*dx+dy*dy <= 16 { // within radius 4
				if !points[i].Selected && selected < 2 {
					points[i].Selected = true
					selected++
				}
				found = true
			}
		}
		if !found {
			// Deselect all
			for i := range points {
				points[i].Selected = false
			}
			// If not near any point, add a new point (in centered coordinates)
			points = append(points, &Point{X: coordX, Y: coordY, index: index})
			index++
		} else {
			// If more than two are selected, deselect extras
			selected = 0
			for i := range points {
				if points[i].Selected {
					selected++
					if selected > 2 {
						points[i].Selected = false
					}
				}
			}
		}
	}
	lastMousePressed = mousePressed

	// Handle text input only if two points are selected
	selectedCount := 0
	for _, pt := range points {
		if pt.Selected {
			selectedCount++
		}
	}
	if selectedCount == 2 {
		// Add runes that are input by the user
		g.inputRunes = ebiten.AppendInputChars(g.inputRunes[:0])
		g.inputText += string(g.inputRunes)
		// If Enter is pressed, create Distance entity
		if ebiten.IsKeyPressed(ebiten.KeyEnter) && len(g.inputText) > 0 {
			var pts [2]*Point
			idx := 0
			for _, pt := range points {
				if pt.Selected && idx < 2 {
					pts[idx] = pt
					idx++
				}
			}
			if idx == 2 {
				// dx := float64(pts[0].X - pts[1].X)
				// dy := float64(pts[0].Y - pts[1].Y)
				// dist := math.Sqrt(dx*dx + dy*dy)
				if val, err := strconv.ParseFloat(g.inputText, 64); err == nil {
					distances = append(distances, &Distance{P1: pts[0], P2: pts[1], Value: val})
					deselect()
				}
			}
			g.inputText = ""
		}
		// If Backspace is pressed, remove last character
		if ebiten.IsKeyPressed(ebiten.KeyBackspace) && len(g.inputText) > 0 {
			g.inputText = g.inputText[:len(g.inputText)-1]
		}
	} else {
		g.inputText = ""
	}
	g.counter++
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	width := screen.Bounds().Dx()
	height := screen.Bounds().Dy()
	// Draw 'SOLVE' text button in bottom right corner
	solveText := "SOLVE"
	solveX := width - 60 // adjust for text width
	solveY := height - 20
	ebitenutil.DebugPrintAt(screen, solveText, solveX, solveY)

	// Show selected point in left bottom corner
	// width := screen.Bounds().Dx()
	// height := screen.Bounds().Dy()
	// Show up to two selected points in left bottom corner (centered coordinates)
	selectedLabels := []string{}
	for _, pt := range points {
		if pt.Selected {
			selectedLabels = append(selectedLabels, fmt.Sprintf("(%d, %d)", pt.X, pt.Y))
		}
	}
	if len(selectedLabels) > 0 {
		ebitenutil.DebugPrintAt(screen, "Selected: "+fmt.Sprintf("%s", selectedLabels), 8, height-20)
	}

	// Draw input box in top right if two points are selected
	selectedCount := 0
	for _, pt := range points {
		if pt.Selected {
			selectedCount++
		}
	}
	if selectedCount == 2 {
		inputDisplay := g.inputText
		if g.counter%60 < 30 {
			inputDisplay += "_"
		}
		ebitenutil.DebugPrintAt(screen, "Distance: "+inputDisplay, width-160, 8)
	}

	// var x0, y0 int
	// x0 = width / 2
	// y0 = height / 2
	// Draw clicked points as small circles (convert centered to screen coordinates)
	x0 := width / 2
	y0 := height / 2
	for _, pt := range points {
		col := color.RGBA{255, 0, 0, 255} // red
		if pt.Selected {
			col = color.RGBA{0, 0, 255, 255} // blue
		}
		sx := pt.X + x0
		sy := y0 - pt.Y
		vector.StrokeCircle(screen, float32(sx), float32(sy), 4, 2, col, false)
		// Show coordinates next to each point (centered)
		label := fmt.Sprintf("(%d, %d)", pt.X, pt.Y)
		ebitenutil.DebugPrintAt(screen, label, sx+8, sy-8)
	}

	// Draw x-axis
	vector.StrokeLine(screen, 0, float32(y0), float32(width), float32(y0), 1, color.White, false)
	// Draw y-axis
	vector.StrokeLine(screen, float32(x0), 0, float32(x0), float32(height), 1, color.White, false)

	// Draw ticks on x-axis
	for x := x0; x < width; x += 20 {
		vector.StrokeLine(screen, float32(x), float32(y0-5), float32(x), float32(y0+5), 1, color.White, false)
	}
	for x := x0; x > 0; x -= 20 {
		vector.StrokeLine(screen, float32(x), float32(y0-5), float32(x), float32(y0+5), 1, color.White, false)
	}

	// Draw ticks on y-axis
	for y := y0; y < height; y += 20 {
		vector.StrokeLine(screen, float32(x0-5), float32(y), float32(x0+5), float32(y), 1, color.White, false)
	}
	for y := y0; y > 0; y -= 20 {
		vector.StrokeLine(screen, float32(x0-5), float32(y), float32(x0+5), float32(y), 1, color.White, false)
	}

	// Draw axis labels
	ebitenutil.DebugPrintAt(screen, "X", width-20, y0+5)
	ebitenutil.DebugPrintAt(screen, "Y", x0+5, 5)
	ebitenutil.DebugPrintAt(screen, "0", x0+5, y0+5)

	ebitenutil.DebugPrint(screen, "Constraint Solver")

	// Draw all distances (convert centered to screen coordinates)
	for _, d := range distances {
		sx1 := d.P1.X + x0
		sy1 := y0 - d.P1.Y
		sx2 := d.P2.X + x0
		sy2 := y0 - d.P2.Y
		vector.StrokeLine(screen, float32(sx1), float32(sy1), float32(sx2), float32(sy2), 2, color.RGBA{0, 255, 0, 255}, false)
		mx := (sx1 + sx2) / 2
		my := (sy1 + sy2) / 2
		ebitenutil.DebugPrintAt(screen, fmt.Sprintf("%.2f", d.Value), mx, my)
	}
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return 640, 480
}


func LaunchUI(onSolve func([]*Distance, []*Point, func([]*Point))) {
	g := &Game{}
	g.onSolve = onSolve
	points = append(points, &Point{100, 0, false, index})
	index++                                              
	points = append(points, &Point{0, 100, false, index}) 
	index++

	ebiten.SetWindowSize(640, 480)
	ebiten.SetWindowTitle("\"CAD\"")
	if err := ebiten.RunGame(g); err != nil {
		log.Fatal(err)
	}
}
