package bring

import (
	"image"
	"image/draw"

	"github.com/tfriedel6/canvas"
	"github.com/tfriedel6/canvas/backend/softwarebackend"
)

type Layer struct {
	width        int
	height       int
	image        *image.RGBA
	gc           *canvas.Canvas
	visible      bool
	modified     bool
	modifiedRect image.Rectangle
	pathOpen     bool
	pathRect     image.Rectangle
	autosize     bool
}

func (l *Layer) updateModifiedRect(modArea image.Rectangle) {
	before := l.modifiedRect
	l.modifiedRect = l.modifiedRect.Union(modArea)
	l.modified = l.modified || !before.Eq(l.modifiedRect)
}

func (l *Layer) resetModified() {
	l.modifiedRect = image.Rectangle{}
	l.modified = false
}

func (l *Layer) setupCanvas() {
	be := softwarebackend.New(l.width, l.height)
	be.Image = l.image
	l.gc = canvas.New(be)
}

func (l *Layer) fitRect(x int, y int, w int, h int) {
	// Calculate bounds
	opBoundX := w + x
	opBoundY := h + y

	// Determine max width
	var resizeWidth int
	if opBoundX > l.width {
		resizeWidth = opBoundX
	} else {
		resizeWidth = l.width
	}

	// Determine max height
	var resizeHeight int
	if opBoundY > l.height {
		resizeHeight = opBoundY
	} else {
		resizeHeight = l.height
	}

	// Resize if necessary
	l.Resize(resizeWidth, resizeHeight)
}

func copyImage(dest draw.Image, x, y int, src image.Image, sr image.Rectangle, op draw.Op) {
	dp := image.Pt(x, y)
	dr := image.Rectangle{Min: dp, Max: dp.Add(sr.Size())}
	draw.Draw(dest, dr, src, sr.Min, op)
}

func (l *Layer) Copy(srcLayer *Layer, srcx, srcy, srcw, srch, x, y int, op draw.Op) {
	srcImg := srcLayer.image
	srcDim := srcImg.Bounds()

	// If entire rectangle outside source canvas, stop
	if srcx >= srcDim.Max.X || srcy >= srcDim.Max.Y {
		return
	}

	// Otherwise, clip rectangle to area
	if srcx+srcw > srcDim.Max.X {
		srcw = srcDim.Max.X - srcx
	}

	if srcy+srch > srcDim.Max.Y {
		srch = srcDim.Max.Y - srcy
	}

	// Stop if nothing to draw.
	if srcw == 0 || srch == 0 {
		return
	}

	if l.autosize {
		l.fitRect(x, y, srcw, srch)
	}

	srcCopyDim := image.Rect(srcx, srcy, srcx+srcw, srcy+srch)
	copyImage(l.image, x, y, srcImg, srcCopyDim, op)
	l.updateModifiedRect(image.Rect(x, y, x+srcw, y+srch))
}

func (l *Layer) Draw(x, y int, src image.Image, op draw.Op) {
	srcDim := src.Bounds()
	if l.autosize {
		l.fitRect(x, y, srcDim.Max.X, srcDim.Max.Y)
	}
	copyImage(l.image, x, y, src, srcDim, op)
	l.updateModifiedRect(image.Rect(x, y, x+srcDim.Max.X, y+srcDim.Max.Y))
}

func (l *Layer) Resize(w int, h int) {
	original := l.image.Bounds()
	if w == l.width && h == l.height {
		return
	}
	newImage := image.NewRGBA(image.Rect(0, 0, w, h))
	draw.Draw(newImage, l.image.Bounds(), l.image, image.Pt(0, 0), draw.Src)
	l.image = newImage
	l.width = w
	l.height = h
	l.setupCanvas()
	l.updateModifiedRect(original.Union(l.image.Bounds()))
}

func (l *Layer) appendToPath(rect image.Rectangle) {
	if !l.pathOpen {
		l.gc.BeginPath()
		l.pathOpen = true
		l.pathRect = image.Rectangle{}
	}
	l.pathRect = l.pathRect.Union(rect)
}

func (l *Layer) endPath() {
	l.updateModifiedRect(l.pathRect)
	l.pathOpen = false
	l.pathRect = image.Rectangle{}
}

func (l *Layer) Rect(x int, y int, width int, height int) {
	l.appendToPath(image.Rect(x, y, x+width, y+height))
	l.gc.Rect(float64(x), float64(y), float64(width), float64(height))
}

func (l *Layer) Fill(r byte, g byte, b byte, a byte, op draw.Op) {
	// Ignores op, as the canvas library does not support it :/
	l.gc.SetFillStyle(r, g, b, a)
	l.gc.Fill()
	l.endPath()
}

type layers map[int]*Layer

func newLayers() layers {
	ls := make(layers)
	ls[0] = newBuffer()
	ls[0].visible = true
	return ls
}

func newBuffer() *Layer {
	l := &Layer{
		image:    image.NewRGBA(image.Rect(0, 0, 0, 0)),
		autosize: true,
	}
	l.setupCanvas()
	return l
}

func newVisibleLayer(l0 *Layer) *Layer {
	l := &Layer{
		width:   l0.width,
		height:  l0.height,
		image:   image.NewRGBA(image.Rect(0, 0, l0.width, l0.height)),
		visible: true,
	}
	l.setupCanvas()
	return l
}

func (ls layers) getDefault() *Layer {
	return ls[0]
}

func (ls layers) get(id int) *Layer {
	if l, ok := ls[id]; ok {
		return l
	}
	if id > 0 {
		ls[id] = newVisibleLayer(ls[0])
	} else {
		ls[id] = newBuffer()
	}
	return ls[id]
}

func (ls layers) delete(id int) {
	if id == 0 {
		return
	}
	ls[0].updateModifiedRect(ls[id].image.Bounds())
	ls[id].image = nil
	ls[id] = nil
	delete(ls, id)
}
