package views

import (
	"image/color"
	"strconv"

	"tinygo.org/x/drivers/st7789"
)

type ImageView struct {
	display         *st7789.Device
	backgroundColor *color.RGBA
	x, y, w, h      int16
	image           string
}

func (iv *ImageView) SetDisplay(device *st7789.Device) *ImageView {
	iv.display = device
	return iv
}

func (iv *ImageView) SetDimensions(x, y, w, h int16) *ImageView {
	iv.x, iv.y, iv.w, iv.h = x, y, w, h
	return iv

}

func (iv *ImageView) SetBackgroundColor(color *color.RGBA) *ImageView {
	iv.backgroundColor = color
	return iv
}

func (iv *ImageView) SetImage(image string) *ImageView {
	if iv.image == image {
		return iv
	}

	iv.image = image
	iv.drawImage()
	return iv
}

func (iv *ImageView) Draw() *ImageView {
	iv.drawImage()
	return iv
}

func (iv *ImageView) drawImage() *ImageView {
	var row = []color.RGBA{}
	row = make([]color.RGBA, iv.w)
	imageData := []byte(iv.image)
	for i := 0; i < int(iv.h); i++ {
		for j := 0; j < int(iv.w); j++ {
			if len(imageData)-1 < 6*(int(iv.w)*i+j) || len(imageData)-1 < 6*(int(iv.w)*i+j+1) {
				row[j] = color.RGBA{
					R: iv.backgroundColor.R,
					G: iv.backgroundColor.G,
					B: iv.backgroundColor.B,
					A: iv.backgroundColor.A,
				}
				continue
			}

			pixel := imageData[6*(int(iv.w)*i+j) : 6*(int(iv.w)*i+j+1)]
			values, _ := strconv.ParseUint(string(pixel), 16, 32)

			row[j] = color.RGBA{
				R: uint8(values >> 16),
				G: uint8((values >> 8) & 0xFF),
				B: uint8(values & 0xFF),
				A: 255,
			}
		}
		iv.display.FillRectangleWithBuffer(iv.x, iv.y+int16(i), iv.w, 1, row)
	}

	return iv
}
