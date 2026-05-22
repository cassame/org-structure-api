package main

import (
	"apiOrganizationStructure/internal/config"
	"apiOrganizationStructure/internal/handler"
	"apiOrganizationStructure/internal/repository"
	"apiOrganizationStructure/internal/service"
	"log"
	"net/http"

	"github.com/joho/godotenv"
)

func main() {

	if err := godotenv.Load(); err != nil {
		log.Println("Warning: No .env file found, using system environment variables")
	}

	cfg := config.Load()

	db, err := repository.InitDB(cfg.DBDSN, "./migrations")
	if err != nil {
		log.Fatalf("Database initialization failed: %v", err)
	}
	_ = db

	deptRepo := repository.NewDepartmentRepository(db)
	empRepo := repository.NewEmployeeRepository(db)

	deptService := service.NewDepartmentService(deptRepo)
	empService := service.NewEmployeeService(empRepo, deptRepo)

	deptHandler := handler.NewDepartmentHandler(deptService)
	empHandler := handler.NewEmployeeHandler(empService)

	mux := http.NewServeMux()

	mux.HandleFunc("POST /api/departments", deptHandler.CreateDepartment)
	mux.HandleFunc("POST /api/departments/{id}/employees", empHandler.CreateEmployeeInDepartment)
	mux.HandleFunc("POST /api/employees", empHandler.CreateEmployee)

	mux.HandleFunc("GET /api/departments", deptHandler.GetTree)
	mux.HandleFunc("GET /api/departments/{id}", deptHandler.GetDepartmentByID)

	mux.HandleFunc("PATCH /api/departments/{id}", deptHandler.UpdateDepartment)
	
	mux.HandleFunc("DELETE /api/departments/{id}", deptHandler.DeleteDepartment)

	log.Printf("Starting server on port %s...", cfg.Port)
	loggedMux := handler.LoggerMiddleware(mux)
	if err := http.ListenAndServe(":"+cfg.Port, loggedMux); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
