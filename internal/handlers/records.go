package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"education/internal/database"
	"education/internal/repository"
)

type RecordHandler struct {
	stores []*database.Store
}

func NewRecordHandler(stores []*database.Store) *RecordHandler {
	return &RecordHandler{stores: stores}
}

func (h *RecordHandler) ListDatabases(w http.ResponseWriter, r *http.Request) {
	respond(w, http.StatusOK, h.stores)
}

func (h *RecordHandler) resolve(w http.ResponseWriter, r *http.Request) (*repository.Records, *database.Table, bool) {
	var store *database.Store
	var table *database.Table
	for _, s := range h.stores {
		if s.Name == r.PathValue("database") {
			store = s
			break
		}
	}
	if store != nil {
		for i := range store.Tables {
			if store.Tables[i].Name == r.PathValue("table") {
				table = &store.Tables[i]
				break
			}
		}
	}
	if table == nil {
		fail(w, 404, "Unknown database or table")
		return nil, nil, false
	}
	records, err := repository.NewRecords(store.DB, table.Name)
	if err != nil {
		dbError(w, err)
		return nil, nil, false
	}
	return records, table, true
}

func recordID(w http.ResponseWriter, r *http.Request) (int64, bool) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id < 1 {
		fail(w, http.StatusBadRequest, "Invalid record ID")
		return 0, false
	}
	return id, true
}

func readValues(w http.ResponseWriter, r *http.Request, table *database.Table) ([]any, bool) {
	var input map[string]any
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 65536))
	decoder.UseNumber()
	if err := decoder.Decode(&input); err != nil {
		fail(w, 400, "Expected a JSON object")
		return nil, false
	}
	if len(input) != len(table.Columns) {
		fail(w, 400, "Supply exactly the table's editable fields")
		return nil, false
	}
	values := []any{}
	for _, c := range table.Columns {
		v, ok := input[c]
		if !ok {
			fail(w, 400, "Missing field: "+c)
			return nil, false
		}
		if v != nil {
			switch val := v.(type) {
			case string:
			case json.Number:
				v = string(val)
			default:
				fail(w, 400, "Fields must be text, numbers, or null")
				return nil, false
			}
		}
		values = append(values, v)
	}
	return values, true
}

func (h *RecordHandler) List(w http.ResponseWriter, r *http.Request) {
	store, _, ok := h.resolve(w, r)
	if !ok {
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	result, err := store.List(ctx)
	if err != nil {
		dbError(w, err)
		return
	}
	respond(w, http.StatusOK, result)
}

func (h *RecordHandler) Create(w http.ResponseWriter, r *http.Request) {
	store, table, ok := h.resolve(w, r)
	if !ok {
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	values, ok := readValues(w, r, table)
	if !ok {
		return
	}

	id, err := store.Create(ctx, values)
	if err != nil {
		dbError(w, err)
		return
	}
	respond(w, http.StatusCreated, map[string]int64{"id": id})
}

func (h *RecordHandler) Update(w http.ResponseWriter, r *http.Request) {
	store, table, ok := h.resolve(w, r)
	if !ok {
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	id, ok := recordID(w, r)
	if !ok {
		return
	}

	values, ok := readValues(w, r, table)
	if !ok {
		return
	}

	result, err := store.Update(ctx, id, values)
	finish(w, result, err)
}

func (h *RecordHandler) Delete(w http.ResponseWriter, r *http.Request) {
	store, _, ok := h.resolve(w, r)
	if !ok {
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	id, ok := recordID(w, r)
	if !ok {
		return
	}

	result, err := store.Delete(ctx, id)
	finish(w, result, err)
}
