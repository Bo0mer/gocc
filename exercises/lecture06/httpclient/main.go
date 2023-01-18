package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
)

type Student struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Age  int    `json:"age"`
}

type StudentClient struct {
	api        string // http://localhost:2345
	httpClient *http.Client
}

func NewStudentClient(api string, httpClient *http.Client) *StudentClient {
	return &StudentClient{
		api:        strings.TrimSuffix(api, "/"),
		httpClient: httpClient,
	}
}

func (c *StudentClient) ListStudents(ctx context.Context) ([]Student, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", c.api+"/api/v1/students", nil)
	if err != nil {
		return nil, fmt.Errorf("error constructing req: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error during do req: %w", err)
	}
	defer resp.Body.Close()

	var students []Student
	err = json.NewDecoder(resp.Body).Decode(&students)
	if err != nil {
		return nil, fmt.Errorf("error during decode: %w", err)
	}

	return students, nil
}

func (c *StudentClient) CreateStudent(ctx context.Context, s Student) (Student, error) {
	body := &bytes.Buffer{}
	err := json.NewEncoder(body).Encode(s)
	if err != nil {
		return Student{}, fmt.Errorf("error during encode: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.api+"/api/v1/students", body)
	if err != nil {
		return Student{}, fmt.Errorf("error constructing req: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return Student{}, fmt.Errorf("error during do req: %w", err)
	}
	defer resp.Body.Close()

	var student Student
	err = json.NewDecoder(resp.Body).Decode(&student)
	if err != nil {
		return Student{}, fmt.Errorf("error during decode: %w", err)
	}

	return student, nil
}

func main() {
	c := NewStudentClient("http://localhost:2345", &http.Client{})

	ctx := context.Background()

	student, err := c.CreateStudent(ctx, Student{Name: "Programmer", Age: 1})
	if err != nil {
		log.Fatalf("error creating student: %v", err)
	}

	fmt.Println("Created student", student)
}
