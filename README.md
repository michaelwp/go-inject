# Go-Inject 🚀

A powerful, thread-safe dependency injection container for Go applications. Go-Inject provides a clean and intuitive API for managing dependencies with support for interfaces, generics, and multiple service lifetimes.

## Features ✨

- **Type-safe dependency injection** with Go generics support
- **Interface-based registration** and resolution
- **Multiple service lifetimes**: Singleton and Transient
- **Thread-safe operations** with concurrent access support
- **Automatic dependency resolution** with circular dependency detection
- **Factory functions** with error handling
- **Container injection** for advanced scenarios
- **Clean, intuitive API** following Go best practices

## Installation 📦

```bash
go get github.com/go-inject/go-inject
```

## Quick Start 🚀

```go
package main

import (
    "fmt"
    "github.com/go-inject/go-inject"
)

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
    u.logger.Log(fmt.Sprintf("Creating user: %s", name))
}

func main() {
    container := inject.NewContainer()

    // Register interface implementation
    inject.RegisterSingletonInterface[Logger, *ConsoleLogger](container, func(c *inject.Container) *ConsoleLogger {
        return &ConsoleLogger{}
    })

    // Register service with dependency
    inject.RegisterTransientType[*UserService](container, func(c *inject.Container) *UserService {
        logger := inject.MustResolve[Logger](c)
        return &UserService{logger: logger}
    })

    // Resolve and use service
    userService := inject.MustResolve[*UserService](container)
    userService.CreateUser("John Doe")
}
```

## Core Concepts 📚

### Container

The `Container` is the central component that manages service registration and resolution:

```go
container := inject.NewContainer()
```

### Service Lifetimes

- **Singleton**: One instance per container, created on first request
- **Transient**: New instance on every request

### Registration Methods

#### Basic Registration

```go
// Register with explicit lifecycle
container.Register((*MyService)(nil), func() *MyService {
    return &MyService{}
}, inject.Singleton)

// Convenience methods
container.RegisterSingleton((*MyService)(nil), func() *MyService {
    return &MyService{}
})

container.RegisterTransient((*MyService)(nil), func() *MyService {
    return &MyService{}
})
```

#### Generic Helpers

```go
// Type registration
inject.RegisterSingletonType[*MyService](container, func(c *inject.Container) *MyService {
    return &MyService{}
})

inject.RegisterTransientType[*MyService](container, func(c *inject.Container) *MyService {
    return &MyService{}
})

// Interface registration
inject.RegisterSingletonInterface[MyInterface, *MyImplementation](container, 
    func(c *inject.Container) *MyImplementation {
        return &MyImplementation{}
    })

// Value registration (always singleton)
myInstance := &MyService{}
inject.RegisterValue[*MyService](container, myInstance)
```

### Resolution Methods

```go
// Standard resolution with error handling
service, err := container.Resolve((*MyService)(nil))
if err != nil {
    // handle error
}

// Generic helper that panics on error
service := inject.MustResolve[*MyService](container)

// Safe resolution that returns bool
service, ok := inject.TryResolve[*MyService](container)
if !ok {
    // service not found
}
```

## Advanced Usage 🔧

### Dependency Injection

Services can automatically receive their dependencies:

```go
type Database interface {
    Query(sql string) []string
}

type UserRepository struct {
    db Database
}

type UserService struct {
    repo *UserRepository
    logger Logger
}

// Register dependencies
inject.RegisterSingletonInterface[Database, *MySQLDatabase](container, 
    func(c *inject.Container) *MySQLDatabase {
        return &MySQLDatabase{connectionString: "..."}
    })

inject.RegisterTransientType[*UserRepository](container, 
    func(c *inject.Container) *UserRepository {
        db := inject.MustResolve[Database](c)
        return &UserRepository{db: db}
    })

inject.RegisterTransientType[*UserService](container, 
    func(c *inject.Container) *UserService {
        repo := inject.MustResolve[*UserRepository](c)
        logger := inject.MustResolve[Logger](c)
        return &UserService{repo: repo, logger: logger}
    })
```

### Factory Functions with Error Handling

```go
container.RegisterSingleton((*DatabaseConnection)(nil), func() (*DatabaseConnection, error) {
    conn, err := sql.Open("mysql", "connection-string")
    if err != nil {
        return nil, fmt.Errorf("failed to connect to database: %w", err)
    }
    return &DatabaseConnection{conn: conn}, nil
})
```

### Container Injection

Access the container within factory functions:

```go
inject.RegisterSingletonType[*ComplexService](container, 
    func(c *inject.Container) *ComplexService {
        // Access other services through the container
        if c.Has((*OptionalService)(nil)) {
            optional := inject.MustResolve[*OptionalService](c)
            return &ComplexService{optional: optional}
        }
        return &ComplexService{}
    })
```

### Utility Methods

```go
// Check if service is registered
if container.Has((*MyService)(nil)) {
    // service is registered
}

// Get all registered service types
types := container.GetServiceTypes()
for _, serviceType := range types {
    fmt.Println("Registered:", serviceType)
}

// Clear all registrations
container.Clear()
```

## Best Practices 💡

### 1. Use Interfaces

Define and register interfaces rather than concrete types:

```go
type Logger interface {
    Log(message string)
}

// Good: Register interface
inject.RegisterSingletonInterface[Logger, *ConsoleLogger](container, ...)

// Avoid: Register concrete type directly
inject.RegisterSingletonType[*ConsoleLogger](container, ...)
```

### 2. Choose Appropriate Lifetimes

- Use **Singleton** for stateless services, configurations, and expensive resources
- Use **Transient** for stateful services and lightweight objects

```go
// Singleton - shared state, expensive to create
inject.RegisterSingletonInterface[Database, *MySQLDatabase](container, ...)

// Transient - request-scoped, stateful
inject.RegisterTransientType[*OrderProcessor](container, ...)
```

### 3. Handle Errors Gracefully

```go
// In production code, handle errors
service, err := container.Resolve((*CriticalService)(nil))
if err != nil {
    log.Fatalf("Failed to resolve critical service: %v", err)
}

// Use MustResolve only when you're certain the service exists
service := inject.MustResolve[*KnownService](container)
```

### 4. Organize Registration

Create a setup function to organize your registrations:

```go
func SetupContainer() *inject.Container {
    container := inject.NewContainer()
    
    // Infrastructure
    registerDatabase(container)
    registerLogging(container)
    
    // Repositories
    registerRepositories(container)
    
    // Services
    registerServices(container)
    
    return container
}

func registerDatabase(container *inject.Container) {
    inject.RegisterSingletonInterface[Database, *MySQLDatabase](container, ...)
}
```

## Error Handling 🚨

The library provides detailed error messages for common issues:

- **Service not registered**: Clear message indicating which service type is missing
- **Factory function errors**: Propagated from factory functions that return errors
- **Type mismatches**: Validation during registration prevents runtime errors
- **Circular dependencies**: Detected and reported with dependency chain

## Thread Safety 🔒

All container operations are thread-safe:

- Multiple goroutines can safely register services
- Concurrent resolution is supported
- Singleton instances are created safely with double-checked locking

## Performance Considerations ⚡

- **Service resolution**: O(1) lookup time
- **Singleton creation**: One-time cost with lazy initialization
- **Memory usage**: Minimal overhead, only stores service descriptors
- **Concurrent access**: Optimized read-write locks for high concurrency

## Testing 🧪

The library includes comprehensive test coverage. Run tests with:

```bash
go test ./...
```

### Testing with Dependency Injection

```go
func TestUserService(t *testing.T) {
    container := inject.NewContainer()
    
    // Register mock dependencies
    inject.RegisterSingletonInterface[Logger, *MockLogger](container, 
        func(c *inject.Container) *MockLogger {
            return &MockLogger{}
        })
    
    inject.RegisterTransientType[*UserService](container, 
        func(c *inject.Container) *UserService {
            logger := inject.MustResolve[Logger](c)
            return &UserService{logger: logger}
        })
    
    // Test the service
    userService := inject.MustResolve[*UserService](container)
    userService.CreateUser("Test User")
    
    // Assert mock expectations...
}
```

## Contributing 🤝

Contributions are welcome! Please read our contributing guidelines and submit pull requests to the main repository.

## License 📄

This project is licensed under the MIT License - see the LICENSE file for details.

## Changelog 📝

### v1.0.0
- Initial release
- Core dependency injection functionality
- Generic helper methods
- Comprehensive test suite
- Thread-safe operations