package views

import (
	"image/color"

	"tinygo.org/x/drivers/st7789"
)

type RectView struct {
	display         *st7789.Device
	color           *color.RGBA
	backgroundColor *color.RGBA
	x, y, w, h      int16
	text            string
}

func (rv *RectView) SetDisplay(device *st7789.Device) *RectView {
	rv.display = device
	return rv
}

func (rv *RectView) SetDimensions(x, y, w, h int16) *RectView {
	rv.x, rv.y, rv.w, rv.h = x, y, w, h
	return rv

}

func (rv *RectView) SetColor(color *color.RGBA) *RectView {
	rv.color = color
	return rv
}

func (rv *RectView) SetBackgroundColor(color *color.RGBA) *RectView {
	rv.backgroundColor = color
	return rv
}

func (rv *RectView) Draw() *RectView {
	rv.display.DrawFastHLine(rv.x, rv.x+rv.w-1, rv.y, *rv.color)
	rv.display.DrawFastHLine(rv.x, rv.x+rv.w-1, rv.y+rv.h-1, *rv.color)
	rv.display.DrawFastVLine(rv.x, rv.y, rv.y+rv.h-1, *rv.color)
	rv.display.DrawFastVLine(rv.x+rv.w-1, rv.y, rv.y+rv.h-1, *rv.color)
	backgroundColor := color.RGBA{0, 0, 0, 255}
	if rv.backgroundColor != nil {
		backgroundColor = *rv.backgroundColor
	}
	rv.display.FillRectangle(rv.x+1, rv.y+1, rv.w-2, rv.h-2, backgroundColor)
	return rv
}
