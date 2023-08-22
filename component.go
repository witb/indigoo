package indigoo

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
)

type component struct {
	Name             string
	Path             string
	Template         string
	Styles           string
	Script           string
	Content          string
	RootDomComponent []string
	CustomClass      string
	Raw              *string
	Components       map[string]*component
}

func (cmpt *component) New(path string, name string) (*component, error) {
	err := cmpt.handleComponentGeneration(path, name)
	if err != nil {
		return nil, err
	}

	return cmpt, nil
}

func (cmpt *component) handleComponentGeneration(path string, name string) error {
	var err error

	cmpt.Raw, err = readFile(path)
	if err != nil {
		return err
	}

	cmpt.Name = name
	cmpt.Path = path
	cmpt.CustomClass = "indigoo-" + generateCustomString(8)

	err = cmpt.handleTemplateSectors()
	if err != nil {
		return err
	}

	err = cmpt.handleComponents()
	if err != nil {
		return err
	}

	for _, component := range cmpt.Components {
		for _, subComponent := range component.Components {
			cmpt.Components[subComponent.Path] = subComponent
		}
	}

	cmpt.parseTemplate()

	return nil
}

func (cmpt *component) handleComponents() error {
	cps := map[string]*component{}

	scriptLines := strings.Split(cmpt.Script, "\n")
	var newTemplateLines []string
	for _, line := range scriptLines {
		match := componentImportPattern.FindStringSubmatch(line)

		if len(match) > 0 {
			componentNamePattern := regexp.MustCompile(`(\s|{|,)[A-Z]+[^.,\s,}]+`)

			componentMatches := componentNamePattern.FindAllString(line, -1)

			if len(componentMatches) > 0 {
				var componentNames []string
				for _, componentMatch := range componentMatches {
					componentNames = append(componentNames, componentMatch[1:])
				}

				componentPath := filepath.Join(filepath.Dir(cmpt.Path), match[1])

				componentData, err := new(component).New(componentPath, componentNames[0])
				if err != nil {
					return err
				}

				cps[componentPath] = componentData
			}
		} else {
			newTemplateLines = append(newTemplateLines, line)
		}
	}

	cmpt.Components = cps

	return nil
}

func (cmpt *component) handleTemplateSectors() error {
	result, err := getStringBetween(*cmpt.Raw, "<script>", "</script>")
	if err != nil {
		return err
	}

	cmpt.Script = strings.TrimSpace(result)

	result, err = getStringBetween(*cmpt.Raw, "<style>", "</style>")
	if err != nil {
		return err
	}

	cmpt.Styles = strings.TrimSpace(result)

	result, err = getStringBetween(*cmpt.Raw, "<component>", "</component>")
	if err != nil {
		return err
	}

	cmpt.Content = strings.TrimSpace(result)

	cmpt.handleCssOnRootDomElements()

	return nil
}

func (cmpt *component) parseTemplate() {
	var tmpl string

	if cmpt.Script != "" {
		tmpl += "{{define \"js-" + cmpt.CustomClass + "\"}}\n<script>\n" + cmpt.RemoveImports() + "\n</script>\n{{end}}\n"
	}

	if cmpt.Content != "" {
		tmpl += "{{define \"content-" + cmpt.CustomClass + "\"}}\n" + cmpt.Content + "\n{{end}}\n"
	}

	if cmpt.Styles != "" {
		tmpl += "{{define \"css-" + cmpt.CustomClass + "\"}}\n<style>\n" + cmpt.Styles + "\n</style>\n{{end}}\n"
	}

	cmpt.Template = tmpl
}

func (cmpt *component) handleCssOnRootDomElements() error {
	hasCssSelectorWithTag := func(tag string, class string) bool {
		selectorPattern := fmt.Sprintf(`(^|\s)%s\s`, tag)
		re := regexp.MustCompile(selectorPattern)

		hasMatch := re.MatchString(cmpt.Styles)

		cmpt.Styles = re.ReplaceAllStringFunc(cmpt.Styles, func(match string) string {
			return strings.TrimSpace(match) + "." + class
		})

		return hasMatch
	}

	pattern := fmt.Sprintf(`<([a-z]{1,10})[^>]*>([\s\S]*?)<\/([a-z]{1,10})>|<([a-z]{1,10})(?:\s[^>]*)?\s*\/>`)

	cmpt.Content = regexp.MustCompile(pattern).ReplaceAllStringFunc(cmpt.Content, func(match string) string {
		// TODO: make it work with components that already have class
		openingTag := strings.Index(match, "<")
		closingTag := strings.Index(match, ">")

		if openingTag != -1 || closingTag != -1 {
			firstTag := match[openingTag : closingTag+1]

			re := regexp.MustCompile(`<([a-zA-Z0-9]+)[\s/>]`)
			tagMatch := re.FindStringSubmatch(firstTag)
			hasTag := hasCssSelectorWithTag(tagMatch[1], cmpt.CustomClass)

			if hasTag {
				classPattern := "class=\""
				if strings.Contains(firstTag, classPattern) {
					match = strings.Replace(firstTag, classPattern, fmt.Sprintf(`class="%s `, cmpt.CustomClass), 1) + match[closingTag+1:]
				} else {
					match = strings.Replace(firstTag, ">", fmt.Sprintf(` class="%s">`, cmpt.CustomClass), 1) + match[closingTag+1:]
				}
			}
		}

		return match
	})

	return nil
}

func (cmpt *component) handleComponentsTemplateChange() {
	handleTemplateReplacement := func(cmpt *component, componentClass string, tag *regexp.Regexp) string {
		return tag.ReplaceAllStringFunc(cmpt.Template, func(match string) string {
			matches := tag.FindStringSubmatch(match)

			for _, match := range matches {
				templateRef := "{{template \"content-" + componentClass + "\" .}}"
				return strings.Replace(match, match, templateRef, 1)
			}

			return match
		})
	}

	for _, component := range cmpt.Components {
		componentTagPattern := regexp.MustCompile(`<` + component.Name + `\s*(.*?)<\/` + component.Name + `>|<` + component.Name + `\s*(.*?)\/>`)

		cmpt.Template = handleTemplateReplacement(cmpt, component.CustomClass, componentTagPattern)

		for _, cmptToReplace := range cmpt.Components {
			if cmptToReplace.CustomClass != component.CustomClass {
				cmptToReplace.Template = handleTemplateReplacement(cmptToReplace, component.CustomClass, componentTagPattern)
			}
		}

		cmpt.Template += component.Template
	}
}

func (cmpt *component) RemoveImports() string {
	var cleanedLines []string
	lines := strings.Split(cmpt.Script, "\n")

	for _, line := range lines {
		if componentImportPattern.MatchString(strings.TrimSpace(line)) {
			cleanedLines = append(cleanedLines, "")
		} else {
			cleanedLines = append(cleanedLines, line)
		}
	}

	cleanedScript := strings.Join(cleanedLines, "\n")

	return cleanedScript
}
