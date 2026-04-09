package tarotweb

import (
	"embed"
	"encoding/json"
	"html/template"
	"io/fs"
	"net/http"
)

//go:embed templates/*.html static/*
var assets embed.FS

var staticFiles fs.FS

func init() {
	staticFiles, _ = fs.Sub(assets, "static")
}

func parseTemplates() (*template.Template, error) {
	return template.New("").Funcs(template.FuncMap{
		"join": func(items []string, sep string) string {
			out := ""
			for i, item := range items {
				if i > 0 {
					out += sep
				}
				out += item
			}
			return out
		},
	}).ParseFS(assets, "templates/*.html")
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_ = json.NewEncoder(w).Encode(v)
}
