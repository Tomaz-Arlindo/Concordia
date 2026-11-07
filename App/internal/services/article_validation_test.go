package services

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestValidatePublicationDate(t *testing.T) {
	for _, date := range []string{"31/02/2026", "2026-02-31", "2026-2-01", "2999-01-01"} {
		if err := validatePublicationDate(date); err != ErrInvalidPublicationDate {
			t.Errorf("data inválida aceita: %q", date)
		}
	}
	if err := validatePublicationDate("2020-02-29"); err != nil {
		t.Fatalf("data válida recusada: %v", err)
	}
	if err := validatePublicationDate(""); err != nil {
		t.Fatalf("data vazia deveria ser opcional: %v", err)
	}
}

func TestImportPDFRejectsMissingFieldsAndInvalidFiles(t *testing.T) {
	tempDir := t.TempDir()
	validFields := ImportArticleInput{
		Title:        "TESTE artigo",
		SourceID:     1,
		OriginalPath: filepath.Join(tempDir, "TESTE_artigo.pdf"),
		UserID:       1,
	}
	if err := os.WriteFile(validFields.OriginalPath, []byte("%PDF-1.4 TESTE conteúdo"), 0o600); err != nil {
		t.Fatal(err)
	}

	textPath := filepath.Join(tempDir, "TESTE_artigo.txt")
	if err := os.WriteFile(textPath, []byte("%PDF-1.4 TESTE conteúdo"), 0o600); err != nil {
		t.Fatal(err)
	}
	invalidHeaderPath := filepath.Join(tempDir, "TESTE_invalido.pdf")
	if err := os.WriteFile(invalidHeaderPath, []byte("TESTE não é PDF"), 0o600); err != nil {
		t.Fatal(err)
	}

	service := NewArticleService(nil, nil, nil)
	missingTitle := validFields
	missingTitle.Title = ""
	missingSource := validFields
	missingSource.SourceID = 0
	missingPDF := validFields
	missingPDF.OriginalPath = ""
	missingUser := validFields
	missingUser.UserID = 0
	invalidExtension := validFields
	invalidExtension.OriginalPath = textPath
	invalidHeader := validFields
	invalidHeader.OriginalPath = invalidHeaderPath
	tests := []struct {
		name  string
		input ImportArticleInput
	}{
		{name: "título ausente", input: missingTitle},
		{name: "fonte ausente", input: missingSource},
		{name: "PDF ausente", input: missingPDF},
		{name: "usuário ausente", input: missingUser},
		{name: "extensão inválida", input: invalidExtension},
		{name: "cabeçalho PDF inválido", input: invalidHeader},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := service.ImportPDF(context.Background(), tt.input); !errors.Is(err, ErrInvalidArticle) {
				t.Fatalf("esperado ErrInvalidArticle, obtido %v", err)
			}
		})
	}
}

func TestInspectPDFReturnsChecksumAndSize(t *testing.T) {
	content := []byte("%PDF-1.4 TESTE conteúdo")
	path := filepath.Join(t.TempDir(), "TESTE_valido.pdf")
	if err := os.WriteFile(path, content, 0o600); err != nil {
		t.Fatal(err)
	}

	checksum, size, err := inspectPDF(path)
	if err != nil {
		t.Fatalf("inspectPDF falhou para cabeçalho válido: %v", err)
	}
	expected := sha256.Sum256(content)
	if checksum != hex.EncodeToString(expected[:]) || size != int64(len(content)) {
		t.Fatalf("resultado inesperado: checksum=%q size=%d", checksum, size)
	}
}

func TestCopyFileCopiesPDFContents(t *testing.T) {
	content := []byte("%PDF-1.4 TESTE cópia")
	source := filepath.Join(t.TempDir(), "TESTE_origem.pdf")
	destination := filepath.Join(t.TempDir(), "TESTE_destino.pdf")
	if err := os.WriteFile(source, content, 0o600); err != nil {
		t.Fatal(err)
	}

	if err := copyFile(source, destination); err != nil {
		t.Fatalf("copyFile falhou: %v", err)
	}
	got, err := os.ReadFile(destination)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(content) {
		t.Fatalf("conteúdo copiado divergente: %q", got)
	}
}

func TestParseAuthorsTrimsDeduplicatesAndSorts(t *testing.T) {
	got := parseAuthors(" Zed, alice, ALICE, , Bob ")
	want := []string{"alice", "Bob", "Zed"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("parseAuthors() = %#v; esperado %#v", got, want)
	}
}

func TestUpdateArticleRejectsMissingRequiredFields(t *testing.T) {
	service := NewArticleService(nil, nil, nil)
	tests := []struct {
		name  string
		input UpdateArticleInput
	}{
		{name: "id ausente", input: UpdateArticleInput{Title: "TESTE artigo", SourceID: 1}},
		{name: "título ausente", input: UpdateArticleInput{ID: 1, SourceID: 1}},
		{name: "fonte ausente", input: UpdateArticleInput{ID: 1, Title: "TESTE artigo"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := service.UpdateArticle(context.Background(), tt.input); !errors.Is(err, ErrInvalidArticle) {
				t.Fatalf("esperado ErrInvalidArticle, obtido %v", err)
			}
		})
	}
}
