// Package MedianFilter implements a simple library for image operations.
// The library can work with pngs or jpgs. Same functions can be
// used for both of those image types.
package MedianFilter

import (
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"image/png"
	"os"
	"sort"
	"strings"
	"sync"
)

// Pixel is a single pixel in 2d array
type Pixel struct {
	R int
	G int
	B int
	A int
}

// Image is the main object that holds information about the
// image file. Also is a wrapper around the decoded image
// from the standard image library.
type Image struct {
	Pixels [][]Pixel
	Width  int
	Height int
	_Rect  image.Rectangle
	_Image image.Image
}

// set pixel value with key name and new value
func (pix *Pixel) set(keyName string, val int) Pixel {
	switch keyName {
	case "R":
		pix.R = val
	case "G":
		pix.G = val
	case "B":
		pix.B = val
	case "A":
		pix.A = val
	}
	return *pix
}

// rgbaToPixel alpha-premultiplied red, green, blue and alpha values
// to 8 bit red, green, blue and alpha values.
func rgbaToPixel(r uint32, g uint32, b uint32, a uint32) Pixel {
	return Pixel{
		R: int(r / 257),
		G: int(g / 257),
		B: int(b / 257),
		A: int(a / 257),
	}
}

// newImage reads an image from the given file path and return a
// new `Image` struct.
func newImage(filePath string) (*Image, error) {
	s := strings.Split(filePath, ".")
	imgType := s[len(s)-1]

	switch imgType {
	case "jpeg", "jpg":
		image.RegisterFormat("jpeg", "jpeg", jpeg.Decode, jpeg.DecodeConfig)
	case "png":
		image.RegisterFormat("png", "png", png.Decode, png.DecodeConfig)
	default:
		return nil, errors.New("unknown image type")
	}

	imgReader, err := os.Open(filePath)
	if err != nil {
		fmt.Println("error opening")
		return nil, err
	}

	img, _, err := image.Decode(imgReader)
	if err != nil {
		fmt.Println("error decoding")
		return nil, err
	}

	bounds := img.Bounds()
	width, height := bounds.Max.X, bounds.Max.Y

	var pixels [][]Pixel
	for y := 0; y < height; y++ {
		var row []Pixel
		for x := 0; x < width; x++ {
			pixel := rgbaToPixel(img.At(x, y).RGBA())
			row = append(row, pixel)
		}
		pixels = append(pixels, row)
	}

	return &Image{
		Pixels: pixels,
		Width:  width,
		Height: height,
		_Rect:  img.Bounds(),
		_Image: img,
	}, nil
}

// writeToFile writes iamges to the given filepath.
// Returns an error if it occurs.
func (img *Image) writeToFile(outputPath string) error {
	cimg := image.NewRGBA(img._Rect)
	draw.Draw(cimg, img._Rect, img._Image, image.Point{}, draw.Over)

	for y := 0; y < img.Height; y++ {
		for x := 0; x < img.Width; x++ {
			rowIndex, colIndex := y, x
			pixel := img.Pixels[rowIndex][colIndex]
			cimg.Set(x, y, color.RGBA{
				uint8(pixel.R),
				uint8(pixel.G),
				uint8(pixel.B),
				uint8(pixel.A),
			})
		}
	}

	s := strings.Split(outputPath, ".")
	imgType := s[len(s)-1]

	switch imgType {
	case "jpeg", "jpg", "png":
		fd, err := os.Create(outputPath)
		if err != nil {
			return err
		}

		switch imgType {
		case "jpeg", "jpg":
			jpeg.Encode(fd, cimg, nil)
		case "png":
			png.Encode(fd, cimg)
		}
	default:
		return errors.New("unknown image type")
	}

	return nil
}

// medianPixel finds the median r, g, b values from the given
// pixel array and creates a new pixel from that median values
func medianPixel(pixels []Pixel) Pixel {
	var (
		rValues []int
		gValues []int
		bValues []int
	)

	for _, pix := range pixels {
		rValues = append(rValues, pix.R)
		gValues = append(gValues, pix.G)
		bValues = append(bValues, pix.B)
	}

	sort.Ints(rValues)
	sort.Ints(gValues)
	sort.Ints(bValues)

	rMedian := rValues[int(len(rValues)/2)]
	gMedian := gValues[int(len(gValues)/2)]
	bMedian := bValues[int(len(bValues)/2)]

	return Pixel{rMedian, gMedian, bMedian, 0}
}

// medianFilter iterates the given filepaths and generates new image
// objects. It then checks to see if all the heights and the widths
// of the images are matching. If they are, each pixel of every image is
// iterated and a median filter is applied to given images. Returns the
// output image object and an error if there is any.
func medianFilter(filePaths []string) (*Image, error) {
	var images []*Image
	for _, filePath := range filePaths {
		img, err := newImage(filePath)
		if err != nil {
			return nil, err
		}
		images = append(images, img)
	}

	if len(images) < 5 {
		return nil, errors.New("not enough images to perform noise reduction")
	}

	outputImage := images[0]

	heigth := outputImage.Height
	width := outputImage.Width
	for _, img := range images {
		if heigth != img.Height || width != img.Width {
			return nil, errors.New("at least one image has a different width or height")
		}
	}

	var wg sync.WaitGroup

	for rowIndex := 0; rowIndex < heigth; rowIndex++ {
		wg.Add(1)

		go (func(rowIndex int) {
			for colIndex := 0; colIndex < width; colIndex++ {
				var pixels []Pixel
				for _, img := range images {
					pixels = append(pixels, img.Pixels[rowIndex][colIndex])
				}

				medPixel := medianPixel(pixels)
				outputImage.Pixels[rowIndex][colIndex].set("R", medPixel.R)
				outputImage.Pixels[rowIndex][colIndex].set("G", medPixel.G)
				outputImage.Pixels[rowIndex][colIndex].set("B", medPixel.B)
			}
			wg.Done()
		})(rowIndex)
	}

	wg.Wait()
	return outputImage, nil
}

// RemoveMovingObjs iterates the given filepaths and generates new image
// image that does not have the moving objects in the given images.
func RemoveMovingObjs(filepaths []string, outputPath string) error {
	img, err := medianFilter(filepaths)
	if err != nil {
		return err
	}
	img.writeToFile(outputPath)
	return nil
}
