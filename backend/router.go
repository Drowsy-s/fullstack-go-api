package main

import (
	"net/http"

	"fullstack-go-api/backend/internal/handlers"
	"fullstack-go-api/backend/internal/store"
)

// SetupRouter initializes the HTTP routes for the application.
func SetupRouter() http.Handler {
	store := store.New()
	handler := handlers.New(store)

	mux := http.NewServeMux()
	mux.HandleFunc("/api/register", handler.Register)
	mux.HandleFunc("/api/login", handler.Login)

	mux.Handle("/api/profile", handler.WithAuth(http.HandlerFunc(handler.Profile)))
	mux.Handle("/api/users", handler.WithAuth(http.HandlerFunc(handler.UsersCollection)))
	mux.Handle("/api/users/", handler.WithAuth(http.HandlerFunc(handler.UserResource)))

	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		respondJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	return mux
}

func respondJSON(w http.ResponseWriter, status int, payload interface{}) {
	handlers.WriteJSON(w, status, payload)
}
