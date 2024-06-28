package views

import (
	"image/color"

	"tinygo.org/x/drivers/st7789"
	"tinygo.org/x/tinyfont"
)

const (
	padding  int16 = 8
	ellipsis       = "..."
)

type TextView struct {
	display         *st7789.Device
	color           *color.RGBA
	backgroundColor *color.RGBA
	fontColor       *color.RGBA
	font            tinyfont.Fonter
	x, y, w, h      int16
	lines           []string
	onDisplay       bool
}

func (tv *TextView) SetDisplay(device *st7789.Device) *TextView {
	tv.display = device
	return tv
}

func (tv *TextView) SetDimensions(x, y, w, h int16) *TextView {
	tv.x, tv.y, tv.w, tv.h = x, y, w, h
	return tv

}

func (tv *TextView) SetColor(color *color.RGBA) *TextView {
	tv.color = color
	return tv
}

func (tv *TextView) SetBackgroundColor(color *color.RGBA) *TextView {
	tv.backgroundColor = color
	return tv
}

func (tv *TextView) SetFontColor(color *color.RGBA) *TextView {
	tv.fontColor = color
	return tv
}

func (tv *TextView) SetFont(font tinyfont.Fonter) *TextView {
	tv.font = font
	return tv
}

func (tv *TextView) SetText(text string) *TextView {
	charHeight := tv.font.GetGlyph(rune('|')).Height
	maximumLines := (tv.h - padding) / (int16(charHeight) + padding)

	charWidth := tv.font.GetGlyph(rune('_')).Width
	maximumChars := (tv.w - padding) / int16(charWidth)

	chunkz := chunks(text, int(maximumChars), int(maximumLines))

	if len(chunkz[len(chunkz)-1]) > int(maximumChars) {
		chunkz[len(chunkz)-1] = chunkz[len(chunkz)-1][:maximumChars-int16(len(ellipsis))] + ellipsis
	}

	tv.lines = chunkz
	tv.drawText()
	return tv
}

func (tv *TextView) Draw() *TextView {
	tv.display.DrawFastHLine(tv.x, tv.x+tv.w-1, tv.y, *tv.color)
	tv.display.DrawFastHLine(tv.x, tv.x+tv.w-1, tv.y+tv.h-1, *tv.color)
	tv.display.DrawFastVLine(tv.x, tv.y, tv.y+tv.h-1, *tv.color)
	tv.display.DrawFastVLine(tv.x+tv.w-1, tv.y, tv.y+tv.h-1, *tv.color)
	backgroundColor := color.RGBA{0, 0, 0, 255}
	if tv.backgroundColor != nil {
		backgroundColor = *tv.backgroundColor
	}
	tv.display.FillRectangle(tv.x+1, tv.y+1, tv.w-2, tv.h-2, backgroundColor)
	tv.onDisplay = true

	if len(tv.lines) > 0 {
		tv.drawText()
	}

	return tv
}

func (tv *TextView) drawText() *TextView {
	if tv.onDisplay {
		charHeight := tv.font.GetGlyph(rune('|')).Height
		backgroundColor := color.RGBA{0, 0, 0, 255}
		fontColor := color.RGBA{255, 255, 255, 255}
		if tv.backgroundColor != nil {
			backgroundColor = *tv.backgroundColor
		}
		if tv.fontColor != nil {
			fontColor = *tv.fontColor
		}
		tv.display.FillRectangle(tv.x+1, tv.y+1, tv.w-2, tv.h-2, backgroundColor)
		for i := 0; i < len(tv.lines); i++ {
			tinyfont.WriteLine(tv.display, tv.font, tv.x+padding/2, tv.y+20+int16(i)*(int16(charHeight)+padding+padding/4), tv.lines[i], fontColor)
		}
	}
	return tv
}
