package media

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"image/png"
	"log"
	"os"

	"github.com/fogleman/gg"
	"github.com/gabriel-vasile/mimetype"
	"github.com/golang/freetype/truetype"
	"github.com/nfnt/resize"
	"github.com/srwiley/oksvg"
	"github.com/srwiley/rasterx"
	"golang.org/x/image/font/gofont/goregular"
)

const (
	width, height = 30, 30
)

func GenerateImageData(iconFilePath, iconFallback string) string {
	imageData := ""
	img := image.NewRGBA(image.Rect(0, 0, width, height))

	if iconFilePath == "" {
		decodedImage, err := charToImg(iconFallback)
		if err != nil {
			log.Fatalln(err)
		}

		decodedImage = resize.Resize(width, height, decodedImage, resize.Lanczos3)
		draw.Draw(img, img.Bounds(), decodedImage, decodedImage.Bounds().Min, draw.Src)

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

	file, err := os.Open(iconFilePath)
	if err != nil {
		return imageData
	}

	mtype, err := mimetype.DetectFile(iconFilePath)
	fmt.Println(mtype.String(), mtype.Extension())
	switch mtype.String() {
	case "image/svg+xml":
		icon, err := oksvg.ReadIconStream(file)
		if err != nil {
			return imageData
		}
		icon.SetTarget(0, 0, float64(width), float64(height))
		scanner := rasterx.NewScannerGV(width, height, img, img.Bounds())
		raster := rasterx.NewDasher(width, height, scanner)
		icon.Draw(raster, 1.0)

		// Some SVGs are badly rasterized, with resulting R, G, B colors as just zeros,
		// so it will be converted to (at least) usable grayscale.
		ok := false
		for x := 0; x < img.Bounds().Max.X; x++ {
			for y := 0; y < img.Bounds().Max.Y; y++ {
				r, g, b, _ := img.At(x, y).RGBA()
				if r != 0 || g != 0 || b != 0 {
					ok = true
					break
				}
			}
			if ok {
				break
			}
		}
		if !ok {
			for x := 0; x < img.Bounds().Max.X; x++ {
				for y := 0; y < img.Bounds().Max.Y; y++ {
					_, _, _, a := img.At(x, y).RGBA()
					img.Set(x, y, color.RGBA{R: uint8(a), G: uint8(a), B: uint8(a), A: 255})
				}
			}
		}

	case "image/jpeg":
		decodedImage, err := jpeg.Decode(file)
		if err != nil {
			return imageData
		}
		decodedImage = resize.Resize(width, height, decodedImage, resize.Lanczos3)
		draw.Draw(img, img.Bounds(), decodedImage, decodedImage.Bounds().Min, draw.Src)
	case "image/png":
		decodedImage, err := png.Decode(file)
		if err != nil {
			return imageData
		}
		decodedImage = resize.Resize(width, height, decodedImage, resize.Lanczos3)
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

func charToImg(letter string) (image.Image, error) {
	x := float64(width / 2)
	y := float64((height / 2) - 4)

	dc := gg.NewContext(width, height)
	font, err := truetype.Parse(goregular.TTF)
	if err != nil {
		panic("")
	}
	face := truetype.NewFace(font, &truetype.Options{
		Size: width,
	})
	dc.SetFontFace(face)
	dc.SetColor(color.White)
	dc.DrawStringWrapped(letter, x, y, 0.5, 0.5, 0, 0, gg.AlignCenter)

	return dc.Image(), nil
}
