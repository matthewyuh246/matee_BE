package main

import (
	"log"

	"github.com/joho/godotenv"
	"github.com/matthewyuh246/Matee/internal/controller"
	"github.com/matthewyuh246/Matee/internal/repository"
	"github.com/matthewyuh246/Matee/internal/router"
	"github.com/matthewyuh246/Matee/internal/usecase"
	db "github.com/matthewyuh246/Matee/pkg/database"
)

func main() {
	if err := godotenv.Load("../../.env"); err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	db := db.NewDB()
	userRepository := repository.NewUserRepository(db)
	userUsecase := usecase.NewUserUsecase(userRepository)
	userController := controller.NewUserController(userUsecase)
	e := router.NewRouter(userController)
	e.Logger.Fatal(e.Start(":8080"))
}
