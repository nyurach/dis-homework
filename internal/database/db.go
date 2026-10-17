package database

import (
	"context"
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/stdlib"
)

//go:embed sql/*.sql
var scripts embed.FS

type Table struct {
	Name    string   `json:"name"`
	Columns []string `json:"columns"`
}
type Store struct {
	Name   string  `json:"name"`
	Title  string  `json:"title"`
	Tables []Table `json:"tables"`
	DB     *sql.DB `json:"-"`
}

func Catalog() []*Store {
	return []*Store{
		{
			Name:  "students",
			Title: "Students",
			Tables: []Table{
				{Name: "students", Columns: []string{"username", "full_name", "grade_level", "region"}},
			},
		},
		{
			Name:  "teachers",
			Title: "Teachers",
			Tables: []Table{
				{Name: "teachers", Columns: []string{"username", "full_name", "subject"}},
			},
		},
		{
			Name:  "courses",
			Title: "Courses",
			Tables: []Table{
				{Name: "courses", Columns: []string{"title", "grade_level", "teacher_id"}},
			},
		},
		{
			Name:  "videos",
			Title: "Videos & content",
			Tables: []Table{
				{Name: "videos", Columns: []string{"name", "address", "course_id"}},
			},
		},
		{
			Name:  "grades",
			Title: "Course grades",
			Tables: []Table{
				{Name: "grades", Columns: []string{"student_id", "course_id", "teacher_id", "grade"}},
			},
		},
	}
}

func Open(s *Store) error {
	config, err := localConfig(s.Name)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	db := stdlib.OpenDB(*config)
	db.SetMaxOpenConns(5)
	err = db.PingContext(ctx)

	var postgresError *pgconn.PgError
	if errors.As(err, &postgresError) && postgresError.Code == "3D000" {
		if createErr := createDatabase(ctx, config); createErr != nil {
			db.Close()
			return fmt.Errorf("create %s: %w", config.Database, createErr)
		}
		err = db.PingContext(ctx)
	}
	if err != nil {
		db.Close()
		return fmt.Errorf("connect to %s at %s:%d: %w", s.Name, config.Host, config.Port, err)
	}

	script, err := scripts.ReadFile("sql/" + s.Name + ".sql")
	if err != nil {
		db.Close()
		return err
	}

	if err = initialize(ctx, db, string(script)); err != nil {
		db.Close()
		return fmt.Errorf("initialize %s: %w", s.Name, err)
	}
	s.DB = db
	return nil
}

// initialize commits schema and sample rows together for one database.
func initialize(ctx context.Context, db *sql.DB, script string) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, script); err != nil {
		return fmt.Errorf("execute initialization: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit initialization: %w", err)
	}
	return nil
}

func createDatabase(ctx context.Context, config *pgx.ConnConfig) error {
	adminConfig := config.Copy()
	adminConfig.Database = "postgres"
	admin := stdlib.OpenDB(*adminConfig)
	defer admin.Close()

	// PostgreSQL requires CREATE DATABASE to run outside a transaction.
	name := pgx.Identifier{config.Database}.Sanitize()
	_, err := admin.ExecContext(ctx, "CREATE DATABASE "+name)

	var postgresError *pgconn.PgError
	if errors.As(err, &postgresError) && postgresError.Code == "42P04" {
		return nil
	}
	return err
}
