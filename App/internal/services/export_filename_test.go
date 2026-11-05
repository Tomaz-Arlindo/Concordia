package services

import (
	"strings"
	"testing"
)

func TestExportPDFFileNameKeepsTitle(t *testing.T) {
	got := ExportPDFFileName("Teste artigo nome teste", 1)
	want := "Teste artigo nome teste.pdf"
	if got != want {
		t.Fatalf("esperado %q, obtido %q", want, got)
	}
}

func TestExportPDFFileNameReplacesInvalidCharacters(t *testing.T) {
	cases := []struct {
		title string
		want  string
	}{
		{"Título: subtítulo", "Título - subtítulo.pdf"},
		{`Teoria/Prática\Aplicada|Final`, "Teoria-Prática-Aplicada-Final.pdf"},
		{`O que é "isso"? <teste>*`, "O que é isso teste.pdf"},
		{"  muitos   espaços  ", "muitos espaços.pdf"},
		{"Título final...", "Título final.pdf"},
	}

	for _, c := range cases {
		if got := ExportPDFFileName(c.title, 1); got != c.want {
			t.Errorf("título %q: esperado %q, obtido %q", c.title, c.want, got)
		}
	}
}

func TestExportPDFFileNameFallbacks(t *testing.T) {
	if got := ExportPDFFileName("", 7); got != "artigo-7.pdf" {
		t.Errorf("título vazio: esperado %q, obtido %q", "artigo-7.pdf", got)
	}
	if got := ExportPDFFileName("???", 9); got != "artigo-9.pdf" {
		t.Errorf("título só com caracteres inválidos: esperado %q, obtido %q", "artigo-9.pdf", got)
	}
	if got := ExportPDFFileName("CON", 1); got != "CON_.pdf" {
		t.Errorf("nome reservado: esperado %q, obtido %q", "CON_.pdf", got)
	}
}

func TestExportPDFFileNameLimitsLength(t *testing.T) {
	got := ExportPDFFileName(strings.Repeat("a", 400), 1)
	if want := maxExportNameRunes + len(".pdf"); len([]rune(got)) != want {
		t.Fatalf("esperado %d caracteres, obtido %d", want, len([]rune(got)))
	}
	if !strings.HasSuffix(got, ".pdf") {
		t.Fatalf("o nome deveria terminar em .pdf: %q", got)
	}
}