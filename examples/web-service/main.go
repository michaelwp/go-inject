package main

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-inject/go-inject"
)

// Domain interfaces
type Logger interface {
	Info(message string)
	Error(message string)
}

type Database interface {
	GetUser(id int) (*User, error)
	CreateUser(user *User) error
	GetAllUsers() ([]*User, error)
}

type UserRepository interface {
	GetByID(id int) (*User, error)
	Create(user *User) error
	GetAll() ([]*User, error)
}

type UserService interface {
	GetUser(id int) (*User, error)
	CreateUser(name, email string) (*User, error)
	GetAllUsers() ([]*User, error)
}

// Domain models
type User struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

// Implementations
type ConsoleLogger struct{}

func (l *ConsoleLogger) Info(message string) {
	log.Printf("[INFO] %s", message)
}

func (l *ConsoleLogger) Error(message string) {
	log.Printf("[ERROR] %s", message)
}

type InMemoryDatabase struct {
	users  map[int]*User
	nextID int
	logger Logger
}

func (db *InMemoryDatabase) GetUser(id int) (*User, error) {
	db.logger.Info(fmt.Sprintf("Database: Getting user %d", id))
	if user, exists := db.users[id]; exists {
		return user, nil
	}
	return nil, fmt.Errorf("user with id %d not found", id)
}

func (db *InMemoryDatabase) CreateUser(user *User) error {
	db.logger.Info(fmt.Sprintf("Database: Creating user %s", user.Name))
	user.ID = db.nextID
	db.nextID++
	db.users[user.ID] = user
	return nil
}

func (db *InMemoryDatabase) GetAllUsers() ([]*User, error) {
	db.logger.Info("Database: Getting all users")
	users := make([]*User, 0, len(db.users))
	for _, user := range db.users {
		users = append(users, user)
	}
	return users, nil
}

type UserRepositoryImpl struct {
	db     Database
	logger Logger
}

func (r *UserRepositoryImpl) GetByID(id int) (*User, error) {
	r.logger.Info(fmt.Sprintf("Repository: Getting user %d", id))
	return r.db.GetUser(id)
}

func (r *UserRepositoryImpl) Create(user *User) error {
	r.logger.Info(fmt.Sprintf("Repository: Creating user %s", user.Name))
	return r.db.CreateUser(user)
}

func (r *UserRepositoryImpl) GetAll() ([]*User, error) {
	r.logger.Info("Repository: Getting all users")
	return r.db.GetAllUsers()
}

type UserServiceImpl struct {
	repo   UserRepository
	logger Logger
}

func (s *UserServiceImpl) GetUser(id int) (*User, error) {
	s.logger.Info(fmt.Sprintf("Service: Getting user %d", id))
	return s.repo.GetByID(id)
}

func (s *UserServiceImpl) CreateUser(name, email string) (*User, error) {
	s.logger.Info(fmt.Sprintf("Service: Creating user %s", name))
	user := &User{Name: name, Email: email}
	err := s.repo.Create(user)
	return user, err
}

func (s *UserServiceImpl) GetAllUsers() ([]*User, error) {
	s.logger.Info("Service: Getting all users")
	return s.repo.GetAll()
}

// HTTP Handlers
type UserHandler struct {
	userService UserService
	logger      Logger
}

func (h *UserHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimPrefix(r.URL.Path, "/users/")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		h.logger.Error(fmt.Sprintf("Invalid user ID: %s", idStr))
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	user, err := h.userService.GetUser(id)
	if err != nil {
		h.logger.Error(fmt.Sprintf("User not found: %v", err))
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"id": %d, "name": "%s", "email": "%s"}`, user.ID, user.Name, user.Email)
}

func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	name := r.FormValue("name")
	email := r.FormValue("email")

	if name == "" || email == "" {
		h.logger.Error("Missing name or email")
		http.Error(w, "Name and email required", http.StatusBadRequest)
		return
	}

	user, err := h.userService.CreateUser(name, email)
	if err != nil {
		h.logger.Error(fmt.Sprintf("Failed to create user: %v", err))
		http.Error(w, "Failed to create user", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, `{"id": %d, "name": "%s", "email": "%s"}`, user.ID, user.Name, user.Email)
}

func (h *UserHandler) GetAllUsers(w http.ResponseWriter, r *http.Request) {
	users, err := h.userService.GetAllUsers()
	if err != nil {
		h.logger.Error(fmt.Sprintf("Failed to get users: %v", err))
		http.Error(w, "Failed to get users", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	fmt.Fprint(w, "[")
	for i, user := range users {
		if i > 0 {
			fmt.Fprint(w, ",")
		}
		fmt.Fprintf(w, `{"id": %d, "name": "%s", "email": "%s"}`, user.ID, user.Name, user.Email)
	}
	fmt.Fprint(w, "]")
}

func setupContainer() *inject.Container {
	container := inject.NewContainer()

	// Register logger as singleton
	inject.RegisterSingletonInterface[Logger, *ConsoleLogger](container, func(c *inject.Container) *ConsoleLogger {
		return &ConsoleLogger{}
	})

	// Register database as singleton
	inject.RegisterSingletonInterface[Database, *InMemoryDatabase](container, func(c *inject.Container) *InMemoryDatabase {
		logger := inject.MustResolve[Logger](c)
		return &InMemoryDatabase{
			users:  make(map[int]*User),
			nextID: 1,
			logger: logger,
		}
	})

	// Register repository as singleton
	inject.RegisterSingletonInterface[UserRepository, *UserRepositoryImpl](container, func(c *inject.Container) *UserRepositoryImpl {
		db := inject.MustResolve[Database](c)
		logger := inject.MustResolve[Logger](c)
		return &UserRepositoryImpl{db: db, logger: logger}
	})

	// Register service as singleton
	inject.RegisterSingletonInterface[UserService, *UserServiceImpl](container, func(c *inject.Container) *UserServiceImpl {
		repo := inject.MustResolve[UserRepository](c)
		logger := inject.MustResolve[Logger](c)
		return &UserServiceImpl{repo: repo, logger: logger}
	})

	// Register HTTP handler as transient (though typically would be singleton in real apps)
	inject.RegisterTransientType[*UserHandler](container, func(c *inject.Container) *UserHandler {
		userService := inject.MustResolve[UserService](c)
		logger := inject.MustResolve[Logger](c)
		return &UserHandler{userService: userService, logger: logger}
	})

	return container
}

func main() {
	fmt.Println("=== Web Service with Dependency Injection Example ===")

	container := setupContainer()

	// Get handler from container
	handler := inject.MustResolve[*UserHandler](container)

	// Setup routes
	http.HandleFunc("/users", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			handler.GetAllUsers(w, r)
		} else if r.Method == http.MethodPost {
			handler.CreateUser(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	http.HandleFunc("/users/", handler.GetUser)

	// Create some sample data
	userService := inject.MustResolve[UserService](container)
	userService.CreateUser("Alice Johnson", "alice@example.com")
	userService.CreateUser("Bob Smith", "bob@example.com")

	fmt.Println("Server starting on :8080")
	fmt.Println("Try these endpoints:")
	fmt.Println("  GET  http://localhost:8080/users")
	fmt.Println("  GET  http://localhost:8080/users/1")
	fmt.Println("  POST http://localhost:8080/users (with form data: name=John&email=john@example.com)")

	server := &http.Server{
		Addr:         ":8080",
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	log.Fatal(server.ListenAndServe())
}
