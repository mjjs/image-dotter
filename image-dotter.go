package main

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	_ "image/jpeg"
	"log"
	"math/rand"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

type imageData struct {
	bounds image.Rectangle
	width  int
	height int
	pixels [][]uint8
}

type circle struct {
	p image.Point
	r int
}

func (c *circle) ColorModel() color.Model {
	return color.AlphaModel
}

// Calculate the bounds of the circle from the center point and radius of the circle
func (c *circle) Bounds() image.Rectangle {
	return image.Rect(c.p.X-c.r, c.p.Y-c.r, c.p.X+c.r, c.p.Y+c.r)
}

// Check whether given coordinates are inside the circle or not
func (c *circle) At(x, y int) color.Color {
	xx, yy, rr := float64(x-c.p.X)+0.5, float64(y-c.p.Y)+0.5, float64(c.r)
	if xx*xx+yy*yy < rr*rr {
		return color.Alpha{255}
	}
	return color.Alpha{0}
}

// Listen the user for a q + enter combo and write to the exit channel if found
func exitListener(ch chan bool) {
	var ask string
	for ask != "q" {
		fmt.Scanln(&ask)
	}
	ch <- true
}

func getRandomCoordinates(width, height int) (x int, y int) {
	return rand.Intn(width), rand.Intn(height)
}

func getRandomPoint(sourceImageData *imageData) image.Point {
	x, y := getRandomCoordinates(sourceImageData.width, sourceImageData.height)
	return image.Pt(x, y)
}

func getRandomRadius() int {
	return rand.Intn(15) + 1
}

func getRandomColourFromPixels(pixels [][]uint8) color.RGBA {
	colorIdx := rand.Intn(len(pixels))
	return color.RGBA{
		pixels[colorIdx][0],
		pixels[colorIdx][1],
		pixels[colorIdx][2],
		pixels[colorIdx][3],
	}
}

func getAndRunExitListener() chan bool {
	// Create a channel for exit signaling and run the exit listening function
	exitSignal := make(chan bool)
	go exitListener(exitSignal)

	return exitSignal
}

func main() {
	srcFilename := getSourceFilename()
	srcImage, imageType, err := getSourceImage(srcFilename)

	if err != nil {
		log.Fatal(err.Error())
	}

	sourceImageData := extractImageData(srcImage)

	dstImage, err := loadImageFromFile(getDestFilename(), sourceImageData.bounds)
	if err != nil {
		log.Fatal(err.Error())
	}

	exitSignal := getAndRunExitListener()

Draw:
	for i := 0; ; i++ {
		if i%5000 == 0 {
			fmt.Printf("Enter q + enter to exit. Current iteration: %d\n", i)
		}

		tempImage := image.NewRGBA(sourceImageData.bounds)

		// Copy the destination image to the temporary image
		draw.Draw(tempImage, sourceImageData.bounds, dstImage, sourceImageData.bounds.Min, draw.Src)

		// Get random coordinates to draw at
		center := getRandomPoint(sourceImageData)

		r := getRandomRadius()
		colour := getRandomColourFromPixels(sourceImageData.pixels)
		mask := &circle{center, r}

		// Draw a random colour on the source file through a given mask
		draw.DrawMask(tempImage, sourceImageData.bounds, &image.Uniform{colour}, image.ZP, mask, image.ZP, draw.Over)

		dstImage, err = getMoreSimilarImage(dstImage, tempImage, srcImage, mask.Bounds())

		if err != nil {
			log.Fatal(err.Error())
		}

		// Check for the exit signal and break the loop if found
		select {
		case <-exitSignal:
			break Draw
		default:
			continue
		}
	}

	log.Println("Writing image to disk")

	if err := writeImageToDisk(imageType, dstImage, getDestFilename()); err != nil {
		log.Fatalf("Could not write destination image to disk: %s\n", err.Error())
	}
}
