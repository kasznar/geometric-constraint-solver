package main

import (
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"log"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// ---- JSON snapshot types ----

type snapPoint struct {
	X     float64 `json:"x"`
	Y     float64 `json:"y"`
	Z     float64 `json:"z"`
	Fixed bool    `json:"fixed"`
}

type snapLine struct {
	Name string `json:"name"`
	A    string `json:"a"`
	B    string `json:"b"`
}

type snapPlane struct {
	NX float64 `json:"nx"`
	NY float64 `json:"ny"`
	NZ float64 `json:"nz"`
	PX float64 `json:"px"`
	PY float64 `json:"py"`
	PZ float64 `json:"pz"`
}

type snapConstraint struct {
	Kind  string  `json:"kind"`
	A     string  `json:"a"`
	B     string  `json:"b,omitempty"`
	Value float64 `json:"value,omitempty"`
}

type snapFile struct {
	Name        string               `json:"name"`
	Dimension   string               `json:"dimension"`
	Before      map[string]snapPoint `json:"before"`
	After       map[string]snapPoint `json:"after"`
	Lines       []snapLine           `json:"lines,omitempty"`
	Planes      []snapPlane          `json:"planes,omitempty"`
	Constraints []snapConstraint     `json:"constraints,omitempty"`
}

// ---- Viewer state ----

const leftPanelW = 200
const rightPanelW = 240

type Viewer struct {
	folder    string
	snapshots []snapFile
	names     []string
	selected  int
	scrollY   int

	// folder path input (shown when folder is empty)
	inputText  string
	inputRunes []rune

	// 3D rotation and zoom
	rotX, rotY float64
	zoom       float64 // 3D zoom
	dragActive bool
	dragLastX  int
	dragLastY  int

	// 2D zoom (mouse wheel in scene area)
	zoom2D float64

	// 2D pan (left-mouse drag in scene area)
	pan2DX, pan2DY   float64
	drag2DActive     bool
	drag2DLastX      int
	drag2DLastY      int
	// world-units per screen pixel — set each Draw(), consumed in Update()
	vpScaleX, vpScaleY float64

	// logical screen size — set each Draw() call, read in Update()
	screenW, screenH int
}

func loadSnapshots(folder string) ([]snapFile, []string, error) {
	entries, err := os.ReadDir(folder)
	if err != nil {
		return nil, nil, err
	}
	var snaps []snapFile
	var names []string
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".json") {
			continue
		}
		path := filepath.Join(folder, e.Name())
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		var sf snapFile
		if err := json.Unmarshal(data, &sf); err != nil {
			continue
		}
		snaps = append(snaps, sf)
		names = append(names, strings.TrimSuffix(e.Name(), ".json"))
	}
	type pair struct {
		name string
		snap snapFile
	}
	pairs := make([]pair, len(snaps))
	for i := range snaps {
		pairs[i] = pair{names[i], snaps[i]}
	}
	sort.Slice(pairs, func(i, j int) bool { return pairs[i].name < pairs[j].name })
	for i := range pairs {
		names[i] = pairs[i].name
		snaps[i] = pairs[i].snap
	}
	return snaps, names, nil
}

func (v *Viewer) Update() error {
	if ebiten.IsKeyPressed(ebiten.KeyQ) {
		return ebiten.Termination
	}

	if v.folder == "" {
		v.inputRunes = ebiten.AppendInputChars(v.inputRunes[:0])
		v.inputText += string(v.inputRunes)
		if ebiten.IsKeyPressed(ebiten.KeyBackspace) && len(v.inputText) > 0 {
			v.inputText = v.inputText[:len(v.inputText)-1]
		}
		if ebiten.IsKeyPressed(ebiten.KeyEnter) && len(v.inputText) > 0 {
			snaps, names, err := loadSnapshots(v.inputText)
			if err == nil {
				v.folder = v.inputText
				v.snapshots = snaps
				v.names = names
			}
		}
		return nil
	}

	// 3D rotation via mouse drag in the scene panel
	if len(v.snapshots) > 0 && v.snapshots[v.selected].Dimension == "3d" {
		mx, my := ebiten.CursorPosition()
		pressed := ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft)
		inScene := mx >= leftPanelW && mx < v.screenW-rightPanelW
		if pressed && inScene {
			if v.dragActive {
				v.rotY += float64(mx-v.dragLastX) * 0.01
				v.rotX -= float64(my-v.dragLastY) * 0.01
				const limit = math.Pi/2 - 0.01
				if v.rotX > limit {
					v.rotX = limit
				}
				if v.rotX < -limit {
					v.rotX = -limit
				}
			}
			v.dragActive = true
			v.dragLastX = mx
			v.dragLastY = my
		} else {
			v.dragActive = false
		}
	} else {
		v.dragActive = false
	}

	// 2D pan via left-mouse drag in the scene area
	if len(v.snapshots) > 0 && v.snapshots[v.selected].Dimension != "3d" {
		mx, my := ebiten.CursorPosition()
		pressed := ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft)
		inScene := mx >= leftPanelW && mx < v.screenW-rightPanelW
		if pressed && inScene {
			if v.drag2DActive && v.vpScaleX > 0 {
				v.pan2DX -= float64(mx-v.drag2DLastX) * v.vpScaleX
				v.pan2DY += float64(my-v.drag2DLastY) * v.vpScaleY
			}
			v.drag2DActive = true
			v.drag2DLastX = mx
			v.drag2DLastY = my
		} else {
			v.drag2DActive = false
		}
	}

	// +/- keys: 3D zoom
	if ebiten.IsKeyPressed(ebiten.KeyEqual) || ebiten.IsKeyPressed(ebiten.KeyNumpadAdd) {
		v.zoom *= 1.02
	}
	if ebiten.IsKeyPressed(ebiten.KeyMinus) || ebiten.IsKeyPressed(ebiten.KeyNumpadSubtract) {
		v.zoom *= 0.98
		if v.zoom < 0.05 {
			v.zoom = 0.05
		}
	}

	// R: reset 2D zoom and pan
	if isJustPressed(ebiten.KeyR) {
		v.zoom2D = 1.0
		v.pan2DX, v.pan2DY = 0, 0
	}

	// Arrow key navigation
	if isJustPressed(ebiten.KeyArrowDown) {
		if v.selected < len(v.snapshots)-1 {
			v.selected++
			v.zoom2D = 1.0
			v.pan2DX, v.pan2DY = 0, 0
		}
	}
	if isJustPressed(ebiten.KeyArrowUp) {
		if v.selected > 0 {
			v.selected--
			v.zoom2D = 1.0
			v.pan2DX, v.pan2DY = 0, 0
		}
	}

	// Mouse click in left panel
	if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
		mx, my := ebiten.CursorPosition()
		if mx < leftPanelW {
			rowH := 18
			idx := (my + v.scrollY) / rowH
			if idx >= 0 && idx < len(v.snapshots) {
				if v.selected != idx {
					v.zoom2D = 1.0
					v.pan2DX, v.pan2DY = 0, 0
				}
				v.selected = idx
			}
		}
	}

	// Mouse wheel: scroll list when over left panel, zoom scene otherwise
	mx, _ := ebiten.CursorPosition()
	_, dy := ebiten.Wheel()
	if dy != 0 {
		if mx < leftPanelW {
			v.scrollY -= int(dy * 20)
			if v.scrollY < 0 {
				v.scrollY = 0
			}
		} else {
			factor := math.Pow(1.12, dy)
			v.zoom2D *= factor
			v.zoom *= factor // also zoom 3D consistently
			if v.zoom2D < 0.05 {
				v.zoom2D = 0.05
			}
			if v.zoom2D > 100 {
				v.zoom2D = 100
			}
			if v.zoom < 0.05 {
				v.zoom = 0.05
			}
		}
	}

	return nil
}

var prevKeys = map[ebiten.Key]bool{}

func isJustPressed(k ebiten.Key) bool {
	pressed := ebiten.IsKeyPressed(k)
	was := prevKeys[k]
	prevKeys[k] = pressed
	return pressed && !was
}

func (v *Viewer) Draw(screen *ebiten.Image) {
	w, h := screen.Bounds().Dx(), screen.Bounds().Dy()
	v.screenW, v.screenH = w, h

	// Left panel background
	drawRect(screen, 0, 0, leftPanelW, h, color.RGBA{30, 30, 30, 255})
	vector.StrokeLine(screen, float32(leftPanelW), 0, float32(leftPanelW), float32(h), 1, color.RGBA{80, 80, 80, 255}, false)

	// Right panel background
	drawRect(screen, w-rightPanelW, 0, rightPanelW, h, color.RGBA{30, 30, 30, 255})
	vector.StrokeLine(screen, float32(w-rightPanelW), 0, float32(w-rightPanelW), float32(h), 1, color.RGBA{80, 80, 80, 255}, false)

	if v.folder == "" {
		ebitenutil.DebugPrintAt(screen, "Snapshot folder:", 8, 8)
		ebitenutil.DebugPrintAt(screen, "> "+v.inputText+"_", 8, 24)
		return
	}

	// Snapshot list
	rowH := 18
	for i, name := range v.names {
		y := i*rowH - v.scrollY
		if y < 0 || y > h {
			continue
		}
		if i == v.selected {
			drawRect(screen, 0, y, leftPanelW, rowH, color.RGBA{50, 80, 150, 255})
		}
		label := name
		if len(label) > 22 {
			label = label[:19] + "..."
		}
		ebitenutil.DebugPrintAt(screen, label, 6, y+2)
	}

	if len(v.snapshots) == 0 {
		ebitenutil.DebugPrintAt(screen, "No snapshots found", 6, 8)
		return
	}

	snap := v.snapshots[v.selected]
	sceneX := leftPanelW + 1
	sceneW := w - sceneX - rightPanelW
	halfW := sceneW / 2

	vector.StrokeLine(screen, float32(sceneX+halfW), 0, float32(sceneX+halfW), float32(h), 1, color.RGBA{80, 80, 80, 255}, false)

	ebitenutil.DebugPrintAt(screen, snap.Name, sceneX+8, 4)

	if snap.Dimension == "3d" {
		ebitenutil.DebugPrintAt(screen, "drag to rotate  scroll to zoom", sceneX+8, h-16)
		draw3DPanel(screen, snap.Before, snap, sceneX, 20, halfW, h-20, "Before", v.rotX, v.rotY, v.zoom)
		draw3DPanel(screen, snap.After, snap, sceneX+halfW+1, 20, halfW, h-20, "After", v.rotX, v.rotY, v.zoom)
	} else {
		ebitenutil.DebugPrintAt(screen, "drag to pan  scroll to zoom  R to reset", sceneX+8, h-16)
		var scale [2]float64
		draw2DPanel(screen, snap.Before, snap, sceneX, 20, halfW, h-20, "Before", v.zoom2D, v.pan2DX, v.pan2DY, &scale)
		v.vpScaleX, v.vpScaleY = scale[0], scale[1]
		draw2DPanel(screen, snap.After, snap, sceneX+halfW+1, 20, halfW, h-20, "After", v.zoom2D, v.pan2DX, v.pan2DY, nil)
	}

	// Right info panel
	drawInfoPanel(screen, snap, w-rightPanelW+8, 8, rightPanelW-16, h)
}

func drawRect(screen *ebiten.Image, x, y, w, h int, c color.Color) {
	vector.DrawFilledRect(screen, float32(x), float32(y), float32(w), float32(h), c, false)
}

// ---- Info panel ----

func drawInfoPanel(screen *ebiten.Image, snap snapFile, x, y, w, h int) {
	lineH := 16
	cur := y

	write := func(s string, c color.RGBA) {
		ebitenutil.DebugPrintAt(screen, s, x, cur)
		_ = c // DebugPrintAt uses white; color param reserved for future use
		cur += lineH
	}
	header := func(s string) {
		write(s, color.RGBA{180, 180, 80, 255})
		cur += 2
	}

	header("POINTS (after)")

	// Sorted point names
	names := make([]string, 0, len(snap.After))
	for n := range snap.After {
		names = append(names, n)
	}
	sort.Strings(names)

	for _, name := range names {
		p := snap.After[name]
		tag := "free"
		if p.Fixed {
			tag = "fix"
		}
		var line string
		if snap.Dimension == "3d" {
			line = fmt.Sprintf("%-4s (%.2f,%.2f,%.2f) %s", name, p.X, p.Y, p.Z, tag)
		} else {
			line = fmt.Sprintf("%-4s (%.2f, %.2f)  %s", name, p.X, p.Y, tag)
		}
		if cur+lineH > h {
			break
		}
		ebitenutil.DebugPrintAt(screen, line, x, cur)
		cur += lineH
	}

	cur += 6
	if cur+lineH > h {
		return
	}
	header("CONSTRAINTS")

	for _, c := range snap.Constraints {
		if cur+lineH > h {
			break
		}
		var line string
		switch c.Kind {
		case "distance":
			line = fmt.Sprintf("dist  %s–%s = %.4g", c.A, c.B, c.Value)
		case "parallel":
			line = fmt.Sprintf("∥     %s, %s", c.A, c.B)
		case "perpendicular":
			line = fmt.Sprintf("⊥     %s, %s", c.A, c.B)
		case "angle":
			line = fmt.Sprintf("∠     %s,%s = %.0f°", c.A, c.B, c.Value*180/math.Pi)
		default:
			line = fmt.Sprintf("%s  %s %s", c.Kind, c.A, c.B)
		}
		ebitenutil.DebugPrintAt(screen, line, x, cur)
		cur += lineH
	}
}

// ---- 2D panel drawing ----

func draw2DPanel(screen *ebiten.Image, points map[string]snapPoint, snap snapFile, offX, offY, w, h int, title string, zoom, panX, panY float64, scaleOut *[2]float64) {
	// Clip all drawing to this panel's column so it never bleeds into adjacent panels.
	screen = screen.SubImage(image.Rect(offX, 0, offX+w, screen.Bounds().Dy())).(*ebiten.Image)
	ebitenutil.DebugPrintAt(screen, title, offX+w/2-len(title)*3, offY)

	const pad = 40.0
	drawW := float64(w) - 2*pad
	drawH := float64(h) - 2*pad - 16

	minX, maxX, minY, maxY := boundsFromAllSnaps2D(snap)
	rangeX := maxX - minX
	rangeY := maxY - minY
	if rangeX < 1 {
		rangeX = 1
	}
	if rangeY < 1 {
		rangeY = 1
	}
	minX -= rangeX * 0.2
	maxX += rangeX * 0.2
	minY -= rangeY * 0.2
	maxY += rangeY * 0.2

	// Apply zoom around the centre, then pan
	cx := (minX + maxX) / 2
	cy := (minY + maxY) / 2
	halfRX := (maxX - minX) / (2 * zoom)
	halfRY := (maxY - minY) / (2 * zoom)
	minX, maxX = cx-halfRX+panX, cx+halfRX+panX
	minY, maxY = cy-halfRY+panY, cy+halfRY+panY

	// Export world-units-per-pixel for drag conversion in Update()
	if scaleOut != nil {
		scaleOut[0] = (maxX - minX) / drawW
		scaleOut[1] = (maxY - minY) / drawH
	}

	toScreen := func(wx, wy float64) (float32, float32) {
		sx := float32(offX) + float32(pad+(wx-minX)/(maxX-minX)*drawW)
		sy := float32(offY+16) + float32(float64(h)-16-pad-(wy-minY)/(maxY-minY)*drawH)
		return sx, sy
	}

	axisC := color.RGBA{70, 70, 70, 255}
	if minX <= 0 && 0 <= maxX {
		ax, ay1 := toScreen(0, minY)
		_, ay2 := toScreen(0, maxY)
		vector.StrokeLine(screen, ax, ay1, ax, ay2, 1, axisC, false)
	}
	if minY <= 0 && 0 <= maxY {
		ax1, ay := toScreen(minX, 0)
		ax2, _ := toScreen(maxX, 0)
		vector.StrokeLine(screen, ax1, ay, ax2, ay, 1, axisC, false)
	}

	lineC := color.RGBA{100, 100, 200, 255}
	for _, l := range snap.Lines {
		a, b := points[l.A], points[l.B]
		ax, ay := toScreen(a.X, a.Y)
		bx, by := toScreen(b.X, b.Y)
		vector.StrokeLine(screen, ax, ay, bx, by, 1.5, lineC, false)
	}

	// Build line-name → endpoints lookup for constraint rendering.
	lineEndpts := map[string][2]snapPoint{}
	for _, l := range snap.Lines {
		a, aok := points[l.A]
		b, bok := points[l.B]
		if aok && bok {
			lineEndpts[l.Name] = [2]snapPoint{a, b}
		}
	}
	for _, c := range snap.Constraints {
		switch c.Kind {
		case "distance":
			a, aok := points[c.A]
			b, bok := points[c.B]
			if !aok || !bok {
				continue
			}
			drawDistanceAnnotation(screen, toScreen, a, b, c.Value)
		case "parallel":
			ab, ok1 := lineEndpts[c.A]
			cd, ok2 := lineEndpts[c.B]
			if !ok1 || !ok2 {
				continue
			}
			drawParallelAnnotation(screen, toScreen, ab[0], ab[1])
			drawParallelAnnotation(screen, toScreen, cd[0], cd[1])
		case "perpendicular":
			ab, ok1 := lineEndpts[c.A]
			cd, ok2 := lineEndpts[c.B]
			if !ok1 || !ok2 {
				continue
			}
			drawPerpAnnotation(screen, toScreen, ab[0], ab[1], cd[0], cd[1])
		case "angle":
			ab, ok1 := lineEndpts[c.A]
			cd, ok2 := lineEndpts[c.B]
			if !ok1 || !ok2 {
				continue
			}
			drawAngleAnnotation(screen, toScreen, ab[0], ab[1], cd[0], cd[1], c.Value)
		}
	}

	for name, p := range points {
		px, py := toScreen(p.X, p.Y)
		var c color.RGBA
		if p.Fixed {
			c = color.RGBA{50, 100, 210, 255}
		} else {
			c = color.RGBA{210, 60, 60, 255}
		}
		vector.DrawFilledCircle(screen, px, py, 5, c, false)
		vector.StrokeCircle(screen, px, py, 5, 1.5, color.White, false)
		ebitenutil.DebugPrintAt(screen, name, int(px)+8, int(py)-6)
	}
}

func boundsFromAllSnaps2D(snap snapFile) (minX, maxX, minY, maxY float64) {
	minX, maxX = math.Inf(1), math.Inf(-1)
	minY, maxY = math.Inf(1), math.Inf(-1)
	for _, pts := range []map[string]snapPoint{snap.Before, snap.After} {
		for _, p := range pts {
			if p.X < minX {
				minX = p.X
			}
			if p.X > maxX {
				maxX = p.X
			}
			if p.Y < minY {
				minY = p.Y
			}
			if p.Y > maxY {
				maxY = p.Y
			}
		}
	}
	return
}

// ---- 3D panel drawing ----

// project3D applies yaw (around world Y) then pitch (around local X), returning orthographic 2D coords.
func project3D(x, y, z, rotY, rotX float64) (float64, float64) {
	// Yaw
	x1 := x*math.Cos(rotY) + z*math.Sin(rotY)
	y1 := y
	z1 := -x*math.Sin(rotY) + z*math.Cos(rotY)
	// Pitch
	y2 := y1*math.Cos(rotX) + z1*math.Sin(rotX)
	return x1, y2
}

// worldBounds3D returns the 3D centroid and bounding radius across both states.
// Using world-space bounds keeps the scale constant as the user rotates.
func worldBounds3D(snap snapFile) (cx, cy, cz, radius float64) {
	minX, maxX := math.Inf(1), math.Inf(-1)
	minY, maxY := math.Inf(1), math.Inf(-1)
	minZ, maxZ := math.Inf(1), math.Inf(-1)
	for _, pts := range []map[string]snapPoint{snap.Before, snap.After} {
		for _, p := range pts {
			if p.X < minX {
				minX = p.X
			}
			if p.X > maxX {
				maxX = p.X
			}
			if p.Y < minY {
				minY = p.Y
			}
			if p.Y > maxY {
				maxY = p.Y
			}
			if p.Z < minZ {
				minZ = p.Z
			}
			if p.Z > maxZ {
				maxZ = p.Z
			}
		}
	}
	if math.IsInf(minX, 1) {
		return 0, 0, 0, 1
	}
	cx = (minX + maxX) / 2
	cy = (minY + maxY) / 2
	cz = (minZ + maxZ) / 2
	dx, dy, dz := maxX-minX, maxY-minY, maxZ-minZ
	// Diagonal of the bounding box — ensures no clipping at any rotation angle.
	radius = math.Sqrt(dx*dx+dy*dy+dz*dz)/2*1.3 + 1
	return
}

func cross3(a, b [3]float64) [3]float64 {
	return [3]float64{
		a[1]*b[2] - a[2]*b[1],
		a[2]*b[0] - a[0]*b[2],
		a[0]*b[1] - a[1]*b[0],
	}
}

func normalize3(a [3]float64) [3]float64 {
	l := math.Sqrt(a[0]*a[0] + a[1]*a[1] + a[2]*a[2])
	if l < 1e-10 {
		return [3]float64{1, 0, 0}
	}
	return [3]float64{a[0] / l, a[1] / l, a[2] / l}
}

func drawPlane3D(screen *ebiten.Image, plane snapPlane, radius float64, toScreen func(float64, float64, float64) (float32, float32)) {
	n := normalize3([3]float64{plane.NX, plane.NY, plane.NZ})

	up := [3]float64{0, 1, 0}
	if math.Abs(n[1]) > 0.9 {
		up = [3]float64{0, 0, 1}
	}
	t1 := normalize3(cross3(n, up))
	t2 := cross3(n, t1)

	halfSize := radius * 0.8
	px, py, pz := plane.PX, plane.PY, plane.PZ

	corners := [4][3]float64{
		{px + halfSize*(+t1[0]+t2[0]), py + halfSize*(+t1[1]+t2[1]), pz + halfSize*(+t1[2]+t2[2])},
		{px + halfSize*(+t1[0]-t2[0]), py + halfSize*(+t1[1]-t2[1]), pz + halfSize*(+t1[2]-t2[2])},
		{px + halfSize*(-t1[0]-t2[0]), py + halfSize*(-t1[1]-t2[1]), pz + halfSize*(-t1[2]-t2[2])},
		{px + halfSize*(-t1[0]+t2[0]), py + halfSize*(-t1[1]+t2[1]), pz + halfSize*(-t1[2]+t2[2])},
	}

	planeC := color.RGBA{80, 180, 210, 200}
	for i := range corners {
		a, b := corners[i], corners[(i+1)%4]
		ax, ay := toScreen(a[0], a[1], a[2])
		bx, by := toScreen(b[0], b[1], b[2])
		vector.StrokeLine(screen, ax, ay, bx, by, 1.5, planeC, false)
	}
	// Diagonals to make the surface read as a face, not just a frame
	dimC := color.RGBA{80, 180, 210, 100}
	ax0, ay0 := toScreen(corners[0][0], corners[0][1], corners[0][2])
	ax2, ay2 := toScreen(corners[2][0], corners[2][1], corners[2][2])
	ax1, ay1 := toScreen(corners[1][0], corners[1][1], corners[1][2])
	ax3, ay3 := toScreen(corners[3][0], corners[3][1], corners[3][2])
	vector.StrokeLine(screen, ax0, ay0, ax2, ay2, 0.5, dimC, false)
	vector.StrokeLine(screen, ax1, ay1, ax3, ay3, 0.5, dimC, false)
}

func draw3DPanel(screen *ebiten.Image, points map[string]snapPoint, snap snapFile, offX, offY, w, h int, title string, rotX, rotY, zoom float64) {
	screen = screen.SubImage(image.Rect(offX, 0, offX+w, screen.Bounds().Dy())).(*ebiten.Image)
	ebitenutil.DebugPrintAt(screen, title, offX+w/2-len(title)*3, offY)

	const pad = 40.0
	drawW := float64(w) - 2*pad
	drawH := float64(h) - 2*pad - 16

	cx, cy, cz, radius := worldBounds3D(snap)
	scale := math.Min(drawW, drawH) / 2 / radius * zoom

	centerX := float64(offX) + float64(w)/2
	centerY := float64(offY+16) + float64(h-16)/2

	toScreen := func(wx, wy, wz float64) (float32, float32) {
		px, py := project3D(wx-cx, wy-cy, wz-cz, rotY, rotX)
		return float32(centerX + px*scale), float32(centerY - py*scale)
	}

	// Planes (drawn before axes and points so they sit in the background)
	for _, plane := range snap.Planes {
		drawPlane3D(screen, plane, radius, toScreen)
	}

	// Axes from world origin
	axisLen := radius * 0.6
	ox, oy := toScreen(0, 0, 0)
	xTx, xTy := toScreen(axisLen, 0, 0)
	yTx, yTy := toScreen(0, axisLen, 0)
	zTx, zTy := toScreen(0, 0, axisLen)
	vector.StrokeLine(screen, ox, oy, xTx, xTy, 1.5, color.RGBA{220, 80, 80, 255}, false)
	vector.StrokeLine(screen, ox, oy, yTx, yTy, 1.5, color.RGBA{60, 180, 60, 255}, false)
	vector.StrokeLine(screen, ox, oy, zTx, zTy, 1.5, color.RGBA{80, 80, 220, 255}, false)
	ebitenutil.DebugPrintAt(screen, "x", int(xTx)+3, int(xTy))
	ebitenutil.DebugPrintAt(screen, "y", int(yTx)+3, int(yTy))
	ebitenutil.DebugPrintAt(screen, "z", int(zTx)+3, int(zTy))

	for name, p := range points {
		sx, sy := toScreen(p.X, p.Y, p.Z)
		var c color.RGBA
		if p.Fixed {
			c = color.RGBA{50, 100, 210, 255}
		} else {
			c = color.RGBA{210, 60, 60, 255}
		}
		vector.DrawFilledCircle(screen, sx, sy, 5, c, false)
		vector.StrokeCircle(screen, sx, sy, 5, 1.5, color.White, false)
		label := fmt.Sprintf("%s (%.1f,%.1f,%.1f)", name, p.X, p.Y, p.Z)
		ebitenutil.DebugPrintAt(screen, label, int(sx)+8, int(sy)-6)
	}
}

// ---- Constraint annotations ----

var annotC = color.RGBA{80, 210, 130, 220}

// drawDistanceAnnotation draws a thin line between two points with the distance label.
func drawDistanceAnnotation(screen *ebiten.Image, toScreen func(float64, float64) (float32, float32), a, b snapPoint, dist float64) {
	ax, ay := toScreen(a.X, a.Y)
	bx, by := toScreen(b.X, b.Y)
	vector.StrokeLine(screen, ax, ay, bx, by, 1, color.RGBA{200, 160, 50, 130}, false)
	mx, my := (ax+bx)/2, (ay+by)/2
	dx := float64(bx - ax)
	dy := float64(by - ay)
	l := math.Sqrt(dx*dx + dy*dy)
	ox, oy := 0.0, -12.0
	if l > 1 {
		ox, oy = -dy/l*12, dx/l*12
	}
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("%.4g", dist), int(mx)+int(ox), int(my)+int(oy))
}

// drawParallelAnnotation draws two short tick marks perpendicular to the line at its midpoint.
func drawParallelAnnotation(screen *ebiten.Image, toScreen func(float64, float64) (float32, float32), a, b snapPoint) {
	ax, ay := toScreen(a.X, a.Y)
	bx, by := toScreen(b.X, b.Y)
	mx, my := (ax+bx)/2, (ay+by)/2
	dx := float64(bx - ax)
	dy := float64(by - ay)
	l := math.Sqrt(dx*dx + dy*dy)
	if l < 1 {
		return
	}
	ux, uy := dx/l, dy/l // unit along line
	px, py := -uy, ux    // unit perpendicular
	tick, gap := 5.0, 4.0
	for _, s := range []float64{-1, 1} {
		cx := float64(mx) + s*gap*ux
		cy := float64(my) + s*gap*uy
		vector.StrokeLine(screen,
			float32(cx-px*tick), float32(cy-py*tick),
			float32(cx+px*tick), float32(cy+py*tick),
			1.5, annotC, false)
	}
}

// drawPerpAnnotation draws a small L-box at the shared vertex of two perpendicular lines.
func drawPerpAnnotation(screen *ebiten.Image, toScreen func(float64, float64) (float32, float32), a, b, c, d snapPoint) {
	sax, say := toScreen(a.X, a.Y)
	sbx, sby := toScreen(b.X, b.Y)
	scx, scy := toScreen(c.X, c.Y)
	sdx, sdy := toScreen(d.X, d.Y)

	type pt struct{ x, y float32 }
	eps1 := [2]pt{{sax, say}, {sbx, sby}}
	eps2 := [2]pt{{scx, scy}, {sdx, sdy}}

	// Find the closest endpoint pair to locate the vertex.
	vx, vy := sax, say
	var d1x, d1y, d2x, d2y float64
	minDist := math.MaxFloat64
	for i, p1 := range eps1 {
		for j, p2 := range eps2 {
			dist := math.Sqrt(float64((p1.x-p2.x)*(p1.x-p2.x) + (p1.y-p2.y)*(p1.y-p2.y)))
			if dist < minDist {
				minDist = dist
				vx = (p1.x + p2.x) / 2
				vy = (p1.y + p2.y) / 2
				o1 := eps1[1-i]
				o2 := eps2[1-j]
				d1x, d1y = float64(o1.x-vx), float64(o1.y-vy)
				d2x, d2y = float64(o2.x-vx), float64(o2.y-vy)
			}
		}
	}
	l1 := math.Sqrt(d1x*d1x + d1y*d1y)
	l2 := math.Sqrt(d2x*d2x + d2y*d2y)
	if l1 < 1 || l2 < 1 {
		return
	}
	d1x, d1y = d1x/l1, d1y/l1
	d2x, d2y = d2x/l2, d2y/l2
	sz := float64(8)
	p1x := float32(float64(vx) + d1x*sz)
	p1y := float32(float64(vy) + d1y*sz)
	p2x := float32(float64(vx) + d1x*sz + d2x*sz)
	p2y := float32(float64(vy) + d1y*sz + d2y*sz)
	p3x := float32(float64(vx) + d2x*sz)
	p3y := float32(float64(vy) + d2y*sz)
	vector.StrokeLine(screen, p1x, p1y, p2x, p2y, 1.5, annotC, false)
	vector.StrokeLine(screen, p2x, p2y, p3x, p3y, 1.5, annotC, false)
}

// drawAngleAnnotation draws a small arc between two lines at their shared vertex with a degree label.
func drawAngleAnnotation(screen *ebiten.Image, toScreen func(float64, float64) (float32, float32), a, b, c, d snapPoint, angle float64) {
	sax, say := toScreen(a.X, a.Y)
	sbx, sby := toScreen(b.X, b.Y)
	scx, scy := toScreen(c.X, c.Y)
	sdx, sdy := toScreen(d.X, d.Y)

	type pt struct{ x, y float32 }
	eps1 := [2]pt{{sax, say}, {sbx, sby}}
	eps2 := [2]pt{{scx, scy}, {sdx, sdy}}

	vx, vy := sax, say
	var d1x, d1y, d2x, d2y float64
	minDist := math.MaxFloat64
	for i, p1 := range eps1 {
		for j, p2 := range eps2 {
			dist := math.Sqrt(float64((p1.x-p2.x)*(p1.x-p2.x) + (p1.y-p2.y)*(p1.y-p2.y)))
			if dist < minDist {
				minDist = dist
				vx = (p1.x + p2.x) / 2
				vy = (p1.y + p2.y) / 2
				o1 := eps1[1-i]
				o2 := eps2[1-j]
				d1x, d1y = float64(o1.x-vx), float64(o1.y-vy)
				d2x, d2y = float64(o2.x-vx), float64(o2.y-vy)
			}
		}
	}
	l1 := math.Sqrt(d1x*d1x + d1y*d1y)
	l2 := math.Sqrt(d2x*d2x + d2y*d2y)
	if l1 < 1 || l2 < 1 {
		return
	}
	d1x, d1y = d1x/l1, d1y/l1
	d2x, d2y = d2x/l2, d2y/l2

	startAngle := math.Atan2(d1y, d1x)
	endAngle := math.Atan2(d2y, d2x)
	sweep := endAngle - startAngle
	for sweep > math.Pi {
		sweep -= 2 * math.Pi
	}
	for sweep < -math.Pi {
		sweep += 2 * math.Pi
	}

	arcR := 18.0
	nSeg := 16
	prev := func(k int) (float32, float32) {
		theta := startAngle + float64(k)/float64(nSeg)*sweep
		return float32(float64(vx) + arcR*math.Cos(theta)), float32(float64(vy) + arcR*math.Sin(theta))
	}
	for k := 1; k <= nSeg; k++ {
		x0, y0 := prev(k - 1)
		x1, y1 := prev(k)
		vector.StrokeLine(screen, x0, y0, x1, y1, 1.5, annotC, false)
	}
	midTheta := startAngle + sweep/2
	lx := float64(vx) + (arcR+12)*math.Cos(midTheta)
	ly := float64(vy) + (arcR+12)*math.Sin(midTheta)
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("%.0f°", angle*180/math.Pi), int(lx)-8, int(ly)-6)
}

func (v *Viewer) Layout(outsideWidth, outsideHeight int) (int, int) {
	return outsideWidth, outsideHeight
}

func LaunchViewer(folder string) {
	v := &Viewer{
		folder: folder,
		rotX:   math.Pi / 6,
		rotY:   math.Pi / 4,
		zoom:   1.0,
		zoom2D: 1.0,
	}

	if folder != "" {
		snaps, names, err := loadSnapshots(folder)
		if err != nil {
			log.Printf("could not load snapshots from %q: %v", folder, err)
		} else {
			v.snapshots = snaps
			v.names = names
		}
	}

	ebiten.SetWindowTitle("Constraint Solver Viewer")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	sw, sh := ebiten.ScreenSizeInFullscreen()
	if sw <= 0 || sh <= 0 {
		sw, sh = 1280, 760
	}
	ebiten.SetWindowSize(sw, sh)
	if err := ebiten.RunGame(v); err != nil {
		log.Fatal(err)
	}
}
