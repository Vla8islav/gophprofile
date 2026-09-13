package worker

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"image/color"
	"image/jpeg"

	// decoders for the three upload formats
	_ "image/jpeg"
	_ "image/png"

	_ "golang.org/x/image/webp"

	"github.com/Vla8islav/gophprofile/internal/broker"
	"github.com/Vla8islav/gophprofile/internal/domain"
	"golang.org/x/image/draw"
)

const jpegQuality = 85

// generateThumbnails downloads the original once and renders every configured size
func (w *Worker) generateThumbnails(ctx context.Context, avatar *domain.Avatar) error {
	original, err := w.fileStorage.Download(ctx, avatar.S3Key)
	if err != nil {
		return fmt.Errorf("download original: %w", err)
	}
	defer func() { _ = original.Close() }()

	src, _, err := image.Decode(original)
	if err != nil {
		return broker.Permanent(fmt.Errorf("decode original: %w", err))
	}

	keys := make(map[string]string, len(domain.ThumbnailSizes))
	for size, edge := range domain.ThumbnailSizes {
		thumbnail, err := renderThumbnail(src, edge)
		if err != nil {
			return broker.Permanent(fmt.Errorf("render %s: %w", size, err))
		}
		key := domain.ThumbnailS3Key(avatar.ID, size)
		err = w.fileStorage.Upload(ctx, key, "image/jpeg", int64(thumbnail.Len()), thumbnail)
		if err != nil {
			return fmt.Errorf("upload %s: %w", size, err)
		}
		keys[size] = key
	}

	if err := w.repository.SetAvatarThumbnails(ctx, avatar.ID, keys); err != nil {
		return fmt.Errorf("record thumbnails: %w", err)
	}
	return nil
}

// renderThumbnail produces a square JPEG: center-crop to square
func renderThumbnail(src image.Image, edge int) (*bytes.Buffer, error) {
	square := centerSquare(src.Bounds())

	dst := image.NewRGBA(image.Rect(0, 0, edge, edge))
	// white background first, so transparent PNG/WebP pixels don't go black
	draw.Draw(dst, dst.Bounds(), image.NewUniform(color.White), image.Point{}, draw.Src)
	draw.CatmullRom.Scale(dst, dst.Bounds(), src, square, draw.Over, nil)

	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, dst, &jpeg.Options{Quality: jpegQuality}); err != nil {
		return nil, err
	}
	return &buf, nil
}

// centerSquare is the largest centered square inside bounds.
func centerSquare(bounds image.Rectangle) image.Rectangle {
	width, height := bounds.Dx(), bounds.Dy()
	edge := min(width, height)
	x0 := bounds.Min.X + (width-edge)/2
	y0 := bounds.Min.Y + (height-edge)/2
	return image.Rect(x0, y0, x0+edge, y0+edge)
}
