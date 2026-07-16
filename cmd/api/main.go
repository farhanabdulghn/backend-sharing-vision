package main

import (
	"log"
	"net/http"

	"postsapi/internal/config"
	"postsapi/internal/database"
	"postsapi/internal/handlers"
	"postsapi/internal/models"
)

func main() {
	cfg := config.Load()

	db, err := database.Connect(cfg.DSN())
	if err != nil {
		log.Fatalf("database connection failed: %v", err)
	}
	defer db.Close()

	repo := models.NewPostRepository(db)
	h := handlers.NewPostHandler(repo)

	mux := http.NewServeMux()

	// Create
	mux.HandleFunc("POST /article/", h.Create)

	// List with pagination
	mux.HandleFunc("GET /article/{limit}/{offset}", h.List)

	// Get one
	mux.HandleFunc("GET /article/{id}", h.Get)

	// Update (PUT and PATCH both accepted)
	mux.HandleFunc("PUT /article/{id}", h.Update)
	mux.HandleFunc("PATCH /article/{id}", h.Update)

	// Delete
	mux.HandleFunc("DELETE /article/{id}", h.Delete)

	log.Printf("listening on :%s", cfg.ServerPort)

	handler := enableCORS(logRequests(mux))

	if err := http.ListenAndServe(":"+cfg.ServerPort, handler); err != nil {
		log.Fatal(err)
	}
}

func enableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// logRequests is a tiny middleware that logs every incoming request.
func logRequests(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s", r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
	})
}
