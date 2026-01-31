package templates

import (
	"bytes"
	"fmt"
	"html/template"
	"net/http"
	"path/filepath"
	"sync"
)

type Renderer struct {
    mu sync.RWMutex
    templates map[string]*template.Template
    devmode bool
}

func MakeRenderer(devmode bool) *Renderer {
    r := &Renderer{
	templates: map[string]*template.Template{},
	devmode: devmode,
    }

    r.loadTemplates()

    return r
}

func (r *Renderer) loadTemplates() {
    r.mu.Lock()
    defer r.mu.Unlock()
    layoutFiles, err := filepath.Glob("views/layouts/*.gohtml")

    if err != nil {
	panic(err)
    }

    pageFiles, err := filepath.Glob("views/pages/*.gohtml")

    if err != nil {
	panic(err)
    }

    componentFiles, err := filepath.Glob("views/components/*.gohtml")

    if err != nil {
	panic(err)
    }

    for _, page := range pageFiles {
	files := append(layoutFiles, page)
	files = append(files, componentFiles...)
	tmpl := template.Must(template.New("").ParseFiles(files...))
	r.templates[page] = tmpl
    }

    for _, component := range componentFiles {
	tmpl := template.Must(template.New("").ParseFiles(component))
	r.templates[component] = tmpl
    }


    // var paths []string
    // componentFiles, _ := filepath.Glob("views/components/*.gohtml")
    // paths := append(pageFiles, layoutFiles...)
    // paths = append(paths, componentFiles...)
    // r.template = template.Must(template.ParseFiles(paths...))
}

func (r *Renderer) Render(w http.ResponseWriter, name string, data any) {

    if r.devmode {
	r.loadTemplates()
    }

    var buffer bytes.Buffer
    
    template, ok := r.templates[name]

    if !ok {
	err := fmt.Errorf("Could not find template: %s", name)
	http.Error(w, err.Error(), http.StatusInternalServerError)
	return
    }

    err := template.ExecuteTemplate(&buffer, "base.gohtml", data)

    if err != nil {
	err = fmt.Errorf("error executing template: %w", err)
	http.Error(w, err.Error(), http.StatusInternalServerError)
	return
    }

    w.Header().Set("Content-Type", "text/html; charset=UTF-8")
    buffer.WriteTo(w)
}

func (r *Renderer) RenderComponent(w http.ResponseWriter, name string, data any) {

    if r.devmode {
	r.loadTemplates()
    }

    var buffer bytes.Buffer
    
    template, ok := r.templates[name]

    if !ok {
	err := fmt.Errorf("Could not find component: %s", name)
	http.Error(w, err.Error(), http.StatusInternalServerError)
	return
    }

    err := template.ExecuteTemplate(&buffer, name, data)

    if err != nil {
	err = fmt.Errorf("error executing component: %w", err)
	http.Error(w, err.Error(), http.StatusInternalServerError)
	return
    }

    w.Header().Set("Content-Type", "text/html; charset=UTF-8")
    buffer.WriteTo(w)
}

