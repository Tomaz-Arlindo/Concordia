package services

import (
	"fmt"
	"regexp"
	"strings"
)

const maxExportNameRunes = 150

var (
	exportSpaces = regexp.MustCompile(`\s+`)

	exportRemoveChars = strings.NewReplacer(`"`, "", "*", "", "?", "", "<", "", ">", "")
	exportDashChars   = strings.NewReplacer("/", "-", `\`, "-", "|", "-")

	// nomes reservados do windows
	exportReservedNames = map[string]bool{
		"CON": true, "PRN": true, "AUX": true, "NUL": true,
		"COM1": true, "COM2": true, "COM3": true, "COM4": true, "COM5": true,
		"COM6": true, "COM7": true, "COM8": true, "COM9": true,
		"LPT1": true, "LPT2": true, "LPT3": true, "LPT4": true, "LPT5": true,
		"LPT6": true, "LPT7": true, "LPT8": true, "LPT9": true,
	}
)

func ExportPDFFileName(title string, id uint64) string {
	name := strings.ReplaceAll(title, ":", " -")
	name = exportDashChars.Replace(name)
	name = exportRemoveChars.Replace(name)
	name = strings.Map(func(r rune) rune {
		if r < 32 {
			return -1
		}
		return r
	}, name)
	name = exportSpaces.ReplaceAllString(name, " ")
	name = strings.TrimSpace(name)

	if runes := []rune(name); len(runes) > maxExportNameRunes {
		name = strings.TrimSpace(string(runes[:maxExportNameRunes]))
	}

	name = strings.TrimRight(name, ". ")

	if name == "" {
		name = fmt.Sprintf("artigo-%d", id)
	}
	if exportReservedNames[strings.ToUpper(name)] {
		name += "_"
	}

	return name + ".pdf"
}