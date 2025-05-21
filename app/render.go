package app

import (
	"fmt"
	"html/template"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin/render"
)

// renderer holds your pre-parsed templates, keyed by name.
type renderer struct {
	templates map[string]*template.Template
}

// NewRenderer creates an HTMLRender by parsing every `.html` in folder.
func NewRenderer(folder string) (render.HTMLRender, error) {
	r := &renderer{templates: make(map[string]*template.Template)}

	err := filepath.WalkDir(folder, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		key := filepath.ToSlash(path)[len(folder)+1:]
		tmpl, err := template.ParseFiles(path)
		if err != nil {
			return err
		}
		r.templates[key] = tmpl
		return nil
	})
	if err != nil {
		return nil, err
	}
	return r, nil
}

// Instance satisfies gin.HTMLRender.  Gin will call this for each request.
func (r *renderer) Instance(name string, data any) render.Render {
	fmt.Printf("%+v\n", r.templates)
	tmpl, ok := r.templates[name]
	if !ok {
		// panic or fallback
		panic("template not found: " + name)
	}
	return &htmlRender{tmpl: tmpl, data: data}
}

type htmlRender struct {
	tmpl *template.Template
	data any
}

// WriteContentType is required by gin.Render
func (h *htmlRender) WriteContentType(w http.ResponseWriter) {
	header := w.Header()
	if val := header["Content-Type"]; len(val) == 0 {
		header.Set("Content-Type", "text/html; charset=utf-8")
	}
}

// Render executes the template.
func (h *htmlRender) Render(w http.ResponseWriter) error {
	return h.tmpl.Execute(w, h.data)
}
