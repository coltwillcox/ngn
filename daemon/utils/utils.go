package utils

import (
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"image/png"
	"os"

	"github.com/gabriel-vasile/mimetype"
	"github.com/nfnt/resize"
	"github.com/srwiley/oksvg"
	"github.com/srwiley/rasterx"
)

func ExtractFilePath(inputs []interface{}) string {
	filePath := ""
	for _, input := range inputs {
		if input == nil {
			continue
		}
		path, ok := input.(string)
		if !ok {
			continue
		}
		if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
			continue
		}
		filePath = path
		break
	}

	return filePath
}

func GenerateImageData(filePath string) string {
	imageData := ""
	if filePath == "" {
		return imageData
	}
	file, err := os.Open(filePath)
	if err != nil {
		return imageData
	}

	w, h := 30, 30
	img := image.NewRGBA(image.Rect(0, 0, w, h))

	mtype, err := mimetype.DetectFile(filePath)
	fmt.Println(mtype.String(), mtype.Extension())
	switch mtype.String() {
	case "image/svg+xml":
		icon, err := oksvg.ReadIconStream(file)
		if err != nil {
			return imageData
		}
		icon.SetTarget(0, 0, float64(w), float64(h))
		scanner := rasterx.NewScannerGV(w, h, img, img.Bounds())
		raster := rasterx.NewDasher(w, h, scanner)
		icon.Draw(raster, 1.0)
	case "image/jpeg":
		decodedImage, err := jpeg.Decode(file)
		if err != nil {
			return imageData
		}
		decodedImage = resize.Resize(30, 30, decodedImage, resize.Lanczos3)
		draw.Draw(img, img.Bounds(), decodedImage, decodedImage.Bounds().Min, draw.Src)
	case "image/png":
		decodedImage, err := png.Decode(file)
		if err != nil {
			return imageData
		}
		decodedImage = resize.Resize(30, 30, decodedImage, resize.Lanczos3)
		draw.Draw(img, img.Bounds(), decodedImage, decodedImage.Bounds().Min, draw.Src)
	default:
		return imageData
	}

	colors := make([]color.RGBA, 0)
	for y := 0; y < img.Bounds().Max.Y; y++ {
		for x := 0; x < img.Bounds().Max.X; x++ {
			r, g, b, _ := img.At(x, y).RGBA()
			colors = append(colors, color.RGBA{R: uint8(r >> 8), G: uint8(g >> 8), B: uint8(b >> 8), A: uint8(255)})
		}
	}

	for _, c := range colors {
		imageData += fmt.Sprintf("%02x%02x%02x", c.R, c.G, c.B)
	}

	return imageData
}
