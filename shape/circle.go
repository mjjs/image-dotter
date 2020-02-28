package shape

import (
	"image"
	"image/color"
)

// Circle represents a circle that can be drawn.
// Implements the image interface.
type Circle struct {
	Center image.Point
	Radius int
}

// ColorModel returns the color model of the circle.
func (c *Circle) ColorModel() color.Model {
	return color.AlphaModel
}

// Bounds calculatesthe bounds of the circle from the center point and radius
// of the circle
func (c *Circle) Bounds() image.Rectangle {
	return image.Rect(
		c.Center.X-c.Radius,
		c.Center.Y-c.Radius,
		c.Center.X+c.Radius,
		c.Center.Y+c.Radius,
	)
}

// At returns an alpha value for the given coordinates x and y based on whether
// the coordinates are inside the circle or not.
func (c *Circle) At(x, y int) color.Color {
	xx := float64(x-c.Center.X) + 0.5
	yy := float64(y-c.Center.Y) + 0.5
	rr := float64(c.Radius)

	if xx*xx+yy*yy < rr*rr {
		return color.Alpha{255}
	}

	return color.Alpha{0}
}
