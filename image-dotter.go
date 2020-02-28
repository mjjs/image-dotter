package main

import (
	"fmt"
	"image"
	"image/draw"
	_ "image/jpeg"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"time"
)

func getAndRunExitListener() chan bool {
	exitSignal := make(chan bool)

	go func() {
		var ask string
		for ask != "q" {
			fmt.Scanln(&ask)
		}
		exitSignal <- true
	}()

	return exitSignal
}

func main() {
	rand.Seed(time.Now().UnixNano())

	if len(os.Args) != 2 {
		fmt.Printf("Usage: %s filename\n", os.Args[0])
		return
	}

	srcFilename, err := filepath.Abs(os.Args[1])
	if err != nil {
		panic(err)
	}

	destFilename, err := filepath.Abs(fmt.Sprintf("out_%s", filepath.Base(srcFilename)))
	if err != nil {
		panic(err)
	}

	srcImage, imageType, err := openImage(srcFilename)

	if err != nil {
		panic(err)
	}

	sourceImageData := extractImageData(srcImage)

	dstImage, _, err := openImage(destFilename)
	if err != nil {
		dstImage = createBlankImage(sourceImageData.bounds)
	}

	exitSignal := getAndRunExitListener()

	for i := 0; ; i++ {
		if i%5000 == 0 {
			fmt.Printf("\rq + enter to write image to disk. Current iteration: %d.", i)
		}

		tempImage := image.NewRGBA(sourceImageData.bounds)

		// Copy the destination image to the temporary image
		draw.Draw(tempImage, sourceImageData.bounds, dstImage, sourceImageData.bounds.Min, draw.Src)
		mask := getRandomCircularMaskWithinImage(sourceImageData)
		colour := getRandomColourFromPixels(sourceImageData.pixels)

		// Draw a random colour on the source file through a given mask
		draw.DrawMask(tempImage, sourceImageData.bounds, &image.Uniform{colour}, image.ZP, mask, image.ZP, draw.Over)

		dstImage, err = getMoreSimilarImage(dstImage, tempImage, srcImage, mask.Bounds())

		if err != nil {
			panic(err)
		}

		// Check for the exit signal and break the loop if found
		select {
		case <-exitSignal:
			fmt.Println("Writing image to disk")
			if err := writeImageToDisk(imageType, dstImage, destFilename); err != nil {
				log.Fatalf("Could not write destination image to disk: %s\n", err.Error())
			}
			return
		default:
			continue
		}
	}

}
