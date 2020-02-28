package main

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/gif"
	"image/png"
	"math"
	"math/rand"
	"os"

	"github.com/mjjs/image-dotter/shape"
)

type imageData struct {
	bounds image.Rectangle
	width  int
	height int
	pixels [][]uint8
}

// createBlankImage creates a blank in-memory image the size of srcRect and returns it
func createBlankImage(srcRect image.Rectangle) *image.RGBA {
	img := image.NewRGBA(srcRect)
	draw.Draw(img, srcRect, &image.Uniform{color.RGBA{0, 0, 0, 0}}, image.ZP, draw.Src)
	return img
}

func cloneImage(src image.Image) *image.RGBA {
	img := image.NewRGBA(src.Bounds())
	draw.Draw(img, src.Bounds(), src, src.Bounds().Min, draw.Src)

	return img
}

func sqrtDiff(x, y uint8) uint64 {
	d := uint64(x) - uint64(y)
	return d * d
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

// getMoreSimilarImage compares image a and image b against a source image and returns the image which is more similar with the source image
func getMoreSimilarImage(a, b, src *image.RGBA, changedArea image.Rectangle) (*image.RGBA, error) {
	// Get subimages of the changed areas of the images so the whole image does not need to be compared
	aChanged := cloneImage(a.SubImage(changedArea))
	bChanged := cloneImage(b.SubImage(changedArea))
	srcChanged := cloneImage(src.SubImage(changedArea))

	aNum, err := compareImages(aChanged, srcChanged)
	if err != nil {
		return nil, err
	}

	bNum, err := compareImages(bChanged, srcChanged)
	if err != nil {
		return nil, err
	}

	if aNum > bNum {
		return b, nil
	}

	return a, nil
}

func openImage(filename string) (rgbaImg *image.RGBA, imgType string, err error) {
	file, err := os.Open(filename)

	if err != nil {
		return nil, "", fmt.Errorf("Could not open file: %s", err.Error())
	}

	defer file.Close()

	img, imgType, err := image.Decode(file)

	if err != nil {
		return nil, "", fmt.Errorf("Could not decode image: %s", err.Error())
	}

	return cloneImage(img), imgType, err
}

// getImagePixels returns an array of uint8 arrays (pixels),
// each of which hold the R, G, B, A values of the given pixel.
func getImagePixels(img *image.RGBA) [][]uint8 {
	pixels := make([][]uint8, img.Bounds().Dx()*img.Bounds().Dy())

	for i := 0; i < len(img.Pix); i += 4 {
		pixels[i/4] = []uint8{
			img.Pix[i],
			img.Pix[i+1],
			img.Pix[i+2],
			img.Pix[i+3],
		}
	}
	return pixels
}

func writeImageToDisk(imageFormat string, image *image.RGBA, outFilename string) error {
	outFile, err := os.OpenFile(outFilename, os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		return err
	}

	defer outFile.Close()

	if imageFormat == "gif" {
		err = gif.Encode(outFile, image, &gif.Options{NumColors: 256})
	} else {
		err = png.Encode(outFile, image)
	}

	if err != nil {
		return err
	}

	return nil
}

func extractImageData(img *image.RGBA) *imageData {
	bounds := img.Bounds()

	return &imageData{
		bounds,
		bounds.Dx(),
		bounds.Dy(),
		getImagePixels(img),
	}
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

func getRandomCircularMaskWithinImage(sourceImageData *imageData) *shape.Circle {
	// Get random coordinates to draw at
	center := getRandomPoint(sourceImageData)

	r := getRandomRadius()
	return &shape.Circle{Center: center, Radius: r}
}
