package indigoo

import (
	"strings"
)

func getJavaScriptCode(tmpl string) (*string, error) {
	result, err := getStringBetween(tmpl, "<script>", "</script>")
	if err != nil {
		return nil, err
	}

	result = strings.TrimSpace(result)

	return &result, nil
}
