package main

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

type SQLStudentDatabase struct {
	db *sql.DB
}

func NewSQLStudentDatabase(db *sql.DB) *SQLStudentDatabase {
	return &SQLStudentDatabase{
		db: db,
	}
}

var _ Database = (*SQLStudentDatabase)(nil)

func (db *SQLStudentDatabase) CreateStudent(ctx context.Context, s Student) (Student, error) {
	result, err := db.db.ExecContext(
		ctx,
		"INSERT INTO students (name, age) VALUES (?, ?)",
		s.Name,
		s.Age,
	)
	if err != nil {
		return Student{}, fmt.Errorf("error during insert: %w", err)
	}

	lastID, err := result.LastInsertId()
	if err != nil {
		return Student{}, fmt.Errorf("error getting last insert id: %w", err)
	}

	return Student{
		ID:   fmt.Sprint(lastID),
		Name: s.Name,
		Age:  s.Age,
	}, nil
}

func (db *SQLStudentDatabase) DeleteStudent(ctx context.Context, id string) (Student, error) {
	tx, err := db.db.BeginTx(ctx, nil)
	if err != nil {
		return Student{}, fmt.Errorf("error during begin tx: %w", err)
	}
	defer tx.Rollback()

	row := tx.QueryRowContext(ctx, "SELECT * FROM students WHERE id = ?", id)
	var toDelete Student
	err = row.Scan(&toDelete.ID, &toDelete.Name, &toDelete.Age)
	if err != nil {
		return Student{}, fmt.Errorf("could not scan student to delete: %w", err)
	}

	_, err = tx.ExecContext(
		ctx,
		"DELETE FROM students WHERE id = ?",
		id,
	)
	if err != nil {
		return Student{}, fmt.Errorf("error during delete: %w", err)
	}

	err = tx.Commit()
	if err != nil {
		return Student{}, fmt.Errorf("error during commit: %w", err)
	}

	return toDelete, nil
}

func (db *SQLStudentDatabase) ListStudents(ctx context.Context) ([]Student, error) {
	time.Sleep(2 * time.Second)
	rows, err := db.db.QueryContext(ctx, "SELECT * FROM students")
	if err != nil {
		return nil, fmt.Errorf("error during select: %w", err)
	}
	defer rows.Close()

	var students []Student
	for rows.Next() {
		var student Student
		err = rows.Scan(&student.ID, &student.Name, &student.Age)
		if err != nil {
			return nil, fmt.Errorf("error during scan: %w", err)
		}

		students = append(students, student)
	}

	return students, nil
}

func (db *SQLStudentDatabase) Student(ctx context.Context, id string) (Student, error) {
	row := db.db.QueryRowContext(ctx, "SELECT * FROM students WHERE id = ?", id)
	var student Student
	err := row.Scan(&student.ID, &student.Name, &student.Age)
	if err != nil {
		return Student{}, fmt.Errorf("could not scan student: %w", err)
	}

	return student, nil
}
