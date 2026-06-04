package main

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
)

const cecilSourcePath = "docs/2026-06-04/小灰抠图.png"

var cecilOutputs = []struct {
	name string
	file string
}{
	{name: "shy", file: "client/atlas-foods-client/public/assets/story/characters/cecil_shy.png"},
	{name: "formal", file: "client/atlas-foods-client/public/assets/story/characters/cecil_formal.png"},
	{name: "smile", file: "client/atlas-foods-client/public/assets/story/characters/cecil_smile.png"},
}

func main() {
	srcFile, err := os.Open(cecilSourcePath)
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

	for index, output := range cecilOutputs {
		x0 := bounds.Min.X + index*cellW
		x1 := x0 + cellW
		if index == len(cecilOutputs)-1 {
			x1 = bounds.Max.X
		}

		rect := image.Rect(x0, bounds.Min.Y, x1, bounds.Max.Y)
		cropped := image.NewNRGBA(image.Rect(0, 0, rect.Dx(), rect.Dy()))
		for y := rect.Min.Y; y < rect.Max.Y; y++ {
			for x := rect.Min.X; x < rect.Max.X; x++ {
				r16, g16, b16, a16 := img.At(x, y).RGBA()
				r := uint8(r16 >> 8)
				g := uint8(g16 >> 8)
				b := uint8(b16 >> 8)
				a := uint8(a16 >> 8)
				if isCheckerboardPixel(r, g, b) {
					a = 0
				}
				cropped.SetNRGBA(x-rect.Min.X, y-rect.Min.Y, color.NRGBA{R: r, G: g, B: b, A: a})
			}
		}

		if err := os.MkdirAll(filepath.Dir(output.file), 0755); err != nil {
			fatal(err)
		}
		outFile, err := os.Create(output.file)
		if err != nil {
			fatal(err)
		}
		if err := png.Encode(outFile, cropped); err != nil {
			_ = outFile.Close()
			fatal(err)
		}
		if err := outFile.Close(); err != nil {
			fatal(err)
		}
		fmt.Printf("wrote %s (%s, %dx%d)\n", output.file, output.name, cropped.Bounds().Dx(), cropped.Bounds().Dy())
	}
}

func isCheckerboardPixel(r, g, b uint8) bool {
	maxChannel := max(int(r), max(int(g), int(b)))
	minChannel := min(int(r), min(int(g), int(b)))
	return maxChannel >= 238 && maxChannel-minChannel <= 10
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
