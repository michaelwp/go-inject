package main

import (
	"fmt"
	"testing"

	"github.com/go-inject/go-inject"
)

// Domain interfaces
type EmailService interface {
	SendEmail(to, subject, body string) error
}

type UserRepository interface {
	Save(user *User) error
	FindByEmail(email string) (*User, error)
}

type Logger interface {
	Log(message string)
}

// Domain model
type User struct {
	ID    int
	Name  string
	Email string
}

// Production implementations
type SMTPEmailService struct {
	host string
	port int
}

func (s *SMTPEmailService) SendEmail(to, subject, body string) error {
	fmt.Printf("SMTP: Sending email to %s: %s\n", to, subject)
	return nil
}

type DatabaseUserRepository struct {
	connectionString string
}

func (r *DatabaseUserRepository) Save(user *User) error {
	fmt.Printf("DB: Saving user %s to database\n", user.Name)
	return nil
}

func (r *DatabaseUserRepository) FindByEmail(email string) (*User, error) {
	fmt.Printf("DB: Finding user by email %s\n", email)
	return &User{ID: 1, Name: "Found User", Email: email}, nil
}

type ConsoleLogger struct{}

func (l *ConsoleLogger) Log(message string) {
	fmt.Printf("LOG: %s\n", message)
}

// Service under test
type UserRegistrationService struct {
	userRepo     UserRepository
	emailService EmailService
	logger       Logger
}

func (s *UserRegistrationService) RegisterUser(name, email string) error {
	s.logger.Log(fmt.Sprintf("Starting registration for %s", email))

	// Check if user already exists
	existingUser, err := s.userRepo.FindByEmail(email)
	if err == nil && existingUser != nil {
		return fmt.Errorf("user with email %s already exists", email)
	}

	// Create new user
	user := &User{Name: name, Email: email}
	if err := s.userRepo.Save(user); err != nil {
		return fmt.Errorf("failed to save user: %w", err)
	}

	// Send welcome email
	welcomeMessage := fmt.Sprintf("Welcome to our service, %s!", name)
	if err := s.emailService.SendEmail(email, "Welcome!", welcomeMessage); err != nil {
		s.logger.Log(fmt.Sprintf("Failed to send welcome email to %s: %v", email, err))
		// Don't fail registration if email fails
	}

	s.logger.Log(fmt.Sprintf("Successfully registered user %s", email))
	return nil
}

// Mock implementations for testing
type MockEmailService struct {
	sentEmails []EmailCall
}

type EmailCall struct {
	To      string
	Subject string
	Body    string
}

func (m *MockEmailService) SendEmail(to, subject, body string) error {
	m.sentEmails = append(m.sentEmails, EmailCall{
		To:      to,
		Subject: subject,
		Body:    body,
	})
	return nil
}

func (m *MockEmailService) GetSentEmails() []EmailCall {
	return m.sentEmails
}

type MockUserRepository struct {
	users  map[string]*User
	nextID int
}

func (m *MockUserRepository) Save(user *User) error {
	user.ID = m.nextID
	m.nextID++
	m.users[user.Email] = user
	return nil
}

func (m *MockUserRepository) FindByEmail(email string) (*User, error) {
	if user, exists := m.users[email]; exists {
		return user, nil
	}
	return nil, fmt.Errorf("user not found")
}

func (m *MockUserRepository) GetUsers() map[string]*User {
	return m.users
}

type MockLogger struct {
	logs []string
}

func (m *MockLogger) Log(message string) {
	m.logs = append(m.logs, message)
}

func (m *MockLogger) GetLogs() []string {
	return m.logs
}

// Test functions
func TestUserRegistrationService_Success(t *testing.T) {
	// Setup container with mocks
	container := inject.NewContainer()

	mockEmail := &MockEmailService{}
	mockRepo := &MockUserRepository{users: make(map[string]*User), nextID: 1}
	mockLogger := &MockLogger{}

	// Register mocks
	inject.RegisterValue[EmailService](container, mockEmail)
	inject.RegisterValue[UserRepository](container, mockRepo)
	inject.RegisterValue[Logger](container, mockLogger)

	// Register service under test
	inject.RegisterTransientType[*UserRegistrationService](container, func(c *inject.Container) *UserRegistrationService {
		return &UserRegistrationService{
			userRepo:     inject.MustResolve[UserRepository](c),
			emailService: inject.MustResolve[EmailService](c),
			logger:       inject.MustResolve[Logger](c),
		}
	})

	// Test
	service := inject.MustResolve[*UserRegistrationService](container)
	err := service.RegisterUser("John Doe", "john@example.com")

	// Assertions
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Check user was saved
	users := mockRepo.GetUsers()
	if len(users) != 1 {
		t.Fatalf("Expected 1 user, got %d", len(users))
	}

	savedUser := users["john@example.com"]
	if savedUser.Name != "John Doe" {
		t.Errorf("Expected user name 'John Doe', got '%s'", savedUser.Name)
	}

	// Check email was sent
	emails := mockEmail.GetSentEmails()
	if len(emails) != 1 {
		t.Fatalf("Expected 1 email, got %d", len(emails))
	}

	if emails[0].To != "john@example.com" {
		t.Errorf("Expected email to 'john@example.com', got '%s'", emails[0].To)
	}

	if emails[0].Subject != "Welcome!" {
		t.Errorf("Expected subject 'Welcome!', got '%s'", emails[0].Subject)
	}

	// Check logs
	logs := mockLogger.GetLogs()
	if len(logs) < 2 {
		t.Fatalf("Expected at least 2 log entries, got %d", len(logs))
	}
}

func TestUserRegistrationService_DuplicateEmail(t *testing.T) {
	container := inject.NewContainer()

	mockRepo := &MockUserRepository{users: make(map[string]*User), nextID: 1}
	mockLogger := &MockLogger{}

	// Pre-populate with existing user
	mockRepo.Save(&User{Name: "Existing User", Email: "john@example.com"})

	inject.RegisterValue[UserRepository](container, mockRepo)
	inject.RegisterValue[Logger](container, mockLogger)
	inject.RegisterValue[EmailService](container, &MockEmailService{})

	inject.RegisterTransientType[*UserRegistrationService](container, func(c *inject.Container) *UserRegistrationService {
		return &UserRegistrationService{
			userRepo:     inject.MustResolve[UserRepository](c),
			emailService: inject.MustResolve[EmailService](c),
			logger:       inject.MustResolve[Logger](c),
		}
	})

	// Test
	service := inject.MustResolve[*UserRegistrationService](container)
	err := service.RegisterUser("John Doe", "john@example.com")

	// Should fail due to duplicate email
	if err == nil {
		t.Fatal("Expected error for duplicate email, got nil")
	}

	expectedError := "user with email john@example.com already exists"
	if err.Error() != expectedError {
		t.Errorf("Expected error '%s', got '%s'", expectedError, err.Error())
	}
}

func setupProductionContainer() *inject.Container {
	container := inject.NewContainer()

	// Register production implementations
	inject.RegisterSingletonInterface[Logger, *ConsoleLogger](container, func(c *inject.Container) *ConsoleLogger {
		return &ConsoleLogger{}
	})

	inject.RegisterSingletonInterface[EmailService, *SMTPEmailService](container, func(c *inject.Container) *SMTPEmailService {
		return &SMTPEmailService{host: "smtp.example.com", port: 587}
	})

	inject.RegisterSingletonInterface[UserRepository, *DatabaseUserRepository](container, func(c *inject.Container) *DatabaseUserRepository {
		return &DatabaseUserRepository{connectionString: "host=localhost dbname=users"}
	})

	inject.RegisterTransientType[*UserRegistrationService](container, func(c *inject.Container) *UserRegistrationService {
		return &UserRegistrationService{
			userRepo:     inject.MustResolve[UserRepository](c),
			emailService: inject.MustResolve[EmailService](c),
			logger:       inject.MustResolve[Logger](c),
		}
	})

	return container
}

func main() {
	fmt.Println("=== Testing with Dependency Injection Example ===")

	// Run the tests
	fmt.Println("\nRunning tests...")

	testing.Main(func(pat, str string) (bool, error) { return true, nil },
		[]testing.InternalTest{
			{
				Name: "TestUserRegistrationService_Success",
				F:    TestUserRegistrationService_Success,
			},
			{
				Name: "TestUserRegistrationService_DuplicateEmail",
				F:    TestUserRegistrationService_DuplicateEmail,
			},
		},
		[]testing.InternalBenchmark{},
		[]testing.InternalExample{})

	fmt.Println("\nTests completed!")

	// Demonstrate production usage
	fmt.Println("\n=== Production Usage ===")
	container := setupProductionContainer()
	service := inject.MustResolve[*UserRegistrationService](container)

	err := service.RegisterUser("Alice Johnson", "alice@production.com")
	if err != nil {
		fmt.Printf("Registration failed: %v\n", err)
	} else {
		fmt.Println("User registered successfully in production mode!")
	}
}
