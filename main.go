package main

import (
	"bytes"
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"log"
	"os"

	"github.com/blackjack/webcam"
)

const ASCIIMAP string = " .:!i><~+_-?][}{1)(/0*#&8%@$"

var CRH int
var CRV int

func main() {
	cx := flag.Int("cx", 20, "X Compression")
	cy := flag.Int("cy", 30, "Y Compression")
	flag.Parse()
	CRH = *cx
	CRV = *cy

	cam, err := webcam.Open("/dev/video0")
	if err != nil {
		log.Fatal(err)
	}
	defer cam.Close()

	err = cam.StartStreaming()
	if err != nil {
		log.Fatal(err)
	}

	for {
		err = cam.WaitForFrame(uint32(5))

		switch err.(type) {
		case nil:
		case *webcam.Timeout:
			fmt.Fprint(os.Stderr, err.Error())
			continue
		default:
			log.Fatal(err)
		}

		frame, err := cam.ReadFrame()
		if err != nil {
			log.Fatal(err)
		}

		matrix := parseImage(byteToImage(frame))
		fmt.Print("\033[H\033[2J")
		for i := 0; i < len(matrix[0]); i++ {
			for j := 0; j < len(matrix); j++ {
				fmt.Print(toAscii(matrix[j][i]))
			}
			fmt.Println()
		}
	}
}

func byteToImage(imgByte []byte) image.Image {
	img, err := jpeg.Decode(bytes.NewReader(imgByte))
	if err != nil {
		log.Fatal(err)
	}
	return img
}

func parseImage(img image.Image) [][]float64 {
	bounds := img.Bounds()
	bx := int((bounds.Max.X - bounds.Min.X) / CRH)
	by := int((bounds.Max.Y - bounds.Min.Y) / CRV)

	img_matrix := make([][]float64, bx)
	for i := range img_matrix {
		img_matrix[i] = make([]float64, by)
	}

	x, y := 0, 0
	for i := bounds.Min.X; i < (bounds.Max.X - (bounds.Max.X % CRH)); i += CRH {
		for j := bounds.Min.Y; j < (bounds.Max.Y - (bounds.Max.Y % CRV)); j += CRV {
			avg_value := 0.0
			for k := i; k < i+CRH; k++ {
				for l := j; l < j+CRV; l++ {
					avg_value += convertRGB(img.At(k, l))
				}
			}
			img_matrix[x][y] = avg_value / float64(CRH*CRV)
			y++
		}
		y = 0
		x++
	}
	return img_matrix
}

func toAscii(shade float64) string {
	index := int(shade / (255.0 / float64(len(ASCIIMAP))))
	return string(ASCIIMAP[index])
}

func convertRGB(c color.Color) float64 {
	r, g, b, _ := c.RGBA()
	shade := 0.299*float64(r) + 0.587*float64(g) + 0.114*float64(b)
	shade *= (255.0 / 65535.0)
	return shade
}
