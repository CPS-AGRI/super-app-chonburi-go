package main

import (
	"fmt"
	"image"
	"image/draw"
	"image/png"
	"log"
	"os"

	"golang.org/x/image/webp"
)

func main() {
	webpPath := `c:\Users\phnjk\super-app-chonburi\public\logo.webp`
	pngPath := `assets/images/logo.png`

	file, err := os.Open(webpPath)
	if err != nil {
		log.Fatalf("failed to open webp file: %v", err)
	}
	defer file.Close()

	img, err := webp.Decode(file)
	if err != nil {
		log.Fatalf("failed to decode webp image: %v", err)
	}

	// Force conversion to standard 8-bit RGBA image
	bounds := img.Bounds()
	rgbaImg := image.NewRGBA(bounds)
	draw.Draw(rgbaImg, bounds, img, bounds.Min, draw.Src)

	outFile, err := os.Create(pngPath)
	if err != nil {
		log.Fatalf("failed to create png file: %v", err)
	}
	defer outFile.Close()

	// Encode the 8-bit NRGBA/RGBA image to PNG
	err = png.Encode(outFile, rgbaImg)
	if err != nil {
		log.Fatalf("failed to encode png image: %v", err)
	}

	fmt.Println("SUCCESS: Converted logo.webp to 8-bit assets/images/logo.png successfully!")
}

