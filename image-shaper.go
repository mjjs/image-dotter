package main

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/gif"
	_ "image/jpeg"
	"image/png"
	"log"
	"math"
	"math/rand"
	"os"
	"time"
)

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

// compareImages compares two images and returns a 64bit integer indicating how similar they are.
// The higher the number, the different the images.
func compareImages(a, b *image.RGBA) (int64, error) {
	if a.Bounds() != b.Bounds() {
		return -1, fmt.Errorf("The images differ in size: %+v, %+v", a.Bounds(), b.Bounds())
	}

	diff := int64(0)

	for i := 0; i < len(a.Pix); i++ {
		diff += int64(sqrtDiff(a.Pix[i], b.Pix[i]))
	}

	return int64(math.Sqrt(float64(diff))), nil
}

func sqrtDiff(x, y uint8) uint64 {
	d := uint64(x) - uint64(y)
	return d * d
}

// getMoreSimilarImage compares image a and image b against a source image and returns the image which is more similar with the source image
func getMoreSimilarImage(a, b, src *image.RGBA) (*image.RGBA, error) {
	aNum, err := compareImages(a, src)
	if err != nil {
		return nil, err
	}
	bNum, err := compareImages(b, src)
	if err != nil {
		panic(err)
	}

	if aNum > bNum {
		return b, nil
	}
	return a, nil
}

// createBlankImage creates a blank in-memory image the size of srcRect and returns it
func createBlankImage(filename string, srcRect image.Rectangle) *image.RGBA {
	m := image.NewRGBA(srcRect)
	draw.Draw(m, srcRect, &image.Uniform{color.RGBA{0, 0, 0, 0}}, image.ZP, draw.Src)
	return m
}

// getImageA opens an image and returns its contents.
// If Image A does not exist, a blank one will be created and returned
func getImageA(fileName string, sourceRect image.Rectangle) *image.RGBA {
	f, err := os.Open(fileName)
	// If file does not exist, create a blank image and return it
	if os.IsNotExist(err) {
		return createBlankImage(fileName, sourceRect)
	}

	src, _, err := image.Decode(f)
	if err != nil {
		panic(err)
	}

	m := image.NewRGBA(image.Rect(0, 0, src.Bounds().Dx(), src.Bounds().Dy()))
	draw.Draw(m, m.Bounds(), src, src.Bounds().Min, draw.Src)

	f.Close()
	return m
}

// Take an image.Image as argument, draw it to an in-memory RGBA image and return the RGBA image
func imageToRGBA(img image.Image) *image.RGBA {
	m := image.NewRGBA(img.Bounds())
	draw.Draw(m, img.Bounds(), img, img.Bounds().Min, draw.Src)
	return m
}

// Listen the user for a q + enter combo and write to the exit channel if found
func saveAndExit(ch chan bool) {
	var ask string
	for ask != "q" {
		fmt.Scanln(&ask)
	}
	ch <- true
}

// getFilename parses the arguments of the program and returns the filename if found.
// If a filename cannot be found in the arguments, usage info is printed.
func getFilename() string {
	args := os.Args[1:]

	if len(args) != 1 {
		log.Fatal("Usage: image-shaper filename")
	}
	return args[0]
}

func main() {
	srcFilename := getFilename()

	// Open the source file to use for comparison
	srcFile, err := os.Open(srcFilename)
	if err != nil {
		log.Println(err)
		return
	}

	// Decode the source file
	srcImage, imageType, err := image.Decode(srcFile)
	if err != nil {
		log.Println(err)
		return
	}

	// Conver the source image to RGBA
	srcRGBA := imageToRGBA(srcImage)

	srcFile.Close()

	rand.Seed(time.Now().UnixNano())

	// Get the source image data
	srcBounds := srcImage.Bounds()
	srcWidth := srcBounds.Dx()
	srcHeight := srcBounds.Dy()
	srcColours := srcRGBA.Pix

	dstImage := getImageA("out_"+srcFilename, srcBounds)
	tempImage := image.NewRGBA(srcBounds)

	// Create a channel for exit signaling and run the exit listening function
	exitSignal := make(chan bool)
	go saveAndExit(exitSignal)

	loopIter := 0

Draw:
	for {
		fmt.Printf("Enter q + enter to exit. Current iteration: %d\n", loopIter)
		loopIter++

		tempImage = image.NewRGBA(srcBounds)

		// Copy the destination image to the temporary image
		draw.Draw(tempImage, srcBounds, dstImage, srcBounds.Min, draw.Src)

		// Get random coordinates to draw at
		x := rand.Intn(srcWidth)
		y := rand.Intn(srcHeight)
		center := image.Pt(x, y)
		r := rand.Intn(15) // Radius for the circular mask
		if r == 0 {
			r++
		}

		// The colours are in a flat array as [R, G, B, A, R, G, B, A ...]
		// so we get the indexes of all the indexes for the colour red
		var possibleIndexes []int
		for i := 0; i < len(srcColours); i++ {
			if i%4 == 0 || i == 0 {
				possibleIndexes = append(possibleIndexes, i)
			}
		}

		num := possibleIndexes[rand.Intn(len(possibleIndexes))]
		if num == len(srcColours) {
			num = 0
		}

		randomColour := color.RGBA{srcColours[num], srcColours[num+1], srcColours[num+2], srcColours[num+3]}

		// Draw a random colour on the source file through a circular mask
		draw.DrawMask(tempImage, srcBounds, &image.Uniform{randomColour}, image.ZP, &circle{center, r}, image.ZP, draw.Over)

		dstImage, err = getMoreSimilarImage(dstImage, tempImage, srcRGBA)
		if err != nil {
			log.Println(err)
			return
		}

		// Check for the exit signal and break the loop if found
		select {
		case <-exitSignal:
			break Draw
		default:
			continue
		}
	}

	// Open the output file
	outFile, err := os.OpenFile("out_"+srcFilename, os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		log.Println(err)
		return
	}

	outImage, err := getMoreSimilarImage(dstImage, tempImage, srcRGBA)
	if err != nil {
		log.Println(err)
		return
	}

	if imageType == "gif" {
		gif.Encode(outFile, outImage, &gif.Options{NumColors: 256})
	} else {
		png.Encode(outFile, outImage)
	}

	outFile.Close()
}
