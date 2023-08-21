package indigoo

import (
	"log"
	"strings"
)

type component struct {
	Name        string
	Path        string
	Template    string
	Styles      *string
	Script      *string
	Content     *string
	CustomClass string
	Raw         *string
	Components  map[string]*component
}

func (cmpt *component) RemoveImports() *string {
	var cleanedLines []string
	lines := strings.Split(*cmpt.Script, "\n")

	for _, line := range lines {
		log.Println(line)
		log.Println(strings.TrimSpace(line))
		log.Println(componentImportPattern.MatchString(line))
		log.Println(componentImportPattern.MatchString(strings.TrimSpace(line)))

		if componentImportPattern.MatchString(strings.TrimSpace(line)) {
			cleanedLines = append(cleanedLines, "")
		} else {
			cleanedLines = append(cleanedLines, line)
		}
	}

	cleanedScript := strings.Join(cleanedLines, "\n")

	return &cleanedScript
}
