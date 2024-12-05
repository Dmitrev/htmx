package templates

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"
	// "io/fs"
	"net/http"
	// "os"
	"path/filepath"
	// "strings"
)


type Renderer struct {
    templates map[string]*template.Template
}

func MakeRenderer(resources embed.FS) *Renderer {


    templates := make(map[string]*template.Template)
    // var paths []string
    pageFiles, _ := filepath.Glob("views/pages/*.gohtml")
    layoutFiles, _ := filepath.Glob("views/layouts/*.gohtml")

    // tmpl := template.Must(template.New("pages/index.gohtml").Parse("This is the index page"))

    for _, page := range pageFiles {
	paths := append(layoutFiles, page)
	templates[filepath.Base(page)] = template.Must(template.ParseFiles(paths...))
    }

	fmt.Printf("%v\n", templates)

    // tmpl := template.Must(template.ParseFS(resources, paths...))

    return &Renderer{templates}
}

func (r *Renderer) Render(w http.ResponseWriter, name string, data interface{}) {
    var buffer bytes.Buffer

    err := r.templates[name].ExecuteTemplate(&buffer, name, data)

    if err != nil {
	err = fmt.Errorf("error executing template: %w", err)
	http.Error(w, err.Error(), http.StatusInternalServerError)
	return
    }

    w.Header().Set("Content-Type", "text/html; charset=UTF-8")
    buffer.WriteTo(w)
}

