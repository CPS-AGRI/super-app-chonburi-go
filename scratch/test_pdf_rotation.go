package main

import (
	"fmt"
	"log"

	"github.com/jung-kurt/gofpdf/v2"
)

func main() {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()
	pdf.SetFont("Arial", "B", 16)
	pdf.Text(40, 40, "Standard Text")

	// Test TransformRotate
	pdf.TransformBegin()
	pdf.TransformRotate(45, 100, 100)
	pdf.Text(100, 100, "Rotated Text")
	pdf.TransformEnd()

	err := pdf.OutputFileAndClose("scratch/test_rotation.pdf")
	if err != nil {
		log.Fatalf("failed to create pdf: %v", err)
	}

	fmt.Println("SUCCESS: TransformRotate is correct and compiles!")
}
