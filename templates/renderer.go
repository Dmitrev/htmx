package templates

import (
	"bytes"
	"fmt"
	"html/template"
	"net/http"
	"path/filepath"
)

type Renderer struct {
    templates map[string]*template.Template
    devmode bool
}

func MakeRenderer(devmode bool) *Renderer {

    r := &Renderer{
	devmode: devmode,
    }

    r.loadTemplates()

    return r
}

func (r *Renderer) loadTemplates() {
    // var paths []string
    r.templates = make(map[string]*template.Template)
    pageFiles, _ := filepath.Glob("views/pages/*.gohtml")
    layoutFiles, _ := filepath.Glob("views/layouts/*.gohtml")

    for _, page := range pageFiles {
	paths := append(layoutFiles, page)
	r.templates[filepath.Base(page)] = template.Must(template.ParseFiles(paths...))
    }

    fmt.Printf("%v\n", r.templates)
}

func (r *Renderer) Render(w http.ResponseWriter, name string, data interface{}) {

    if r.devmode {
	r.loadTemplates()
    }

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

