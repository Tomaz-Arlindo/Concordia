package services

import (
	"context"
	"errors"
	"os"
	"strings"

	pdfextract "github.com/giraffesyo/pdf"
)

var ErrNoExtractableText = errors.New("PDF sem texto extraível")

func ExtractPDFText(ctx context.Context, path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return "", err
	}

	doc, err := pdfextract.Extract(ctx, file, info.Size())
	if err != nil {
		return "", err
	}

	text := strings.TrimSpace(doc.Text())
	if text == "" {
		return "", ErrNoExtractableText
	}
	return text, nil
}
