package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"fullstack-go-api/backend/internal/auth"
	"fullstack-go-api/backend/internal/models"
	"fullstack-go-api/backend/internal/store"
)

type contextKey string

const userContextKey contextKey = "user"

// Handler bundles dependencies for HTTP handlers.
type Handler struct {
	Store     *store.Store
	JWTSecret string
	TokenTTL  time.Duration
}

// New creates a Handler with sane defaults.
func New(store *store.Store) *Handler {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "supersecretkey"
	}
	return &Handler{Store: store, JWTSecret: secret, TokenTTL: 24 * time.Hour}
}

// WriteJSON serializes payload into JSON response.
func WriteJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if payload == nil {
		return
	}
	encoder := json.NewEncoder(w)
	_ = encoder.Encode(payload)
}

func errorResponse(w http.ResponseWriter, status int, message string) {
	WriteJSON(w, status, map[string]string{"error": message})
}

type registerRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type updateRequest struct {
	Name     *string `json:"name"`
	Email    *string `json:"email"`
	Password *string `json:"password"`
}

// Register handles user creation.
func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", http.MethodPost)
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var req registerRequest
	if err := decodeJSON(r, &req); err != nil {
		errorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	name := strings.TrimSpace(req.Name)
	email := strings.TrimSpace(req.Email)
	password := req.Password

	if name == "" || email == "" || password == "" {
		errorResponse(w, http.StatusBadRequest, "name, email, and password are required")
		return
	}

	if !strings.Contains(email, "@") {
		errorResponse(w, http.StatusBadRequest, "email must be valid")
		return
	}

	if len(password) < 6 {
		errorResponse(w, http.StatusBadRequest, "password must be at least 6 characters")
		return
	}

	hashed := auth.HashPassword(password)
	user := models.User{
		Name:         name,
		Email:        email,
		PasswordHash: hashed,
	}

	created, err := h.Store.CreateUser(user)
	if err != nil {
		if errors.Is(err, store.ErrEmailExists) {
			errorResponse(w, http.StatusConflict, "email already registered")
			return
		}
		errorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	token, err := auth.GenerateToken(created.ID, created.Email, h.JWTSecret, h.TokenTTL)
	if err != nil {
		errorResponse(w, http.StatusInternalServerError, "failed to create token")
		return
	}

	WriteJSON(w, http.StatusCreated, map[string]interface{}{
		"user":  created.Sanitized(),
		"token": token,
	})
}

// Login handles user authentication.
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", http.MethodPost)
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var req loginRequest
	if err := decodeJSON(r, &req); err != nil {
		errorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	email := strings.TrimSpace(req.Email)
	password := req.Password

	if email == "" || password == "" {
		errorResponse(w, http.StatusBadRequest, "email and password are required")
		return
	}

	user, err := h.Store.GetUserByEmail(email)
	if err != nil || !auth.ComparePassword(user.PasswordHash, password) {
		errorResponse(w, http.StatusUnauthorized, "invalid credentials")
		return
	}

	token, err := auth.GenerateToken(user.ID, user.Email, h.JWTSecret, h.TokenTTL)
	if err != nil {
		errorResponse(w, http.StatusInternalServerError, "failed to create token")
		return
	}

	WriteJSON(w, http.StatusOK, map[string]interface{}{
		"user":  user.Sanitized(),
		"token": token,
	})
}

// UsersCollection handles requests to /api/users.
func (h *Handler) UsersCollection(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	users := h.Store.ListUsers()
	sanitized := make([]models.User, 0, len(users))
	for _, user := range users {
		sanitized = append(sanitized, user.Sanitized())
	}

	WriteJSON(w, http.StatusOK, map[string]interface{}{"users": sanitized})
}

// UserResource handles requests to /api/users/{id}.
func (h *Handler) UserResource(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimPrefix(r.URL.Path, "/api/users/")
	if idStr == "" {
		errorResponse(w, http.StatusNotFound, "user not found")
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		errorResponse(w, http.StatusBadRequest, "invalid user id")
		return
	}

	switch r.Method {
	case http.MethodGet:
		user, err := h.Store.GetUser(id)
		if err != nil {
			errorResponse(w, http.StatusNotFound, "user not found")
			return
		}
		WriteJSON(w, http.StatusOK, map[string]interface{}{"user": user.Sanitized()})
	case http.MethodPut:
		var req updateRequest
		if err := decodeJSON(r, &req); err != nil {
			errorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		if req.Name != nil {
			trimmed := strings.TrimSpace(*req.Name)
			if trimmed == "" {
				errorResponse(w, http.StatusBadRequest, "name cannot be empty")
				return
			}
			req.Name = &trimmed
		}
		if req.Email != nil {
			trimmedEmail := strings.TrimSpace(*req.Email)
			if trimmedEmail == "" || !strings.Contains(trimmedEmail, "@") {
				errorResponse(w, http.StatusBadRequest, "email must be valid")
				return
			}
			req.Email = &trimmedEmail
		}
		if req.Password != nil && *req.Password != "" && len(*req.Password) < 6 {
			errorResponse(w, http.StatusBadRequest, "password must be at least 6 characters")
			return
		}

		updated, err := h.Store.UpdateUser(id, func(existing models.User) (models.User, error) {
			if req.Name != nil {
				existing.Name = *req.Name
			}
			if req.Email != nil {
				existing.Email = *req.Email
			}
			if req.Password != nil && *req.Password != "" {
				existing.PasswordHash = auth.HashPassword(*req.Password)
			}
			return existing, nil
		})
		if err != nil {
			switch {
			case errors.Is(err, store.ErrUserNotFound):
				errorResponse(w, http.StatusNotFound, "user not found")
			case errors.Is(err, store.ErrEmailExists):
				errorResponse(w, http.StatusConflict, "email already registered")
			default:
				errorResponse(w, http.StatusInternalServerError, err.Error())
			}
			return
		}
		WriteJSON(w, http.StatusOK, map[string]interface{}{"user": updated.Sanitized()})
	case http.MethodDelete:
		if err := h.Store.DeleteUser(id); err != nil {
			errorResponse(w, http.StatusNotFound, "user not found")
			return
		}
		w.WriteHeader(http.StatusNoContent)
	default:
		w.Header().Set("Allow", strings.Join([]string{http.MethodGet, http.MethodPut, http.MethodDelete}, ", "))
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

// Profile returns the authenticated user's details.
func (h *Handler) Profile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	user, ok := GetUserFromContext(r.Context())
	if !ok {
		errorResponse(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	WriteJSON(w, http.StatusOK, map[string]interface{}{"user": user.Sanitized()})
}

// WithAuth ensures a request is authenticated before invoking the handler.
func (h *Handler) WithAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			errorResponse(w, http.StatusUnauthorized, "missing authorization header")
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			errorResponse(w, http.StatusUnauthorized, "invalid authorization header")
			return
		}

		userID, email, err := auth.ParseToken(parts[1], h.JWTSecret)
		if err != nil {
			errorResponse(w, http.StatusUnauthorized, "invalid token")
			return
		}

		user, err := h.Store.GetUser(userID)
		if err != nil || !strings.EqualFold(user.Email, email) {
			errorResponse(w, http.StatusUnauthorized, "invalid token")
			return
		}

		ctx := context.WithValue(r.Context(), userContextKey, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetUserFromContext retrieves the authenticated user from the context.
func GetUserFromContext(ctx context.Context) (models.User, bool) {
	user, ok := ctx.Value(userContextKey).(models.User)
	return user, ok
}

func decodeJSON(r *http.Request, v interface{}) error {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(v); err != nil {
		return err
	}
	return nil
}
