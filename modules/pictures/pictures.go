package pictures

import (
	"bytes"
	log "chat-app/modules/logging"
	"chat-app/modules/macros"
	"image"
	"image/jpeg"
	"time"

	"github.com/disintegration/imaging"
)

func CheckAvatarPic(imgBytes *[]byte, userID uint64) string {
	start := time.Now().UnixMilli()
	// decode
	img, _, err := image.Decode(bytes.NewReader(*imgBytes))
	if err != nil {
		log.Error("%s", err.Error())
		log.Hack("Received avatar pic from user ID [%d] is not a picture", userID)
		return "Not a picture"
	}

	// check if picture is too small
	if img.Bounds().Dx() < 64 || img.Bounds().Dy() < 64 {
		log.Trace("Received avatar pic from user ID [%d] is too small", userID)
		return "Picture is too small, minimum 64x64"
	}

	// check if picture is either too wide or too tall
	widthRatio := float64(img.Bounds().Dx()) / float64(img.Bounds().Dy())
	heightRatio := float64(img.Bounds().Dy()) / float64(img.Bounds().Dx())
	if widthRatio > 2 {
		log.Trace("Received avatar pic from user ID [%d] is too wide", userID)
		return "Picture is too wide, must be less than 1:2 ratio"
	} else if heightRatio > 2 {
		log.Trace("Received avatar pic from user ID [%d] is too tall", userID)
		return "Picture is too tall, must be less than 1:2 ratio"
	}

	// if height is larger than width, crop height to same size as width,
	// else if width is larger than height, crop width to the same size as height
	if img.Bounds().Dy() > img.Bounds().Dx() {
		img = imaging.CropCenter(img, img.Bounds().Dx(), img.Bounds().Dx())
	} else if img.Bounds().Dx() > img.Bounds().Dy() {
		img = imaging.CropCenter(img, img.Bounds().Dy(), img.Bounds().Dy())
	}

	// check if picture is in square dimension
	if img.Bounds().Dx() != img.Bounds().Dy() {
		log.Impossible("Avatar pic of user ID [%d] cropped isnt in square dimension: [%dx%d]", userID, img.Bounds().Dx(), img.Bounds().Dy())
		return ""
	}

	// resize to 256px width if wider
	if img.Bounds().Dx() > 256 && img.Bounds().Dy() > 256 {
		img = imaging.Resize(img, 256, 256, imaging.Lanczos)
	}

	// recompress into jpg
	var buf bytes.Buffer
	err = jpeg.Encode(&buf, img, &jpeg.Options{Quality: 90})
	if err != nil {
		log.FatalError(err.Error(), "Error compressing avatar pic from user ID [%d]", userID)
		return ""
	}

	macros.MeasureTime(start, "checking avatar pic")
	*imgBytes = buf.Bytes()
	return ""
}
