package img

import (
	"image"
	"os"
)

// Operation ...
type Operation interface {
	Crop(img image.Image, w, h int) (image.Image, error)
	Open(filename string) (image.Image, error)
	Save(img image.Image, filename string) error
	CropAndSave(src, dst string, w, h int) error
}

// NewOperation ...
func NewOperation() Operation {
	return &Imaging{}
}

func ValidPosterProportion(imgPath string) bool {
	if reader, err := os.Open(imgPath); err == nil {
		defer reader.Close()
		im, _, err := image.Decode(reader)
		if err != nil {
			return false
		}

		width := im.Bounds().Dx()
		height := im.Bounds().Dy()

		if width == 0 || height == 0 {
			return false
		}
		if width > height {
			return false
		}
		return (1.3 < float64(height)/float64(width) && float64(height)/float64(width) < 1.6)
	}
	return false
}
