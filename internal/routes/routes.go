package routes

import (
	"net/http"

	"education/internal/handlers"
)

func RegisterRoutes(mux *http.ServeMux, records *handlers.RecordHandler, pages *handlers.PageHandler) {
	mux.HandleFunc("/api/databases", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			records.ListDatabases(w, r)
		default:
			methodNotAllowed(w, "GET")
		}
	})

	mux.HandleFunc("/api/{database}/{table}", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			records.List(w, r)
		case http.MethodPost:
			records.Create(w, r)
		default:
			methodNotAllowed(w, "GET, POST")
		}
	})

	mux.HandleFunc("/api/{database}/{table}/{id}", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPut:
			records.Update(w, r)
		case http.MethodDelete:
			records.Delete(w, r)
		default:
			methodNotAllowed(w, "PUT, DELETE")
		}
	})

	files := http.FileServer(http.Dir("./static"))
	mux.Handle("/static/", http.StripPrefix("/static/", files))

	mux.HandleFunc("/{$}", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			pages.Home(w, r)
		default:
			methodNotAllowed(w, "GET")
		}
	})
}

func methodNotAllowed(w http.ResponseWriter, allowed string) {
	w.Header().Set("Allow", allowed)
	http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
}
