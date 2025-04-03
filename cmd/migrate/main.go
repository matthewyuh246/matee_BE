package main

import (
	"fmt"

	"github.com/matthewyuh246/Matee/internal/domain"
	db "github.com/matthewyuh246/Matee/pkg/database"
)

func main() {
	dbConn := db.NewDB()
	defer fmt.Println("Successfully Migrated")
	defer db.CloseDB(dbConn)
	dbConn.AutoMigrate(&domain.User{})
}
