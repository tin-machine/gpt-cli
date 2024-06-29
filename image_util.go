package main

import (
	"encoding/base64"
	"fmt"
	"image"
	"image/jpeg"
	"os"
	"path/filepath"
)

func imageToBase64(imagePath string) (string, string, error) {
	file, err := os.Open(imagePath)
	if err != nil {
		return "", "", err
	}
	defer file.Close()

	fileInfo, _ := file.Stat()
	var size = fileInfo.Size()
	buf := make([]byte, size)
	file.Read(buf)

	base64Image := base64.StdEncoding.EncodeToString(buf)
	ext := filepath.Ext(imagePath)

	var mimeType string
	switch ext {
	case ".jpg", ".jpeg":
		mimeType = "image/jpeg"
	case ".png":
		mimeType = "image/png"
	default:
		return "", "", fmt.Errorf("unsupported image format: %s", ext)
	}

	return base64Image, mimeType, nil
}

func saveJPG(img image.Image, path string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	opts := &jpeg.Options{Quality: 80}
	err = jpeg.Encode(file, img, opts)
	if err != nil {
		return err
	}

	return nil
}
