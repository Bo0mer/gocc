package main

import (
	"database/sql"
	"errors"
	"fmt"
	"log"

	_ "github.com/go-sql-driver/mysql" // side effects
)

func exampleDatabaseSQL() {
	db, err := sql.Open("mysql", "root:my-secret-pw@tcp(127.0.0.1:3306)/students")
	if err != nil {
		log.Fatalf("error connecting to database: %v", err)
	}

	err = db.Ping()
	if err != nil {
		log.Fatalf("error pinging database: %v", err)
	}

	rows, err := db.Query("SELECT * FROM students")
	if err != nil {
		log.Fatalf("error during query: %v", err)
	}
	defer rows.Close()
	for rows.Next() {
		var id int
		var name string
		var age int
		err = rows.Scan(&id, &name, &age)
		if err != nil {
			log.Fatalf("error during scan: %v", err)
		}

		fmt.Println(id, name, age)
	}

	row := db.QueryRow("SELECT * FROM students WHERE id = ?", 5)

	var s Student
	err = row.Scan(&s.ID, &s.Name, &s.Age)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			fmt.Println("student not found")
		} else {
			log.Fatalf("error during scan of single row: %v", err)
		}
	}

	fmt.Println(s)

	result, err := db.Exec(
		"INSERT INTO students (name, age) VALUES (?, ?)",
		"gopher 2",
		27,
	)
	if err != nil {
		log.Fatalf("error during exec: %v", err)
	}
	fmt.Println(result.LastInsertId())
	fmt.Println(result.RowsAffected())
}
