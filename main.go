package indigoo

import (
	"errors"
	"fmt"
	"github.com/go-chi/chi/v5"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var Cache = true
var appFolder string
var baseTemplate string
var entryPage string
var templateCache = map[string]*template.Template{}
var componentImportPattern = regexp.MustCompile(`import\s+\w+\s+from\s+['"]([^'"]+)['"]`)

func init() {
	err := validateStructuralFiles()
	if err != nil {
		log.Fatal(err)
	}
}

func RenderApplication() *chi.Mux {
	mux := chi.NewRouter()

	RenderApplicationWithMux(mux)

	return mux
}

func RenderApplicationWithMux(mux *chi.Mux) *chi.Mux {
	renderRoutesFromFolderStructure(mux)

	return mux
}

func renderRoutesFromFolderStructure(mux *chi.Mux) *chi.Mux {
	err := filepath.Walk(appFolder, func(path string, info os.FileInfo, err error) error {
		if err == nil && !info.IsDir() && info.Name() == "page.goo" {
			relativePath := path[len(appFolder)+1:]
			pagePath := "/" + relativePath[:len(relativePath)-len(info.Name())]

			if len(pagePath) > 1 && pagePath[len(pagePath)-1:] == "/" {
				pagePath = pagePath[:len(pagePath)-1]
			}

			mux.Get(pagePath, func(w http.ResponseWriter, r *http.Request) {
				err := renderPage(w, path)
				if err != nil {
					log.Fatal(err)
				}
			})
		}
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}

	return mux
}

func renderPage(w http.ResponseWriter, pagePath string) error {
	ts, ok := templateCache[pagePath]
	name := filepath.Base(pagePath)

	if !ok {
		page, err := createComponentTemplate(pagePath, name)
		if err != nil {
			return err
		}

		page, err = handleComponentsTemplateChange(page)
		if err != nil {
			return err
		}

		ts, err = template.New(name).Parse("{{template \"base\" .}}\n" + page.Template)
		if err != nil {
			return err
		}

		for _, component := range page.Components {
			ts, err = ts.Parse(component.Template)
			if err != nil {
				return err
			}
		}

		ts, err = ts.Parse(generateBaseTemplate(baseTemplate, page))
	}

	if Cache {
		templateCache[pagePath] = ts
	}

	err := ts.ExecuteTemplate(w, name, nil)
	if err != nil {
		return err
	}

	return nil
}

func createComponentTemplate(path string, name string) (*component, error) {
	var err error

	pg := component{}
	pg.Raw, err = readFile(path)
	if err != nil {
		return nil, err
	}

	pg.Name = name
	pg.Path = path
	pg.CustomClass = "indigoo" + generateCustomString(8)

	pg.Script, err = getJavaScriptCode(*pg.Raw)
	if err != nil {
		return nil, err
	}

	pg.Styles, err = getCSSCode(*pg.Raw)
	if err != nil {
		return nil, err
	}

	pg.Content, err = getHTMLCode(*pg.Raw)
	if err != nil {
		return nil, err
	}

	pg.Components, err = handleComponents(*pg.Script, path)
	if err != nil {
		return nil, err
	}

	for _, component := range pg.Components {
		for _, subComponent := range component.Components {
			pg.Components[subComponent.Path] = subComponent
		}
	}

	pg.Template, err = parseTemplate(pg)
	if err != nil {
		return nil, err
	}

	return &pg, nil
}

func getHTMLCode(tmpl string) (*string, error) {
	result, err := getStringBetween(tmpl, "<component>", "</component>")
	if err != nil {
		return nil, err
	}

	result = strings.TrimSpace(result)

	return &result, nil
}

func handleComponents(tmpl string, path string) (map[string]*component, error) {
	cps := map[string]*component{}

	fileLines := strings.Split(tmpl, "\n")
	var newTemplateLines []string
	for _, line := range fileLines {
		match := componentImportPattern.FindStringSubmatch(line)

		if len(match) > 0 {
			componentNamePattern := regexp.MustCompile(`(\s|{|,)[A-Z]+[^.,\s,}]+`)

			componentMatches := componentNamePattern.FindAllString(line, -1)

			if len(componentMatches) > 0 {
				var componentNames []string
				for _, componentMatch := range componentMatches {
					componentNames = append(componentNames, componentMatch[1:])
				}

				componentPath := filepath.Join(filepath.Dir(path), match[1])

				componentData, err := createComponentTemplate(componentPath, componentNames[0])
				if err != nil {
					return nil, err
				}

				cps[componentPath] = componentData
			}
		} else {
			newTemplateLines = append(newTemplateLines, line)
		}
	}

	return cps, nil
}

func handleComponentsTemplateChange(page *component) (*component, error) {
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

	for _, component := range page.Components {
		componentTagPattern := regexp.MustCompile(`<` + component.Name + `\s*(.*?)<\/` + component.Name + `>|<` + component.Name + `\s*(.*?)\/>`)

		page.Template = handleTemplateReplacement(page, component.CustomClass, componentTagPattern)

		for _, cmptToReplace := range page.Components {
			if cmptToReplace.CustomClass != component.CustomClass {
				cmptToReplace.Template = handleTemplateReplacement(cmptToReplace, component.CustomClass, componentTagPattern)
			}
		}

		page.Template += component.Template
	}

	return page, nil
}

func parseTemplate(cmpt component) (string, error) {
	var tmpl string

	if cmpt.Script != nil {
		tmpl += "{{define \"js-" + cmpt.CustomClass + "\"}}\n<script>\n" + *cmpt.RemoveImports() + "\n</script>\n{{end}}\n"
	}

	if cmpt.Content != nil {
		tmpl += "{{define \"content-" + cmpt.CustomClass + "\"}}\n" + *cmpt.Content + "\n{{end}}\n"
	}

	if cmpt.Styles != nil {
		tmpl += "{{define \"css-" + cmpt.CustomClass + "\"}}\n<style>\n" + *cmpt.Styles + "\n</style>\n{{end}}\n"
	}

	return tmpl, nil
}

func readFile(path string) (*string, error) {
	file, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	tmpl := string(file)

	return &tmpl, nil
}

func generateBaseTemplate(path string, page *component) string {
	file, err := os.ReadFile(path)
	if err != nil {
		log.Fatal(err)
	}

	tmpl := "{{define \"base\"}}\n"
	tmpl += string(file)
	tmpl += "{{end}}\n"

	tmpl = regexp.MustCompile("</head>").ReplaceAllString(tmpl, fmt.Sprintf("{{block \"css-%s\" .}}\n{{end}}\n</head>", page.CustomClass))
	tmpl = regexp.MustCompile("<body>").ReplaceAllString(tmpl, fmt.Sprintf("<body>\n{{block \"content-%s\" .}}\n{{end}}\n", page.CustomClass))
	tmpl = regexp.MustCompile("</body>").ReplaceAllString(tmpl, fmt.Sprintf("{{block \"js-%s\" .}}\n{{end}}\n</body>", page.CustomClass))

	for _, component := range page.Components {
		tmpl = regexp.MustCompile("</head>").ReplaceAllString(tmpl, fmt.Sprintf("{{block \"css-%s\" .}}\n{{end}}\n</head>", component.CustomClass))
		tmpl = regexp.MustCompile("</body>").ReplaceAllString(tmpl, fmt.Sprintf("{{block \"js-%s\" .}}\n{{end}}\n</body>", component.CustomClass))
	}

	return tmpl
}

func validateStructuralFiles() error {
	libRegEx, err := regexp.Compile("(app|index.html|app/page.goo)$")
	if err != nil {
		return err
	}

	err = filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err == nil && libRegEx.MatchString(path) {
			if info.IsDir() && path == "app" {
				appFolder = path
			} else if !info.IsDir() && path == "index.html" {
				baseTemplate = path
			} else if !info.IsDir() && path == "app/page.goo" {
				entryPage = path
			}
		}

		return nil
	})
	if err != nil {
		return err
	}

	if appFolder == "" || baseTemplate == "" || entryPage == "" {
		return errors.New("no app folder, entry page.goo and/or index.html file found")
	}

	return nil
}
