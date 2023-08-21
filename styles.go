package indigoo

import "strings"

func getCSSCode(tmpl string) (*string, error) {
	result, err := getStringBetween(tmpl, "<style>", "</style>")
	if err != nil {
		return nil, err
	}

	result = strings.TrimSpace(result)

	return &result, nil
}

//func handleStyles(tmpl string, regionName string) (string, error) {
//	// get the root elements that start with a dom element and add a class to it
//	domElements := regexp.MustCompile("<(\\w+)").FindAllStringSubmatch(tmpl, -1)
//
//	for _, element := range domElements {
//		log.Println(element)
//	}
//
//	if regexp.MustCompile("<style>").MatchString(tmpl) {
//		tmpl = regexp.MustCompile("<style>").ReplaceAllString(tmpl, "{{end}}\n<style>")
//		tmpl = regexp.MustCompile("<style>").ReplaceAllString(tmpl, fmt.Sprintf("{{define \"css%s\"}}\n<style>\n", regionName))
//		tmpl = regexp.MustCompile("</style>").ReplaceAllString(tmpl, "</style>\n{{end}}\n")
//	} else {
//		tmpl += "{{end}}"
//	}
//
//	return tmpl, nil
//}
