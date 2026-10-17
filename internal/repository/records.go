package repository

import (
	"context"
	"database/sql"
	"fmt"
)

type queries struct {
	list, create, update, delete string
}

var tableQueries = map[string]queries{
	"students": {
		list:   `SELECT id, username, full_name, grade_level, region FROM students ORDER BY id LIMIT 500`,
		create: `INSERT INTO students (username, full_name, grade_level, region) VALUES ($1, $2, $3, $4) RETURNING id`,
		update: `UPDATE students SET username = $1, full_name = $2, grade_level = $3, region = $4 WHERE id = $5`,
		delete: `DELETE FROM students WHERE id = $1`,
	},
	"teachers": {
		list:   `SELECT id, username, full_name, subject FROM teachers ORDER BY id LIMIT 500`,
		create: `INSERT INTO teachers (username, full_name, subject) VALUES ($1, $2, $3) RETURNING id`,
		update: `UPDATE teachers SET username = $1, full_name = $2, subject = $3 WHERE id = $4`,
		delete: `DELETE FROM teachers WHERE id = $1`,
	},
	"courses": {
		list:   `SELECT id, title, grade_level, teacher_id FROM courses ORDER BY id LIMIT 500`,
		create: `INSERT INTO courses (title, grade_level, teacher_id) VALUES ($1, $2, $3) RETURNING id`,
		update: `UPDATE courses SET title = $1, grade_level = $2, teacher_id = $3 WHERE id = $4`,
		delete: `DELETE FROM courses WHERE id = $1`,
	},
	"videos": {
		list:   `SELECT id, name, address, course_id FROM videos ORDER BY id LIMIT 500`,
		create: `INSERT INTO videos (name, address, course_id) VALUES ($1, $2, $3) RETURNING id`,
		update: `UPDATE videos SET name = $1, address = $2, course_id = $3 WHERE id = $4`,
		delete: `DELETE FROM videos WHERE id = $1`,
	},
	"grades": {
		list:   `SELECT id, student_id, course_id, teacher_id, grade FROM grades ORDER BY id LIMIT 500`,
		create: `INSERT INTO grades (student_id, course_id, teacher_id, grade) VALUES ($1, $2, $3, $4) RETURNING id`,
		update: `UPDATE grades SET student_id = $1, course_id = $2, teacher_id = $3, grade = $4 WHERE id = $5`,
		delete: `DELETE FROM grades WHERE id = $1`,
	},
}

type Records struct {
	table   string
	db      *sql.DB
	queries queries
}

func NewRecords(db *sql.DB, table string) (*Records, error) {
	q, ok := tableQueries[table]
	if !ok {
		return nil, fmt.Errorf("unsupported table: %s", table)
	}
	
	return &Records{db: db, queries: q, table: table}, nil
}

func (r *Records) List(ctx context.Context) ([]map[string]any, error) {
	rows, err := r.db.QueryContext(ctx, r.queries.list)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	result := []map[string]any{}
	for rows.Next() {
		values := make([]any, len(columns))
		pointers := make([]any, len(columns))

		for i := range values {
			pointers[i] = &values[i]
		}

		if err := rows.Scan(pointers...); err != nil {
			return nil, err
		}

		item := map[string]any{}
		for i, column := range columns {
			if value, ok := values[i].([]byte); ok {
				item[column] = string(value)
			} else {
				item[column] = values[i]
			}
		}

		result = append(result, item)
	}
	return result, rows.Err()
}

func (r *Records) Create(ctx context.Context, values []any) (int64, error) {
	var id int64
	err := r.db.QueryRowContext(ctx, r.queries.create, values...).Scan(&id)
	return id, err
}

func (r *Records) Update(ctx context.Context, id int64, values []any) (bool, error) {
	args := append(append([]any{}, values...), id)
	result, err := r.db.ExecContext(ctx, r.queries.update, args...)
	return changed(result, err)
}

func (r *Records) Delete(ctx context.Context, id int64) (bool, error) {
	result, err := r.db.ExecContext(ctx, r.queries.delete, id)
	return changed(result, err)
}

func changed(result sql.Result, err error) (bool, error) {
	if err != nil {
		return false, err
	}

	count, err := result.RowsAffected()
	
	return count > 0, err
}
