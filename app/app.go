package app

import (
	"context"
	"database/sql"
	"html/template"
	"net/http"
)

type Context struct {
    context.Context
    ResponseWriter http.ResponseWriter
    Request *http.Request
    Templates *template.Template
    Repo *Repository
}

type Repository struct {
    DB *sql.DB
}

func NewContext(w http.ResponseWriter, r *http.Request, t *template.Template, repo *Repository) *Context {
    return &Context {
	Context: r.Context(),
	ResponseWriter: w,
	Request: r,
	Templates: t,
	Repo: repo,
    }
}
