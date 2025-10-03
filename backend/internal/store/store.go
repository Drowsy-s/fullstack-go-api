package store

import (
	"errors"
	"strings"
	"sync"
	"time"

	"fullstack-go-api/backend/internal/models"
)

// ErrUserNotFound indicates requested user does not exist.
var ErrUserNotFound = errors.New("user not found")

// ErrEmailExists indicates email already registered.
var ErrEmailExists = errors.New("email already registered")

// Store provides in-memory persistence for users.
type Store struct {
	mu     sync.RWMutex
	users  map[int]models.User
	byMail map[string]int
	nextID int
}

// New returns a new Store.
func New() *Store {
	return &Store{
		users:  make(map[int]models.User),
		byMail: make(map[string]int),
		nextID: 1,
	}
}

// CreateUser saves a new user and returns it.
func (s *Store) CreateUser(user models.User) (models.User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	key := strings.ToLower(user.Email)
	if _, exists := s.byMail[key]; exists {
		return models.User{}, ErrEmailExists
	}

	now := time.Now().UTC()
	user.ID = s.nextID
	user.CreatedAt = now
	user.UpdatedAt = now

	s.users[user.ID] = user
	s.byMail[key] = user.ID
	s.nextID++

	return user, nil
}

// UpdateUser updates an existing user by id.
func (s *Store) UpdateUser(id int, fn func(models.User) (models.User, error)) (models.User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	user, ok := s.users[id]
	if !ok {
		return models.User{}, ErrUserNotFound
	}

	updated, err := fn(user)
	if err != nil {
		return models.User{}, err
	}

	if !strings.EqualFold(user.Email, updated.Email) {
		// ensure new email unique
		newKey := strings.ToLower(updated.Email)
		if existingID, exists := s.byMail[newKey]; exists && existingID != user.ID {
			return models.User{}, ErrEmailExists
		}
		delete(s.byMail, strings.ToLower(user.Email))
		s.byMail[newKey] = user.ID
	}

	updated.ID = user.ID
	updated.CreatedAt = user.CreatedAt
	updated.UpdatedAt = time.Now().UTC()
	s.users[id] = updated
	return updated, nil
}

// GetUser fetches a user by id.
func (s *Store) GetUser(id int) (models.User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	user, ok := s.users[id]
	if !ok {
		return models.User{}, ErrUserNotFound
	}
	return user, nil
}

// GetUserByEmail returns user by email.
func (s *Store) GetUserByEmail(email string) (models.User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	id, ok := s.byMail[strings.ToLower(email)]
	if !ok {
		return models.User{}, ErrUserNotFound
	}
	return s.users[id], nil
}

// ListUsers returns all stored users.
func (s *Store) ListUsers() []models.User {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]models.User, 0, len(s.users))
	for _, user := range s.users {
		result = append(result, user)
	}
	return result
}

// DeleteUser removes a user by id.
func (s *Store) DeleteUser(id int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	user, ok := s.users[id]
	if !ok {
		return ErrUserNotFound
	}

	delete(s.users, id)
	delete(s.byMail, strings.ToLower(user.Email))
	return nil
}
