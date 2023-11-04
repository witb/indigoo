package indigoo

import (
	"fmt"
	"github.com/witb/indigoo/utils"
	"path/filepath"
	"regexp"
	"strings"
)

type component struct {
	Name             string
	WebComponentName string
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
	cmpt.WebComponentName = strings.ToLower(path)
	cmpt.WebComponentName = strings.TrimRight(cmpt.WebComponentName, ".goo")
	cmpt.WebComponentName = "goo-" + strings.ReplaceAll(strings.ReplaceAll(cmpt.WebComponentName, "app/", ""), "/", "-")
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

	return nil
}

func (cmpt *component) parseTemplate() {
	var tmpl string

	if cmpt.Script != "" {
		tmpl += "{{define \"js-" + cmpt.CustomClass + "\"}}\n<script>\n" + cmpt.handleJavascript() + "\n</script>\n{{end}}\n"
	}

	if cmpt.Content != "" {
		openWebComponent := ""
		closeWebComponent := ""

		if cmpt.Script != "" {
			openWebComponent = "<" + cmpt.WebComponentName + ">\n"
			closeWebComponent = "</" + cmpt.WebComponentName + ">\n"
		}

		tmpl += fmt.Sprintf("{{define \"content-%s\"}}\n%s%s\n%s\n{{end}}\n", cmpt.CustomClass, openWebComponent, strings.TrimSpace(cmpt.Content), closeWebComponent)
	}

	cmpt.Template = tmpl
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

func (cmpt *component) handleJavascript() string {
	var cleanedLines []string
	lines := strings.Split(cmpt.Script, "\n")
	componentName := utils.ToCamel(cmpt.WebComponentName)
	startWebComponent := fmt.Sprintf("class %s extends HTMLElement {\nstatic observedAttributes = [\"name\"];\n\nconstructor() {\nsuper();\n\n", componentName)
	attachCSS := ""

	if cmpt.Styles != "" {
		//attachCSS = "\nconnectedCallback() {\nconst shadow = this.attachShadow({ mode: \"open\" });\n\nconst style = document.createElement(\"style\");\n\n"
		//attachCSS += "console.log(this.innerHtml);"
		//attachCSS += fmt.Sprintf("style.textContent = `%s`;\n\nshadow.appendChild(style);\n}", cmpt.Styles)
	}

	endWebComponent := fmt.Sprintf("\n}%s\n}\n\ncustomElements.define(\"%s\", %s);", attachCSS, cmpt.WebComponentName, componentName)

	for _, line := range lines {
		if componentImportPattern.MatchString(strings.TrimSpace(line)) {
			cleanedLines = append(cleanedLines, "")
		} else {
			cleanedLines = append(cleanedLines, line)
		}
	}

	cleanedScript := startWebComponent + strings.TrimSpace(strings.Join(cleanedLines, "\n")) + endWebComponent

	return cleanedScript
}
