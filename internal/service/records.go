// Package service coordinates record operations across the separate databases.
package service

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"sync"

	"education/internal/database"
	"education/internal/repository"
)

type recordStore interface {
	Create(context.Context, []any) (int64, error)
	Update(context.Context, int64, []any) (bool, error)
	Delete(context.Context, int64) (bool, error)
	Exists(context.Context, int64) (bool, error)
	CountReferences(context.Context, string, int64) (int64, error)
}

type ConflictError struct{ Message string }

func (e *ConflictError) Error() string { return e.Message }

type relation struct{ child, column, parent string }

var relations = []relation{
	{"courses", "teacher_id", "teachers"},
	{"videos", "course_id", "courses"},
	{"grades", "student_id", "students"},
	{"grades", "teacher_id", "teachers"},
	{"grades", "course_id", "courses"},
}

// The lock prevents app requests from inserting a reference between a delete's
// dependency check and the deletion. Direct database edits bypass this lock.
type Records struct {
	mu      sync.Mutex
	stores  map[string]recordStore
	columns map[string][]string
}

func NewRecords(stores []*database.Store) *Records {
	s := &Records{stores: map[string]recordStore{}, columns: map[string][]string{}}
	for _, store := range stores {
		for _, table := range store.Tables {
			records, err := repository.NewRecords(store.DB, table.Name)
			if err == nil {
				s.stores[table.Name] = records
				s.columns[table.Name] = table.Columns
			}
		}
	}
	return s
}

func (s *Records) store(table string) (recordStore, error) {
	store, ok := s.stores[table]
	if !ok {
		return nil, fmt.Errorf("unavailable table: %s", table)
	}
	return store, nil
}

func (s *Records) validate(ctx context.Context, table string, values []any) error {
	if len(values) != len(s.columns[table]) {
		return fmt.Errorf("invalid field count for %s", table)
	}
	for _, rel := range relations {
		if rel.child != table {
			continue
		}
		for i, column := range s.columns[table] {
			if column != rel.column {
				continue
			}
			id, err := strconv.ParseInt(fmt.Sprint(values[i]), 10, 64)
			if err != nil || id < 1 {
				return &ConflictError{Message: "Select a valid record for " + strings.ReplaceAll(column, "_", " ") + "."}
			}
			parent, err := s.store(rel.parent)
			if err != nil {
				return err
			}
			exists, err := parent.Exists(ctx, id)
			if err != nil {
				return err
			}
			if !exists {
				return &ConflictError{Message: fmt.Sprintf("Cannot save: %s record #%d no longer exists. Select another record.", rel.parent, id)}
			}
		}
	}
	return nil
}

func (s *Records) Create(ctx context.Context, table string, values []any) (int64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	store, err := s.store(table)
	if err != nil {
		return 0, err
	}
	if err := s.validate(ctx, table, values); err != nil {
		return 0, err
	}
	return store.Create(ctx, values)
}

func (s *Records) Update(ctx context.Context, table string, id int64, values []any) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	store, err := s.store(table)
	if err != nil {
		return false, err
	}
	exists, err := store.Exists(ctx, id)
	if err != nil || !exists {
		return false, err
	}
	if err := s.validate(ctx, table, values); err != nil {
		return false, err
	}
	return store.Update(ctx, id, values)
}

func (s *Records) Delete(ctx context.Context, table string, id int64) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	store, err := s.store(table)
	if err != nil {
		return false, err
	}
	exists, err := store.Exists(ctx, id)
	if err != nil || !exists {
		return false, err
	}
	var related []string
	for _, rel := range relations {
		if rel.parent != table {
			continue
		}
		child, err := s.store(rel.child)
		if err != nil {
			return false, err
		}
		count, err := child.CountReferences(ctx, rel.column, id)
		if err != nil {
			return false, err
		}
		if count > 0 {
			related = append(related, fmt.Sprintf("%s (%d)", rel.child, count))
		}
	}
	if len(related) > 0 {
		return false, &ConflictError{Message: fmt.Sprintf("Cannot delete record #%d: related records exist in %s. Reassign or delete those records first.", id, strings.Join(related, ", "))}
	}
	return store.Delete(ctx, id)
}
