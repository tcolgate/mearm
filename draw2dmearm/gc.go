package draw2dmearm

import (
	"fmt"

	"github.com/llgcode/draw2d"
	"github.com/llgcode/draw2d/draw2dbase"
)

type GraphicContext struct {
	*draw2dbase.StackGraphicContext
	drawer Drawer
}

type Drawer interface {
	MoveTo(x, y float64)
	LineTo(x, y float64)
	End()
}

// NewGraphicContext creates a new Graphic context from an image.
func NewGraphicContext(d Drawer) *GraphicContext {
	gc := &GraphicContext{
		draw2dbase.NewStackGraphicContext(),
		d,
	}
	return gc
}

func (gc *GraphicContext) Stroke(paths ...*draw2d.Path) {
	paths = append(paths, gc.Current.Path)

	stroker := draw2dbase.NewLineStroker(
		gc.Current.Cap,
		gc.Current.Join,
		draw2dbase.Transformer{
			Tr:        gc.Current.Tr,
			Flattener: &mearmLineBuilder{0, 0, gc.drawer}})
	stroker.HalfLineWidth = gc.Current.LineWidth / 2

	for i, p := range paths {
		fmt.Printf("path[%d]: %v\n", i, p)
		draw2dbase.Flatten(p, stroker, gc.Current.Tr.GetScale())
	}
}

type mearmLineBuilder struct {
	startx, starty float64
	d              Drawer
}

// MoveTo Start a New line from the point (x, y)
func (m *mearmLineBuilder) MoveTo(x, y float64) {
	m.startx, m.starty = x, y
	m.d.MoveTo(x, y)
}

// LineTo Draw a line from the current position to the point (x, y)
func (m *mearmLineBuilder) LineTo(x, y float64) {
	m.d.LineTo(x, y)
}

// LineJoin use Round, Bevel or miter to join points
func (m *mearmLineBuilder) LineJoin() {
}

// Close add the most recent starting point to close the path to create a polygon
func (m *mearmLineBuilder) Close() {
	m.LineTo(m.startx, m.starty)
}

// End mark the current line as finished so we can draw caps
func (m *mearmLineBuilder) End() {
	m.d.End()
}
