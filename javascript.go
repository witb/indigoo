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

//func handleJavascript(tmpl string, regionName string) (string, error) {
//	// get the root elements that start with a dom element and add a class to it
//	domElements := regexp.MustCompile("<(\\w+)").FindAllStringSubmatch(tmpl, -1)
//
//	for _, element := range domElements {
//		log.Println(element)
//	}
//
//	if regexp.MustCompile("<script>").MatchString(tmpl) {
//		tmpl = regexp.MustCompile("<script>").ReplaceAllString(tmpl, fmt.Sprintf("{{define \"js%s\"}}\n<script>\n", regionName))
//		tmpl = regexp.MustCompile("</script>").ReplaceAllString(tmpl, fmt.Sprintf("</script>\n{{define \"components/%s\"}}\n", regionName))
//		tmpl = regexp.MustCompile("</script>").ReplaceAllString(tmpl, "</script>\n{{end}}\n")
//	} else {
//		tmpl = fmt.Sprintf("{{define \"components/%s\"}}\n", regionName) + tmpl
//	}
//
//	return tmpl, nil
//}
