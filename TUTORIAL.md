# Go-Inject Tutorial: Complete Guide to Dependency Injection in Go

Welcome to the comprehensive tutorial for Go-Inject! This guide will take you from basic concepts to advanced usage patterns, helping you master dependency injection in Go.

## Table of Contents

1. [What is Dependency Injection?](#what-is-dependency-injection)
2. [Why Use Dependency Injection?](#why-use-dependency-injection)
3. [Getting Started](#getting-started)
4. [Basic Usage](#basic-usage)
5. [Service Lifetimes](#service-lifetimes)
6. [Interface-based Design](#interface-based-design)
7. [Dependency Injection Patterns](#dependency-injection-patterns)
8. [Error Handling](#error-handling)
9. [Testing with DI](#testing-with-di)
10. [Advanced Topics](#advanced-topics)
11. [Best Practices](#best-practices)
12. [Common Patterns](#common-patterns)

## What is Dependency Injection?

Dependency Injection (DI) is a design pattern where objects receive their dependencies from external sources rather than creating them internally. Instead of a class instantiating its dependencies directly, they are "injected" from the outside.

### Without Dependency Injection

```go
type UserService struct {
    logger *Logger  // Hard-coded dependency
}

func NewUserService() *UserService {
    return &UserService{
        logger: &Logger{},  // Tightly coupled
    }
}
```

### With Dependency Injection

```go
type UserService struct {
    logger Logger  // Interface dependency
}

func NewUserService(logger Logger) *UserService {
    return &UserService{
        logger: logger,  // Injected dependency
    }
}
```

## Why Use Dependency Injection?

### 1. **Testability**
Easily substitute real dependencies with mocks during testing.

### 2. **Flexibility**
Swap implementations without changing dependent code.

### 3. **Loose Coupling**
Reduce dependencies between components.

### 4. **Single Responsibility**
Components focus on their core logic, not dependency management.

### 5. **Configuration Management**
Centralize dependency configuration.

## Getting Started

### Installation

```bash
go mod init your-project
go get github.com/go-inject/go-inject
```

### Basic Setup

```go
package main

import "github.com/go-inject/go-inject"

func main() {
    container := inject.NewContainer()
    // Register and resolve services here
}
```

## Basic Usage

### Step 1: Define Your Services

```go
type Logger interface {
    Log(message string)
}

type ConsoleLogger struct{}

func (c *ConsoleLogger) Log(message string) {
    fmt.Println("LOG:", message)
}

type UserService struct {
    logger Logger
}

func (u *UserService) CreateUser(name string) {
    u.logger.Log("Creating user: " + name)
}
```

### Step 2: Register Services

```go
container := inject.NewContainer()

// Register the logger implementation
inject.RegisterSingletonInterface[Logger, *ConsoleLogger](container, 
    func(c *inject.Container) *ConsoleLogger {
        return &ConsoleLogger{}
    })

// Register the user service
inject.RegisterTransientType[*UserService](container, 
    func(c *inject.Container) *UserService {
        logger := inject.MustResolve[Logger](c)
        return &UserService{logger: logger}
    })
```

### Step 3: Resolve and Use Services

```go
userService := inject.MustResolve[*UserService](container)
userService.CreateUser("Alice")
```

## Service Lifetimes

Go-Inject supports two service lifetimes:

### Singleton
- **One instance per container**
- **Created on first request**
- **Shared across all resolutions**

```go
inject.RegisterSingletonType[*DatabaseConnection](container, 
    func(c *inject.Container) *DatabaseConnection {
        return &DatabaseConnection{url: "db://localhost"}
    })
```

Use for:
- Database connections
- Configuration objects
- Expensive-to-create services
- Stateless services

### Transient
- **New instance on every request**
- **Independent instances**

```go
inject.RegisterTransientType[*OrderProcessor](container, 
    func(c *inject.Container) *OrderProcessor {
        return &OrderProcessor{}
    })
```

Use for:
- Stateful services
- Request-scoped objects
- Lightweight objects

### Comparison Example

```go
func demonstrateLifetimes() {
    container := inject.NewContainer()

    // Singleton - same instance
    inject.RegisterSingletonType[*SingletonService](container, 
        func(c *inject.Container) *SingletonService {
            return &SingletonService{id: time.Now().UnixNano()}
        })

    // Transient - different instances
    inject.RegisterTransientType[*TransientService](container, 
        func(c *inject.Container) *TransientService {
            return &TransientService{id: time.Now().UnixNano()}
        })

    // Test singleton behavior
    s1 := inject.MustResolve[*SingletonService](container)
    s2 := inject.MustResolve[*SingletonService](container)
    fmt.Printf("Singleton: s1.id == s2.id? %t\n", s1.id == s2.id) // true

    // Test transient behavior
    t1 := inject.MustResolve[*TransientService](container)
    t2 := inject.MustResolve[*TransientService](container)
    fmt.Printf("Transient: t1.id == t2.id? %t\n", t1.id == t2.id) // false
}
```

## Interface-based Design

### Why Interfaces?

Interfaces provide the foundation for flexible, testable code:

```go
// Define behavior, not implementation
type PaymentProcessor interface {
    ProcessPayment(amount float64) error
}

type NotificationService interface {
    SendNotification(message string) error
}

type OrderService struct {
    payment PaymentProcessor
    notifier NotificationService
}
```

### Multiple Implementations

```go
// Production implementation
type StripePaymentProcessor struct {
    apiKey string
}

func (s *StripePaymentProcessor) ProcessPayment(amount float64) error {
    // Stripe API call
    return nil
}

// Test implementation
type MockPaymentProcessor struct {
    shouldFail bool
}

func (m *MockPaymentProcessor) ProcessPayment(amount float64) error {
    if m.shouldFail {
        return errors.New("payment failed")
    }
    return nil
}
```

### Registration Patterns

```go
// Production container
func setupProduction() *inject.Container {
    container := inject.NewContainer()
    
    inject.RegisterSingletonInterface[PaymentProcessor, *StripePaymentProcessor](container,
        func(c *inject.Container) *StripePaymentProcessor {
            return &StripePaymentProcessor{apiKey: "sk_live_..."}
        })
    
    return container
}

// Test container
func setupTest() *inject.Container {
    container := inject.NewContainer()
    
    inject.RegisterSingletonInterface[PaymentProcessor, *MockPaymentProcessor](container,
        func(c *inject.Container) *MockPaymentProcessor {
            return &MockPaymentProcessor{shouldFail: false}
        })
    
    return container
}
```

## Dependency Injection Patterns

### Constructor Injection

The most common pattern - dependencies provided via constructor:

```go
type EmailService struct {
    smtp SMTPClient
    logger Logger
}

inject.RegisterTransientType[*EmailService](container, 
    func(c *inject.Container) *EmailService {
        smtp := inject.MustResolve[SMTPClient](c)
        logger := inject.MustResolve[Logger](c)
        return &EmailService{smtp: smtp, logger: logger}
    })
```

### Method Injection

Less common - dependencies provided via methods:

```go
type ReportGenerator struct {
    data DataSource
}

func (r *ReportGenerator) SetDataSource(data DataSource) {
    r.data = data
}

inject.RegisterTransientType[*ReportGenerator](container, 
    func(c *inject.Container) *ReportGenerator {
        generator := &ReportGenerator{}
        generator.SetDataSource(inject.MustResolve[DataSource](c))
        return generator
    })
```

### Property Injection

Setting dependencies via struct fields:

```go
type ServiceWithOptionalDeps struct {
    Required Logger
    Optional *CacheService // May be nil
}

inject.RegisterTransientType[*ServiceWithOptionalDeps](container, 
    func(c *inject.Container) *ServiceWithOptionalDeps {
        service := &ServiceWithOptionalDeps{
            Required: inject.MustResolve[Logger](c),
        }
        
        // Optional dependency
        if cache, ok := inject.TryResolve[*CacheService](c); ok {
            service.Optional = cache
        }
        
        return service
    })
```

## Error Handling

### Factory Functions with Errors

```go
inject.RegisterSingleton((*DatabaseConnection)(nil), 
    func() (*DatabaseConnection, error) {
        conn, err := sql.Open("postgres", "connection-string")
        if err != nil {
            return nil, fmt.Errorf("failed to connect to database: %w", err)
        }
        return &DatabaseConnection{conn: conn}, nil
    })

// Resolution will return the error
db, err := container.Resolve((*DatabaseConnection)(nil))
if err != nil {
    log.Fatalf("Database connection failed: %v", err)
}
```

### Graceful Error Handling

```go
func setupWithErrorHandling() *inject.Container {
    container := inject.NewContainer()
    
    // Register critical services with error handling
    err := inject.RegisterSingletonType[*CriticalService](container, 
        func(c *inject.Container) *CriticalService {
            return &CriticalService{}
        })
    if err != nil {
        panic(fmt.Sprintf("Failed to register critical service: %v", err))
    }
    
    return container
}
```

### Safe Resolution

```go
// Use TryResolve for optional dependencies
if cache, ok := inject.TryResolve[*CacheService](container); ok {
    // Use cache
    cache.Set("key", "value")
} else {
    // Fallback behavior
    log.Println("Cache not available, using fallback")
}

// Use MustResolve for required dependencies
logger := inject.MustResolve[Logger](container) // Panics if not found
```

## Testing with DI

### Test Setup Pattern

```go
func setupTestContainer() *inject.Container {
    container := inject.NewContainer()
    
    // Register mocks
    inject.RegisterValue[Logger](container, &MockLogger{})
    inject.RegisterValue[Database](container, &MockDatabase{})
    
    // Register service under test
    inject.RegisterTransientType[*UserService](container, 
        func(c *inject.Container) *UserService {
            return &UserService{
                logger: inject.MustResolve[Logger](c),
                db:     inject.MustResolve[Database](c),
            }
        })
    
    return container
}
```

### Mock Implementations

```go
type MockLogger struct {
    logs []string
}

func (m *MockLogger) Log(message string) {
    m.logs = append(m.logs, message)
}

func (m *MockLogger) GetLogs() []string {
    return m.logs
}

type MockDatabase struct {
    users map[string]*User
}

func (m *MockDatabase) SaveUser(user *User) error {
    m.users[user.Email] = user
    return nil
}

func (m *MockDatabase) GetUsers() map[string]*User {
    return m.users
}
```

### Test Example

```go
func TestUserService_CreateUser(t *testing.T) {
    // Setup
    container := setupTestContainer()
    service := inject.MustResolve[*UserService](container)
    
    // Execute
    err := service.CreateUser("John", "john@example.com")
    
    // Verify
    assert.NoError(t, err)
    
    // Check mock interactions
    mockDB := inject.MustResolve[Database](container).(*MockDatabase)
    users := mockDB.GetUsers()
    assert.Len(t, users, 1)
    assert.Equal(t, "John", users["john@example.com"].Name)
    
    mockLogger := inject.MustResolve[Logger](container).(*MockLogger)
    logs := mockLogger.GetLogs()
    assert.Contains(t, logs[0], "Creating user: John")
}
```

## Advanced Topics

### Container Composition

```go
func createWebContainer() *inject.Container {
    container := inject.NewContainer()
    
    // Add infrastructure services
    addInfrastructureServices(container)
    
    // Add business services
    addBusinessServices(container)
    
    // Add web-specific services
    addWebServices(container)
    
    return container
}

func addInfrastructureServices(container *inject.Container) {
    inject.RegisterSingletonType[*DatabaseConnection](container, ...)
    inject.RegisterSingletonType[*CacheService](container, ...)
}
```

### Dynamic Service Registration

```go
func registerDynamicServices(container *inject.Container, config *Config) {
    // Register different implementations based on configuration
    if config.UseRedisCache {
        inject.RegisterSingletonInterface[CacheService, *RedisCache](container, ...)
    } else {
        inject.RegisterSingletonInterface[CacheService, *MemoryCache](container, ...)
    }
    
    // Register environment-specific services
    if config.Environment == "production" {
        inject.RegisterSingletonInterface[Logger, *FileLogger](container, ...)
    } else {
        inject.RegisterSingletonInterface[Logger, *ConsoleLogger](container, ...)
    }
}
```

### Service Decoration

```go
// Base implementation
type BasicEmailService struct{}

func (b *BasicEmailService) SendEmail(to, subject, body string) error {
    // Send email logic
    return nil
}

// Decorator that adds logging
type LoggingEmailDecorator struct {
    inner  EmailService
    logger Logger
}

func (l *LoggingEmailDecorator) SendEmail(to, subject, body string) error {
    l.logger.Log(fmt.Sprintf("Sending email to %s: %s", to, subject))
    err := l.inner.SendEmail(to, subject, body)
    if err != nil {
        l.logger.Log(fmt.Sprintf("Failed to send email: %v", err))
    } else {
        l.logger.Log("Email sent successfully")
    }
    return err
}

// Registration
inject.RegisterSingletonType[*BasicEmailService](container, ...)
inject.RegisterSingletonInterface[EmailService, *LoggingEmailDecorator](container,
    func(c *inject.Container) *LoggingEmailDecorator {
        basic := inject.MustResolve[*BasicEmailService](c)
        logger := inject.MustResolve[Logger](c)
        return &LoggingEmailDecorator{inner: basic, logger: logger}
    })
```

### Conditional Registration

```go
func registerConditionalServices(container *inject.Container) {
    // Register base services
    inject.RegisterSingletonType[*ConfigService](container, ...)
    
    // Conditionally register additional services
    container.RegisterFunc(func(config *ConfigService) *OptionalService {
        if config.FeatureEnabled("optional-feature") {
            return &OptionalService{enabled: true}
        }
        return nil
    }, inject.Singleton)
}
```

## Best Practices

### 1. Design with Interfaces

```go
// Good: Interface-first design
type UserRepository interface {
    GetUser(id string) (*User, error)
    SaveUser(user *User) error
}

type UserService struct {
    repo UserRepository // Interface dependency
}

// Avoid: Concrete dependencies
type UserService struct {
    repo *MySQLUserRepository // Concrete dependency
}
```

### 2. Use Appropriate Lifetimes

```go
// Singleton for stateless, expensive services
inject.RegisterSingletonInterface[Database, *PostgreSQLDatabase](container, ...)

// Transient for stateful, lightweight services
inject.RegisterTransientType[*RequestProcessor](container, ...)
```

### 3. Organize Registration

```go
type ContainerBuilder struct {
    container *inject.Container
}

func NewContainerBuilder() *ContainerBuilder {
    return &ContainerBuilder{
        container: inject.NewContainer(),
    }
}

func (b *ContainerBuilder) AddLogging() *ContainerBuilder {
    inject.RegisterSingletonInterface[Logger, *StructuredLogger](b.container, ...)
    return b
}

func (b *ContainerBuilder) AddDatabase() *ContainerBuilder {
    inject.RegisterSingletonInterface[Database, *PostgreSQLDatabase](b.container, ...)
    return b
}

func (b *ContainerBuilder) Build() *inject.Container {
    return b.container
}

// Usage
container := NewContainerBuilder().
    AddLogging().
    AddDatabase().
    Build()
```

### 4. Error Handling Strategy

```go
func mustRegister[T any](container *inject.Container, factory func(*inject.Container) T, lifecycle inject.Lifecycle) {
    err := inject.RegisterType[T](container, factory, lifecycle)
    if err != nil {
        panic(fmt.Sprintf("Failed to register service %T: %v", *new(T), err))
    }
}
```

### 5. Testing Strategy

```go
// Create test-specific registration functions
func registerMockServices(container *inject.Container) {
    inject.RegisterValue[Logger](container, &MockLogger{})
    inject.RegisterValue[Database](container, &MockDatabase{})
}

func registerProductionServices(container *inject.Container) {
    inject.RegisterSingletonInterface[Logger, *FileLogger](container, ...)
    inject.RegisterSingletonInterface[Database, *PostgreSQLDatabase](container, ...)
}
```

## Common Patterns

### 1. Repository Pattern

```go
type UserRepository interface {
    GetByID(id string) (*User, error)
    GetByEmail(email string) (*User, error)
    Save(user *User) error
    Delete(id string) error
}

type DatabaseUserRepository struct {
    db Database
}

type CachedUserRepository struct {
    inner UserRepository
    cache Cache
}

func (c *CachedUserRepository) GetByID(id string) (*User, error) {
    // Try cache first
    if user, found := c.cache.Get("user:" + id); found {
        return user.(*User), nil
    }
    
    // Fallback to inner repository
    user, err := c.inner.GetByID(id)
    if err == nil {
        c.cache.Set("user:"+id, user)
    }
    return user, err
}
```

### 2. Service Layer Pattern

```go
type UserService interface {
    CreateUser(req CreateUserRequest) (*User, error)
    UpdateUser(id string, req UpdateUserRequest) (*User, error)
    GetUser(id string) (*User, error)
}

type UserServiceImpl struct {
    repo      UserRepository
    validator UserValidator
    events    EventPublisher
    logger    Logger
}

func (s *UserServiceImpl) CreateUser(req CreateUserRequest) (*User, error) {
    // Validate
    if err := s.validator.ValidateCreateRequest(req); err != nil {
        return nil, err
    }
    
    // Business logic
    user := &User{
        ID:    generateID(),
        Name:  req.Name,
        Email: req.Email,
    }
    
    // Persist
    if err := s.repo.Save(user); err != nil {
        return nil, err
    }
    
    // Publish event
    s.events.Publish(UserCreatedEvent{UserID: user.ID})
    
    s.logger.Log(fmt.Sprintf("User created: %s", user.ID))
    return user, nil
}
```

### 3. Factory Pattern

```go
type ConnectionFactory interface {
    CreateConnection(config ConnectionConfig) (Connection, error)
}

type DatabaseConnectionFactory struct {
    logger Logger
}

func (f *DatabaseConnectionFactory) CreateConnection(config ConnectionConfig) (Connection, error) {
    f.logger.Log(fmt.Sprintf("Creating connection to %s", config.Host))
    
    switch config.Type {
    case "postgres":
        return &PostgreSQLConnection{config: config}, nil
    case "mysql":
        return &MySQLConnection{config: config}, nil
    default:
        return nil, fmt.Errorf("unsupported database type: %s", config.Type)
    }
}
```

### 4. Event-Driven Architecture

```go
type EventBus interface {
    Subscribe(eventType string, handler EventHandler)
    Publish(event Event)
}

type UserService struct {
    repo     UserRepository
    eventBus EventBus
}

func (s *UserService) CreateUser(req CreateUserRequest) (*User, error) {
    user := &User{...}
    
    if err := s.repo.Save(user); err != nil {
        return nil, err
    }
    
    // Publish event for other services to react
    s.eventBus.Publish(UserCreatedEvent{
        UserID: user.ID,
        Email:  user.Email,
    })
    
    return user, nil
}

// Other services can subscribe to events
type EmailService struct {
    smtp SMTPClient
}

func (e *EmailService) HandleUserCreated(event UserCreatedEvent) {
    e.smtp.SendWelcomeEmail(event.Email)
}
```

## Conclusion

Dependency Injection with Go-Inject provides a powerful foundation for building maintainable, testable, and flexible Go applications. By following the patterns and practices outlined in this tutorial, you can:

- Create loosely coupled, easily testable code
- Manage complex dependency graphs efficiently
- Build applications that are easy to configure and extend
- Implement clean architecture patterns

Remember to:
- Design with interfaces first
- Choose appropriate service lifetimes
- Organize your registration code
- Handle errors gracefully
- Write comprehensive tests with mocks

Start with simple examples and gradually incorporate more advanced patterns as your application grows in complexity.

### Next Steps

1. **Practice**: Start with the basic examples and build up to more complex scenarios
2. **Experiment**: Try different registration patterns and lifetimes
3. **Test**: Write comprehensive tests using the DI container
4. **Refactor**: Apply DI patterns to existing code gradually
5. **Learn**: Study the examples in the `/examples` directory

Happy coding with Go-Inject! 🚀