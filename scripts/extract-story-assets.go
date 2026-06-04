package main

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
)

const (
	sourcePath = "client/atlas-foods-client/public/assets/story/source/boat_sheet_green.png"
	outputPath = "client/atlas-foods-client/public/assets/story/props/arrival_boat.png"
)

func main() {
	srcFile, err := os.Open(sourcePath)
	if err != nil {
		fatal(err)
	}
	defer srcFile.Close()

	img, err := png.Decode(srcFile)
	if err != nil {
		fatal(err)
	}

	bounds := img.Bounds()
	cellW := bounds.Dx() / 3
	cellH := bounds.Dy() / 3

	// Middle-left cell: a clean side-view boat that reads well when approaching the dock.
	cell := image.Rect(bounds.Min.X, bounds.Min.Y+cellH, bounds.Min.X+cellW, bounds.Min.Y+cellH*2)
	transparent := chromaKey(img, cell)
	cropped := cropTransparent(transparent, 8)

	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		fatal(err)
	}
	outFile, err := os.Create(outputPath)
	if err != nil {
		fatal(err)
	}
	defer outFile.Close()

	if err := png.Encode(outFile, cropped); err != nil {
		fatal(err)
	}
	fmt.Printf("wrote %s (%dx%d)\n", outputPath, cropped.Bounds().Dx(), cropped.Bounds().Dy())
}

func chromaKey(img image.Image, rect image.Rectangle) *image.NRGBA {
	out := image.NewNRGBA(image.Rect(0, 0, rect.Dx(), rect.Dy()))
	for y := rect.Min.Y; y < rect.Max.Y; y++ {
		for x := rect.Min.X; x < rect.Max.X; x++ {
			r16, g16, b16, a16 := img.At(x, y).RGBA()
			r := uint8(r16 >> 8)
			g := uint8(g16 >> 8)
			b := uint8(b16 >> 8)
			a := uint8(a16 >> 8)

			alpha := a
			if isGreenScreen(r, g, b) {
				alpha = 0
			} else if isGreenEdge(r, g, b) {
				alpha = 100
			}

			out.SetNRGBA(x-rect.Min.X, y-rect.Min.Y, color.NRGBA{R: r, G: g, B: b, A: alpha})
		}
	}
	return out
}

func isGreenScreen(r, g, b uint8) bool {
	return g > 150 && int(g) > int(r)*2 && int(g) > int(b)*2
}

func isGreenEdge(r, g, b uint8) bool {
	return g > 110 && int(g) > int(r)*3/2 && int(g) > int(b)*3/2
}

func cropTransparent(img *image.NRGBA, padding int) image.Image {
	bounds := img.Bounds()
	minX, minY := bounds.Max.X, bounds.Max.Y
	maxX, maxY := bounds.Min.X, bounds.Min.Y

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			if img.NRGBAAt(x, y).A == 0 {
				continue
			}
			if x < minX {
				minX = x
			}
			if y < minY {
				minY = y
			}
			if x > maxX {
				maxX = x
			}
			if y > maxY {
				maxY = y
			}
		}
	}

	if minX > maxX || minY > maxY {
		return img
	}

	minX = max(bounds.Min.X, minX-padding)
	minY = max(bounds.Min.Y, minY-padding)
	maxX = min(bounds.Max.X, maxX+padding+1)
	maxY = min(bounds.Max.Y, maxY+padding+1)
	return img.SubImage(image.Rect(minX, minY, maxX, maxY))
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
