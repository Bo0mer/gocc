package main

import (
	"context"
	"fmt"
	"strconv"

	"gorm.io/gorm"
)

type DBStudent struct {
	ID   int    `gorm:"column:id"`
	Name string `gorm:"column:name"`
	Age  int    `gorm:"column:age"`
}

func (DBStudent) TableName() string {
	return "students"
}

func (s DBStudent) toModelStudent() Student {
	return Student{
		ID:   fmt.Sprint(s.ID),
		Name: s.Name,
		Age:  s.Age,
	}
}

func fromModelStudent(s Student) (DBStudent, error) {
	id, err := strconv.ParseInt(s.ID, 10, 64)
	if err != nil {
		return DBStudent{}, err
	}

	return DBStudent{
		ID:   int(id),
		Name: s.Name,
		Age:  s.Age,
	}, nil
}

type GORMStudentDatabase struct {
	db *gorm.DB
}

func (db *GORMStudentDatabase) CreateStudent(ctx context.Context, s Student) (Student, error) {
	err := db.db.Create(&s).Error
	if err != nil {
		return Student{}, nil
	}

	return s, nil
}

func (db *GORMStudentDatabase) DeleteStudent(ctx context.Context, id string) (Student, error) {
	panic("todo")
}

func (db *GORMStudentDatabase) ListStudents(ctx context.Context) ([]Student, error) {
	panic("todo")
}

func (db *GORMStudentDatabase) Student(ctx context.Context, id string) (Student, error) {
	panic("todo")
}
