package handlers

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"

	"education/internal/service"
	"github.com/jackc/pgx/v5/pgconn"
)

func respond(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func fail(w http.ResponseWriter, status int, message string) {
	respond(w, status, map[string]string{"error": message})
}

func dbError(w http.ResponseWriter, err error) {
	var conflict *service.ConflictError
	if errors.As(err, &conflict) {
		fail(w, http.StatusConflict, conflict.Message)
		return
	}

	var pe *pgconn.PgError
	if errors.As(err, &pe) && (strings.HasPrefix(pe.Code, "23") || strings.HasPrefix(pe.Code, "22")) {
		fail(w, 400, "Invalid data: "+pe.Message)
		return
	}
	
	log.Printf("database operation: %v", err)
	fail(w, 503, "Database unavailable; please try again.")
}

func finish(w http.ResponseWriter, found bool, err error) {
	if err != nil {
		dbError(w, err)
		return
	}

	if !found {
		fail(w, 404, "Record not found")
		return
	}

	respond(w, 200, map[string]bool{"ok": true})
}
