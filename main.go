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
		page, err := new(component).New(pagePath, name)
		if err != nil {
			return err
		}

		page.handleComponentsTemplateChange()

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
