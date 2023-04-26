package epubimage

import (
	"image"

	epubfilters "github.com/celogeek/go-comic-converter/v2/internal/epub/filters"
	"github.com/disintegration/gift"
)

// create filter to apply to the source
func NewGift(img image.Image, options *Options) *gift.GIFT {
	g := gift.New()
	g.SetParallelization(false)

	if options.Crop {
		g.Add(epubfilters.AutoCrop(
			img,
			options.CropRatioLeft,
			options.CropRatioUp,
			options.CropRatioRight,
			options.CropRatioBottom,
		))
	}
	if options.AutoRotate && img.Bounds().Dx() > img.Bounds().Dy() {
		g.Add(gift.Rotate90())
	}

	if options.Contrast != 0 {
		g.Add(gift.Contrast(float32(options.Contrast)))
	}

	if options.Brightness != 0 {
		g.Add(gift.Brightness(float32(options.Brightness)))
	}

	g.Add(
		epubfilters.Resize(options.ViewWidth, options.ViewHeight, gift.LanczosResampling),
		epubfilters.Pixel(),
	)
	return g
}

// create filters to cut image into 2 equal pieces
func NewGiftSplitDoublePage(options *Options) []*gift.GIFT {
	gifts := make([]*gift.GIFT, 2)

	gifts[0] = gift.New(
		epubfilters.CropSplitDoublePage(options.Manga),
	)

	gifts[1] = gift.New(
		epubfilters.CropSplitDoublePage(!options.Manga),
	)

	for _, g := range gifts {
		if options.Contrast != 0 {
			g.Add(gift.Contrast(float32(options.Contrast)))
		}
		if options.Brightness != 0 {
			g.Add(gift.Brightness(float32(options.Brightness)))
		}

		g.Add(
			epubfilters.Resize(options.ViewWidth, options.ViewHeight, gift.LanczosResampling),
		)
	}

	return gifts
}
