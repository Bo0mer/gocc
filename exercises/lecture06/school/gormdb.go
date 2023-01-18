package main

import (
	"fmt"
	"log"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func exampleGORM() {
	const dsn = "root:my-secret-pw@tcp(127.0.0.1:3306)/students"

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("error connecting to database: %v", err)
	}

	var students []Student
	err = db.Find(&students).Error
	if err != nil {
		log.Fatalf("error during find: %v", err)
	}

	fmt.Println(students)

	gormer := Student{
		Name: "gormer",
		Age:  11,
	}
	err = db.Create(&gormer).Error
	if err != nil {
		log.Fatalf("error during create: %v", err)
	}

	fmt.Println(gormer)

	var specific Student
	err = db.Where("id = ?", 1).Take(&specific).Error
	if err != nil {
		log.Fatalf("error during take: %v", err)
	}
	fmt.Println(specific)

	var all []Student
	err = db.Raw("SELECT * FROM students").Scan(&all).Error
	if err != nil {
		log.Fatalf("error during raw scan: %v", err)
	}
	fmt.Println(all)
}
