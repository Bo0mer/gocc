package main

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"sync"
)

type Student struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Age  int    `json:"age"`
}

type Database interface {
	ListStudents(ctx context.Context) ([]Student, error)
	Student(ctx context.Context, id string) (Student, error)
	// CreateStudent creates the specified student and returns it with ID
	// provided.
	CreateStudent(ctx context.Context, s Student) (Student, error)
	DeleteStudent(ctx context.Context, id string) (Student, error)
}

type InMemoryStudentDatabase struct {
	mu        sync.RWMutex
	students  map[string]Student
	idCounter int
}

func NewInMemoryStudentDatabase() *InMemoryStudentDatabase {
	return &InMemoryStudentDatabase{
		students: make(map[string]Student),
	}
}

func (db *InMemoryStudentDatabase) CreateStudent(s Student) (Student, error) {
	db.mu.Lock()
	defer db.mu.Unlock()

	db.idCounter++

	s.ID = fmt.Sprintf("%d", db.idCounter)
	db.students[s.ID] = s

	return s, nil
}

func (db *InMemoryStudentDatabase) DeleteStudent(id string) (Student, error) {
	db.mu.Lock()
	defer db.mu.Unlock()

	student, ok := db.students[id]
	if !ok {
		return Student{}, fmt.Errorf("unknown student with id: %s", id)
	}
	delete(db.students, id)

	return student, nil
}

func (db *InMemoryStudentDatabase) ListStudents() ([]Student, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	students := make([]Student, 0, len(db.students))
	for _, student := range db.students {
		students = append(students, student)
	}

	return students, nil
}

func (db *InMemoryStudentDatabase) Student(id string) (Student, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	student, ok := db.students[id]
	if !ok {
		return Student{}, fmt.Errorf("unknown student with id: %s", id)
	}

	return student, nil
}

type StudentServer struct {
	db Database
}

func NewStudentServer(db Database) *StudentServer {
	return &StudentServer{
		db: db,
	}
}

var _ http.Handler = (*StudentServer)(nil)

func (s *StudentServer) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	/*
		GET /api/v1/students
		GET /api/v1/students/<id>
		POST /api/v1/students/<id>
		DELETE /api/v1/students/<id>
	*/
	switch {
	case req.Method == "GET" && req.URL.Path == "/api/v1/students":
		s.listStudents(w, req)
	case req.Method == "GET" && strings.HasPrefix(req.URL.Path, "/api/v1/students/"):
		s.getStudent(w, req)
	case req.Method == "POST" && req.URL.Path == "/api/v1/students":
		s.createStudent(w, req)
	case req.Method == "DELETE" && strings.HasPrefix(req.URL.Path, "/api/v1/students/"):
		s.deleteStudent(w, req)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		fmt.Fprintf(w, "unknown method or path\n")
	}
}

func (s *StudentServer) listStudents(w http.ResponseWriter, req *http.Request) {
	students, err := s.db.ListStudents(req.Context())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "error listing students")
		log.Printf("error listing students: %v", err)
		return
	}

	err = json.NewEncoder(w).Encode(students)
	log.Printf("error writing students: %v", err)
}

func (s *StudentServer) getStudent(w http.ResponseWriter, req *http.Request) {
	// GET /api/v1/students/<id>
	segments := strings.Split(req.URL.Path, "/")
	id := segments[len(segments)-1]
	student, err := s.db.Student(req.Context(), id)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "error listing students")
		log.Printf("error listing student: %v", err)
		return
	}

	err = json.NewEncoder(w).Encode(student)
	log.Printf("error writing student: %v", err)
}

func (s *StudentServer) createStudent(w http.ResponseWriter, req *http.Request) {
	var student Student

	body, err := io.ReadAll(req.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "cannot read request")
		log.Printf("cannot read remote request: %v", err)
		return
	}

	err = json.NewDecoder(bytes.NewReader(body)).Decode(&student)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "cannot understand request")
		log.Printf("got bad request: \n%s\n", body)
		return
	}

	student, err = s.db.CreateStudent(req.Context(), student)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "error creating student")
		log.Printf("error creating student: %v", err)
		return
	}

	err = json.NewEncoder(w).Encode(student)
	if err != nil {
		log.Printf("error writing create student response: %v", err)
	}
}

func (s *StudentServer) deleteStudent(w http.ResponseWriter, req *http.Request) {
	// DELETE /api/v1/students/<id>
	segments := strings.Split(req.URL.Path, "/")
	id := segments[len(segments)-1]
	student, err := s.db.DeleteStudent(req.Context(), id)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "error deleting student")
		log.Printf("error deleting students: %v", err)
		return
	}

	err = json.NewEncoder(w).Encode(student)
	if err != nil {
		log.Printf("error writing delete student response: %v", err)
	}
}

func main() {
	sqlDB, err := sql.Open("mysql", "root:my-secret-pw@tcp(127.0.0.1:3306)/students")
	if err != nil {
		log.Fatalf("error connecting to database: %v", err)
	}

	db := NewSQLStudentDatabase(sqlDB)

	srv := NewStudentServer(db)

	http.ListenAndServe(":2345", srv)
}
