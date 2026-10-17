package repository

import (
	"context"
	"fmt"
)

var referenceQueries = map[string]map[string]string{
	"courses": {"teacher_id": `SELECT count(*) FROM courses WHERE teacher_id = $1`},
	"videos":  {"course_id": `SELECT count(*) FROM videos WHERE course_id = $1`},
	"grades": {
		"student_id": `SELECT count(*) FROM grades WHERE student_id = $1`,
		"teacher_id": `SELECT count(*) FROM grades WHERE teacher_id = $1`,
		"course_id":  `SELECT count(*) FROM grades WHERE course_id = $1`,
	},
}

var existsQueries = map[string]string{
	"students": `SELECT EXISTS (SELECT 1 FROM students WHERE id = $1)`,
	"teachers": `SELECT EXISTS (SELECT 1 FROM teachers WHERE id = $1)`,
	"courses":  `SELECT EXISTS (SELECT 1 FROM courses WHERE id = $1)`,
	"videos":   `SELECT EXISTS (SELECT 1 FROM videos WHERE id = $1)`,
	"grades":   `SELECT EXISTS (SELECT 1 FROM grades WHERE id = $1)`,
}

func (r *Records) Exists(ctx context.Context, id int64) (bool, error) {
	var exists bool
	err := r.db.QueryRowContext(ctx, existsQueries[r.table], id).Scan(&exists)
	return exists, err
}

func (r *Records) CountReferences(ctx context.Context, column string, id int64) (int64, error) {
	query, ok := referenceQueries[r.table][column]
	if !ok {
		return 0, fmt.Errorf("unsupported reference: %s.%s", r.table, column)
	}
	var count int64
	err := r.db.QueryRowContext(ctx, query, id).Scan(&count)
	return count, err
}
