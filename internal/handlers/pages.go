package handlers

import (
	"html/template"
	"log"
	"net/http"
)

type PageHandler struct {
	index *template.Template
}

func NewPageHandler() (*PageHandler, error) {
	index, err := template.ParseFiles("templates/index.html")
	if err != nil {
		return nil, err
	}

	return &PageHandler{index: index}, nil
}

func (h *PageHandler) Home(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := h.index.Execute(w, nil); err != nil {
		log.Printf("render home page: %v", err)
	}
}
