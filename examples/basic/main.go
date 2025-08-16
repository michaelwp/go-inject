package main

import (
	"fmt"

	"github.com/go-inject/go-inject"
)

type Logger interface {
	Log(message string)
}

type ConsoleLogger struct {
	prefix string
}

func (c *ConsoleLogger) Log(message string) {
	fmt.Printf("[%s] %s\n", c.prefix, message)
}

type UserService struct {
	logger Logger
}

func (u *UserService) CreateUser(name string) {
	u.logger.Log(fmt.Sprintf("Creating user: %s", name))
}

func (u *UserService) DeleteUser(name string) {
	u.logger.Log(fmt.Sprintf("Deleting user: %s", name))
}

func main() {
	fmt.Println("=== Basic Dependency Injection Example ===")

	container := inject.NewContainer()

	// Register logger as singleton
	err := inject.RegisterSingletonInterface[Logger, *ConsoleLogger](container, func(c *inject.Container) *ConsoleLogger {
		return &ConsoleLogger{prefix: "APP"}
	})
	if err != nil {
		return
	}

	// Register user service as transient
	err = inject.RegisterTransientType[*UserService](container, func(c *inject.Container) *UserService {
		logger := inject.MustResolve[Logger](c)
		return &UserService{logger: logger}
	})
	if err != nil {
		return
	}

	// Resolve and use services
	userService1 := inject.MustResolve[*UserService](container)
	userService2 := inject.MustResolve[*UserService](container)

	userService1.CreateUser("Alice")
	userService2.CreateUser("Bob")
	userService1.DeleteUser("Alice")

	// Demonstrate that services are different instances (transient)
	// but share the same logger instance (singleton)
	fmt.Printf("UserService1 address: %p\n", userService1)
	fmt.Printf("UserService2 address: %p\n", userService2)
	fmt.Printf("Same UserService instances? %t\n", userService1 == userService2)
}
