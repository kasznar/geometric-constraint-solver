package main

import (
	"fmt"
	"image/color"
	"log"
	"strconv"
	"strings"

	"equation-solver/pkg/solver"

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

	inspEnabled bool
	inspector   *solver.Inspector
	// pendingPoints receives final solved point positions from the solver
	// goroutine; drained safely on the UI goroutine in Update.
	pendingPoints chan []*Point

	onSolve func(
		distances []*Distance,
		pts []*Point,
		inspEnabled bool,
		update func([]*Point),
		setInspector func(*solver.Inspector),
	)
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

// prevKeys tracks key state for edge detection.
var prevKeys = map[ebiten.Key]bool{}

func justPressed(k ebiten.Key) bool {
	pressed := ebiten.IsKeyPressed(k)
	was := prevKeys[k]
	prevKeys[k] = pressed
	return pressed && !was
}

func (g *Game) Update() error {
	if ebiten.IsKeyPressed(ebiten.KeyQ) {
		return ebiten.Termination
	}

	// Apply any pending point updates from the solver goroutine.
	select {
	case p := <-g.pendingPoints:
		points = p
	default:
	}

	// Toggle inspection mode with I key (only when not currently solving).
	if g.inspector == nil && justPressed(ebiten.KeyI) {
		g.inspEnabled = !g.inspEnabled
	}

	// STEP / RELEASE keys when inspector is active and paused.
	if g.inspector != nil {
		if g.inspector.IsPaused() {
			if justPressed(ebiten.KeySpace) || justPressed(ebiten.KeyArrowRight) {
				g.inspector.Step()
			}
			if justPressed(ebiten.KeyR) {
				g.inspector.Release()
			}
		}
		// Clean up once the solve goroutine is done.
		select {
		case <-g.inspector.Done:
			g.inspector = nil
		default:
		}
		return nil
	}

	if ebiten.IsKeyPressed(ebiten.KeyR) {
		if len(points) > 2 {
			points = points[:2]
		}
		distances = distances[:0]
	}
	if ebiten.IsKeyPressed(ebiten.KeyEscape) {
		deselect()
	}
	if ebiten.IsKeyPressed(ebiten.KeyD) || (justPressed(ebiten.KeyBackspace) && len(g.inputText) == 0) {
		newPoints := make([]*Point, 0, len(points))
		for _, pt := range points {
			if !pt.Selected {
				newPoints = append(newPoints, pt)
			}
		}
		points = newPoints
	}

	mousePressed := ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft)
	if mousePressed && !lastMousePressed {
		x, y := ebiten.CursorPosition()
		screenWidth, screenHeight := ebiten.WindowSize()
		x0 := screenWidth / 2
		y0 := screenHeight / 2
		coordX := x - x0
		coordY := y0 - y
		log.Printf("Clicked at screen: (%d, %d), coordinate system: (%d, %d)", x, y, coordX, coordY)

		// SOLVE button
		solveX := screenWidth - 60
		solveY := screenHeight - 20
		if x >= solveX && x <= solveX+50 && y >= solveY && y <= solveY+16 {
			if !lastMousePressed {
				g.onSolve(
					distances, points, g.inspEnabled,
					func(p []*Point) {
						select {
						case g.pendingPoints <- p:
						default:
						}
					},
					func(ins *solver.Inspector) { g.inspector = ins },
				)
				log.Println("SOLVE button clicked!")
			}
			lastMousePressed = mousePressed
			return nil
		}

		// INSPECT toggle button
		inspX := screenWidth - 130
		inspY := screenHeight - 20
		if x >= inspX && x <= inspX+65 && y >= inspY && y <= inspY+16 {
			if !lastMousePressed {
				g.inspEnabled = !g.inspEnabled
			}
			lastMousePressed = mousePressed
			return nil
		}

		// Point selection / creation
		selected := 0
		found := false
		for i, pt := range points {
			dx := pt.X - coordX
			dy := pt.Y - coordY
			if dx*dx+dy*dy <= 144 {
				if !points[i].Selected && selected < 2 {
					points[i].Selected = true
					selected++
				}
				found = true
			}
		}
		if !found {
			for i := range points {
				points[i].Selected = false
			}
			points = append(points, &Point{X: coordX, Y: coordY, index: index})
			index++
		} else {
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

	selectedCount := 0
	for _, pt := range points {
		if pt.Selected {
			selectedCount++
		}
	}
	if selectedCount == 2 {
		g.inputRunes = ebiten.AppendInputChars(g.inputRunes[:0])
		g.inputText += string(g.inputRunes)
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
				if val, err := strconv.ParseFloat(g.inputText, 64); err == nil {
					distances = append(distances, &Distance{P1: pts[0], P2: pts[1], Value: val})
					deselect()
				}
			}
			g.inputText = ""
		}
		if ebiten.IsKeyPressed(ebiten.KeyBackspace) && len(g.inputText) > 0 {
			g.inputText = g.inputText[:len(g.inputText)-1]
		}
	} else {
		g.inputText = ""
	}
	g.counter++
	return nil
}

const canvasInspectW = 640 // geometry canvas width when inspector panel is open

func (g *Game) Draw(screen *ebiten.Image) {
	width := screen.Bounds().Dx()
	height := screen.Bounds().Dy()

	// Constrain geometry canvas to the left when inspector panel is active.
	canvasW := width
	if g.inspector != nil {
		canvasW = canvasInspectW
	}

	x0 := canvasW / 2
	y0 := height / 2

	pointX := func(pt *Point) int { return pt.X }
	pointY := func(pt *Point) int { return pt.Y }

	// Axes (clipped to canvas width)
	vector.StrokeLine(screen, 0, float32(y0), float32(canvasW), float32(y0), 1, color.White, false)
	vector.StrokeLine(screen, float32(x0), 0, float32(x0), float32(height), 1, color.White, false)
	for x := x0; x < canvasW; x += 20 {
		vector.StrokeLine(screen, float32(x), float32(y0-5), float32(x), float32(y0+5), 1, color.White, false)
	}
	for x := x0; x > 0; x -= 20 {
		vector.StrokeLine(screen, float32(x), float32(y0-5), float32(x), float32(y0+5), 1, color.White, false)
	}
	for y := y0; y < height; y += 20 {
		vector.StrokeLine(screen, float32(x0-5), float32(y), float32(x0+5), float32(y), 1, color.White, false)
	}
	for y := y0; y > 0; y -= 20 {
		vector.StrokeLine(screen, float32(x0-5), float32(y), float32(x0+5), float32(y), 1, color.White, false)
	}
	ebitenutil.DebugPrintAt(screen, "X", canvasW-20, y0+5)
	ebitenutil.DebugPrintAt(screen, "Y", x0+5, 5)
	ebitenutil.DebugPrintAt(screen, "0", x0+5, y0+5)
	ebitenutil.DebugPrint(screen, "Constraint Solver")

	// Constraints
	for _, d := range distances {
		sx1 := pointX(d.P1) + x0
		sy1 := y0 - pointY(d.P1)
		sx2 := pointX(d.P2) + x0
		sy2 := y0 - pointY(d.P2)
		vector.StrokeLine(screen, float32(sx1), float32(sy1), float32(sx2), float32(sy2), 2, color.RGBA{0, 255, 0, 255}, false)
		mx := (sx1 + sx2) / 2
		my := (sy1 + sy2) / 2
		ebitenutil.DebugPrintAt(screen, fmt.Sprintf("%.2f", d.Value), mx, my)
	}

	// Points
	for _, pt := range points {
		col := color.RGBA{255, 0, 0, 255}
		if pt.Selected {
			col = color.RGBA{0, 0, 255, 255}
		}
		sx := pointX(pt) + x0
		sy := y0 - pointY(pt)
		vector.StrokeCircle(screen, float32(sx), float32(sy), 4, 2, col, false)
		ebitenutil.DebugPrintAt(screen, fmt.Sprintf("(%d, %d)", pointX(pt), pointY(pt)), sx+8, sy-8)
	}

	// Inspector overlay when active
	if g.inspector != nil {
		g.drawInspectorOverlay(screen, width, height)
		return
	}

	// Normal mode bottom bar
	ebitenutil.DebugPrintAt(screen, "SOLVE", width-60, height-20)

	inspLabel := "INSPECT:OFF"
	if g.inspEnabled {
		inspLabel = "INSPECT:ON"
	}
	ebitenutil.DebugPrintAt(screen, inspLabel, width-130, height-20)

	selectedLabels := []string{}
	for _, pt := range points {
		if pt.Selected {
			selectedLabels = append(selectedLabels, fmt.Sprintf("(%d, %d)", pt.X, pt.Y))
		}
	}
	if len(selectedLabels) > 0 {
		ebitenutil.DebugPrintAt(screen, "Selected: "+strings.Join(selectedLabels, ", "), 8, height-20)
	}

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
}

func (g *Game) drawInspectorOverlay(screen *ebiten.Image, width, height int) {
	// Vertical divider between geometry canvas and data panel.
	vector.StrokeLine(screen, float32(canvasInspectW), 0, float32(canvasInspectW), float32(height), 1, color.RGBA{80, 80, 80, 255}, false)

	px := canvasInspectW + 10
	py := 8

	iter, paramNames, J, F := g.inspector.DisplayData()

	var statusLine string
	if g.inspector.IsPaused() {
		statusLine = fmt.Sprintf("Iteration: %d    SPACE/-> step   R release", iter+1)
	} else {
		statusLine = fmt.Sprintf("Iteration: %d    running...", iter+1)
	}
	ebitenutil.DebugPrintAt(screen, "Newton Inspector", px, py)
	py += 16
	ebitenutil.DebugPrintAt(screen, statusLine, px, py)
	py += 20

	// F(x) residuals
	ebitenutil.DebugPrintAt(screen, "F(x) residuals:", px, py)
	py += 14
	for i, v := range F {
		ebitenutil.DebugPrintAt(screen, fmt.Sprintf("  f%d:  %+.6f", i, v), px, py)
		py += 13
	}
	py += 6

	// J(x) Jacobian
	ebitenutil.DebugPrintAt(screen, "J(x) Jacobian:", px, py)
	py += 14

	// Column headers (param names)
	colW := 80
	headerX := px + 36
	for k, name := range paramNames {
		ebitenutil.DebugPrintAt(screen, name, headerX+k*colW, py)
	}
	py += 13

	// Rows
	for i, row := range J {
		ebitenutil.DebugPrintAt(screen, fmt.Sprintf("f%d:", i), px, py)
		for k, v := range row {
			ebitenutil.DebugPrintAt(screen, fmt.Sprintf("%+8.3f", v), headerX+k*colW, py)
		}
		py += 13
	}
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return 1200, 900
}

func LaunchUI(onSolve func(
	distances []*Distance,
	pts []*Point,
	inspEnabled bool,
	update func([]*Point),
	setInspector func(*solver.Inspector),
)) {
	g := &Game{
		pendingPoints: make(chan []*Point, 1),
	}
	g.onSolve = onSolve
	points = append(points, &Point{100, 0, false, index})
	index++
	points = append(points, &Point{0, 100, false, index})
	index++

	ebiten.SetWindowSize(1200, 900)
	ebiten.SetWindowTitle("\"CAD\"")
	if err := ebiten.RunGame(g); err != nil {
		log.Fatal(err)
	}
}
