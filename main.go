package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/theluckiestsoul/employeemanager/database"
	"github.com/theluckiestsoul/employeemanager/handlers"

	httpSwagger "github.com/swaggo/http-swagger/v2"
	_ "github.com/theluckiestsoul/employeemanager/docs"
)

// @title Employee Manager API
// @version 1.0
// @description This is a server for the Employee Manager API.
// @termsOfService  http://swagger.io/terms/

// @contact.name   Kiran Kumar Mohanty
// @contact.email  kiranmohanty.remote@gmail.com

// @host      localhost:8080
// @BasePath  /api/v1
func main() {
	cfg, err := loadConfig()
	if err != nil {
		log.Fatal(err)
	}
	db, err := database.NewDatabase(cfg.DbURL)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	if err := database.Initialize(db); err != nil {
		log.Fatal(err)
	}

	empDB := database.NewEmployee(db)

	h := handlers.NewHandler(empDB)

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)

	r.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL("/swagger/doc.json"),
	))

	r.Route("/api/v1/employees", func(r chi.Router) {
		r.Post("/", h.CreateEmployeeHandler)
		r.Get("/", h.ListEmployeesHandler)
		r.Route("/{id}", func(r chi.Router) {
			r.Get("/", h.GetEmployeeHandler)
			r.Put("/", h.UpdateEmployeeHandler)
			r.Delete("/", h.DeleteEmployeeHandler)
		})
	})

	server := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: r,
	}

	go func() {
		log.Printf("Server started on port %s\n", cfg.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	sigs := make(chan os.Signal, 1)

	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	sig := <-sigs
	log.Printf("Received signal: %s \n", sig)
	log.Println("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Failed to shutdown server: %v", err)
	}
}
