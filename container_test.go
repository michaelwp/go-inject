package inject

import (
	"errors"
	"testing"
)

type TestInterface interface {
	GetValue() string
}

type TestImplementation struct {
	value string
}

func (t *TestImplementation) GetValue() string {
	return t.value
}

type TestService struct {
	dependency TestInterface
}

func (t *TestService) GetDependency() TestInterface {
	return t.dependency
}

type TestRepository struct {
	data map[string]string
}

func (r *TestRepository) Get(key string) string {
	return r.data[key]
}

func (r *TestRepository) Set(key, value string) {
	r.data[key] = value
}

func TestNewContainer(t *testing.T) {
	container := NewContainer()
	if container == nil {
		t.Fatal("NewContainer should not return nil")
	}
	if container.services == nil {
		t.Fatal("Container services map should be initialized")
	}
}

func TestRegisterAndResolveTransient(t *testing.T) {
	container := NewContainer()

	err := container.RegisterTransient((*TestImplementation)(nil), func() *TestImplementation {
		return &TestImplementation{value: "test"}
	})
	if err != nil {
		t.Fatalf("Failed to register service: %v", err)
	}

	service1, err := container.Resolve((*TestImplementation)(nil))
	if err != nil {
		t.Fatalf("Failed to resolve service: %v", err)
	}

	service2, err := container.Resolve((*TestImplementation)(nil))
	if err != nil {
		t.Fatalf("Failed to resolve service: %v", err)
	}

	impl1 := service1.(*TestImplementation)
	impl2 := service2.(*TestImplementation)

	if impl1 == impl2 {
		t.Error("Transient services should return different instances")
	}

	if impl1.GetValue() != "test" || impl2.GetValue() != "test" {
		t.Error("Service instances should have correct values")
	}
}

func TestRegisterAndResolveSingleton(t *testing.T) {
	container := NewContainer()

	err := container.RegisterSingleton((*TestImplementation)(nil), func() *TestImplementation {
		return &TestImplementation{value: "singleton"}
	})
	if err != nil {
		t.Fatalf("Failed to register singleton service: %v", err)
	}

	service1, err := container.Resolve((*TestImplementation)(nil))
	if err != nil {
		t.Fatalf("Failed to resolve service: %v", err)
	}

	service2, err := container.Resolve((*TestImplementation)(nil))
	if err != nil {
		t.Fatalf("Failed to resolve service: %v", err)
	}

	impl1 := service1.(*TestImplementation)
	impl2 := service2.(*TestImplementation)

	if impl1 != impl2 {
		t.Error("Singleton services should return the same instance")
	}

	if impl1.GetValue() != "singleton" {
		t.Error("Singleton instance should have correct value")
	}
}

func TestInterfaceRegistration(t *testing.T) {
	container := NewContainer()

	err := container.Register((*TestInterface)(nil), func() TestInterface {
		return &TestImplementation{value: "interface"}
	}, Transient)
	if err != nil {
		t.Fatalf("Failed to register interface: %v", err)
	}

	service, err := container.Resolve((*TestInterface)(nil))
	if err != nil {
		t.Fatalf("Failed to resolve interface: %v", err)
	}

	impl := service.(TestInterface)
	if impl.GetValue() != "interface" {
		t.Error("Interface implementation should have correct value")
	}
}

func TestDependencyInjection(t *testing.T) {
	container := NewContainer()

	err := container.RegisterSingleton((*TestImplementation)(nil), func() *TestImplementation {
		return &TestImplementation{value: "dependency"}
	})
	if err != nil {
		t.Fatalf("Failed to register dependency: %v", err)
	}

	err = container.Register((*TestInterface)(nil), func() TestInterface {
		return &TestImplementation{value: "interface"}
	}, Transient)
	if err != nil {
		t.Fatalf("Failed to register interface: %v", err)
	}

	err = container.RegisterTransient((*TestService)(nil), func(dep TestInterface) *TestService {
		return &TestService{dependency: dep}
	})
	if err != nil {
		t.Fatalf("Failed to register service with dependency: %v", err)
	}

	service, err := container.Resolve((*TestService)(nil))
	if err != nil {
		t.Fatalf("Failed to resolve service with dependency: %v", err)
	}

	testService := service.(*TestService)
	if testService.GetDependency().GetValue() != "interface" {
		t.Error("Injected dependency should have correct value")
	}
}

func TestFactoryWithError(t *testing.T) {
	container := NewContainer()

	err := container.RegisterTransient((*TestImplementation)(nil), func() (*TestImplementation, error) {
		return nil, errors.New("factory error")
	})
	if err != nil {
		t.Fatalf("Failed to register service with error: %v", err)
	}

	_, err = container.Resolve((*TestImplementation)(nil))
	if err == nil {
		t.Error("Expected error from factory function")
	}
	if err.Error() != "factory error" {
		t.Errorf("Expected 'factory error', got '%s'", err.Error())
	}
}

func TestUnregisteredService(t *testing.T) {
	container := NewContainer()

	_, err := container.Resolve((*TestImplementation)(nil))
	if err == nil {
		t.Error("Expected error when resolving unregistered service")
	}
}

func TestInvalidFactory(t *testing.T) {
	container := NewContainer()

	err := container.Register((*TestImplementation)(nil), "not a function", Transient)
	if err == nil {
		t.Error("Expected error when registering non-function factory")
	}
}

func TestFactoryWithWrongReturnType(t *testing.T) {
	container := NewContainer()

	err := container.Register((*TestImplementation)(nil), func() string {
		return "wrong type"
	}, Transient)
	if err == nil {
		t.Error("Expected error when factory returns wrong type")
	}
}

func TestContainerInjection(t *testing.T) {
	container := NewContainer()

	err := container.RegisterSingleton((*TestRepository)(nil), func(c *Container) *TestRepository {
		repo := &TestRepository{data: make(map[string]string)}
		repo.Set("container", "injected")
		return repo
	})
	if err != nil {
		t.Fatalf("Failed to register service with container injection: %v", err)
	}

	service, err := container.Resolve((*TestRepository)(nil))
	if err != nil {
		t.Fatalf("Failed to resolve service: %v", err)
	}

	repo := service.(*TestRepository)
	if repo.Get("container") != "injected" {
		t.Error("Container should be injected into factory function")
	}
}

func TestClear(t *testing.T) {
	container := NewContainer()

	err := container.RegisterTransient((*TestImplementation)(nil), func() *TestImplementation {
		return &TestImplementation{value: "test"}
	})
	if err != nil {
		t.Fatalf("Failed to register service: %v", err)
	}

	if !container.Has((*TestImplementation)(nil)) {
		t.Error("Service should be registered")
	}

	container.Clear()

	if container.Has((*TestImplementation)(nil)) {
		t.Error("Service should be cleared")
	}
}

func TestHas(t *testing.T) {
	container := NewContainer()

	if container.Has((*TestImplementation)(nil)) {
		t.Error("Container should not have unregistered service")
	}

	err := container.RegisterTransient((*TestImplementation)(nil), func() *TestImplementation {
		return &TestImplementation{value: "test"}
	})
	if err != nil {
		t.Fatalf("Failed to register service: %v", err)
	}

	if !container.Has((*TestImplementation)(nil)) {
		t.Error("Container should have registered service")
	}
}

func TestGetServiceTypes(t *testing.T) {
	container := NewContainer()

	types := container.GetServiceTypes()
	if len(types) != 0 {
		t.Error("Empty container should have no service types")
	}

	err := container.RegisterTransient((*TestImplementation)(nil), func() *TestImplementation {
		return &TestImplementation{value: "test"}
	})
	if err != nil {
		t.Fatalf("Failed to register service: %v", err)
	}

	err = container.Register((*TestInterface)(nil), func() TestInterface {
		return &TestImplementation{value: "interface"}
	}, Singleton)
	if err != nil {
		t.Fatalf("Failed to register interface: %v", err)
	}

	types = container.GetServiceTypes()
	if len(types) != 2 {
		t.Errorf("Expected 2 service types, got %d", len(types))
	}
}
