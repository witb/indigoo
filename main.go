package indigoo

import (
	"errors"
	"github.com/go-chi/chi/v5"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
)

var appFolder string
var baseTemplate string
var entryPage string

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
	page, components, err := createPageTemplate(pagePath)
	if err != nil {
		return err
	}

	log.Println(components)

	name := filepath.Base(pagePath)
	ts, err := template.New(name).Parse(*page)
	if err != nil {
		return err
	}

	ts, err = ts.Parse(baseTemplate)

	err = ts.ExecuteTemplate(w, name, nil)
	if err != nil {
		return err
	}

	return nil
}

func createPageTemplate(pagePath string) (*string, *[]string, error) {
	file, err := os.ReadFile(pagePath)
	if err != nil {
		return nil, nil, err
	}

	componentPaths := []string{}
	fileString := string(file)
	fileString = regexp.MustCompile("import\\s+\\(([^)]+)\\)").ReplaceAllStringFunc(fileString, func(s string) string {
		componentPaths = append(componentPaths, regexp.MustCompile("\"(.+)\"").FindStringSubmatch(s)[1])
		return ""
	})

	tmpl := "{{template \"base\" .}}\n"
	tmpl += fileString

	if regexp.MustCompile("<script>").MatchString(tmpl) {
		tmpl = regexp.MustCompile("<script>").ReplaceAllString(tmpl, "{{define \"js\"}}\n<script>")
		tmpl = regexp.MustCompile("</script>").ReplaceAllString(tmpl, "</script>\n{{define \"content\"}}\n")
		tmpl = regexp.MustCompile("</script>").ReplaceAllString(tmpl, "</script>\n{{end}}\n")
	} else {
		tmpl = regexp.MustCompile("{{template \"base\" .}}").ReplaceAllString(tmpl, "{{template \"base\" .}}\n{{define \"content\"}}\n")
	}

	if regexp.MustCompile("<style>").MatchString(tmpl) {
		tmpl = regexp.MustCompile("<style>").ReplaceAllString(tmpl, "{{end}}\n<style>")
		tmpl = regexp.MustCompile("<style>").ReplaceAllString(tmpl, "{{define \"css\"}}\n<style>")
		tmpl = regexp.MustCompile("</style>").ReplaceAllString(tmpl, "</style>\n{{end}}\n")
	} else {
		tmpl += "{{end}}"
	}

	return &tmpl, &componentPaths, nil
}

func generateBaseTemplate(path string) string {
	file, err := os.ReadFile(path)
	if err != nil {
		log.Fatal(err)
	}

	tmpl := "{{define \"base\"}}\n"
	tmpl += string(file)
	tmpl += "{{end}}\n"

	tmpl = regexp.MustCompile("</head>").ReplaceAllString(tmpl, "{{block \"css\" .}}\n{{end}}\n</head>")
	tmpl = regexp.MustCompile("<body>").ReplaceAllString(tmpl, "<body>\n{{block \"content\" .}}\n{{end}}\n")
	tmpl = regexp.MustCompile("</body>").ReplaceAllString(tmpl, "{{block \"js\" .}}\n{{end}}\n</html>")

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
				baseTemplate = generateBaseTemplate(path)
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
