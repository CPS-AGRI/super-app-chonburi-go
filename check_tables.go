package main

import (
	"fmt"
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	dsn := "postgres://uat_admin:C807BI%2B%2BwFizSBH%2FFkNgL7J3qW646npW@122.155.169.235:5489/uat_chonburi?sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal(err)
	}

	var tables []string
	db.Raw("SELECT table_name FROM information_schema.tables WHERE table_schema = 'public'").Scan(&tables)

	fmt.Println("Tables in DB:")
	for _, t := range tables {
		fmt.Println("-", t)
	}
}
