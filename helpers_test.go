package inject

import (
	"testing"
)

func TestMustResolve(t *testing.T) {
	container := NewContainer()

	err := RegisterTransientType[*TestImplementation](container, func(c *Container) *TestImplementation {
		return &TestImplementation{value: "must resolve"}
	})
	if err != nil {
		t.Fatalf("Failed to register service: %v", err)
	}

	service := MustResolve[*TestImplementation](container)
	if service.GetValue() != "must resolve" {
		t.Error("MustResolve should return correct service")
	}
}

func TestMustResolvePanic(t *testing.T) {
	container := NewContainer()

	defer func() {
		if r := recover(); r == nil {
			t.Error("MustResolve should panic when service is not registered")
		}
	}()

	MustResolve[*TestImplementation](container)
}

func TestTryResolve(t *testing.T) {
	container := NewContainer()

	service, ok := TryResolve[*TestImplementation](container)
	if ok {
		t.Error("TryResolve should return false for unregistered service")
	}
	if service != nil {
		t.Error("TryResolve should return nil for unregistered service")
	}

	err := RegisterTransientType[*TestImplementation](container, func(c *Container) *TestImplementation {
		return &TestImplementation{value: "try resolve"}
	})
	if err != nil {
		t.Fatalf("Failed to register service: %v", err)
	}

	service, ok = TryResolve[*TestImplementation](container)
	if !ok {
		t.Error("TryResolve should return true for registered service")
	}
	if service.GetValue() != "try resolve" {
		t.Error("TryResolve should return correct service")
	}
}

func TestRegisterInterface(t *testing.T) {
	container := NewContainer()

	err := RegisterInterface[TestInterface, *TestImplementation](container, func(c *Container) *TestImplementation {
		return &TestImplementation{value: "interface helper"}
	}, Transient)
	if err != nil {
		t.Fatalf("Failed to register interface: %v", err)
	}

	service, err := container.Resolve((*TestInterface)(nil))
	if err != nil {
		t.Fatalf("Failed to resolve interface: %v", err)
	}

	impl := service.(TestInterface)
	if impl.GetValue() != "interface helper" {
		t.Error("Interface should be resolved correctly")
	}
}

func TestRegisterSingletonInterface(t *testing.T) {
	container := NewContainer()

	err := RegisterSingletonInterface[TestInterface, *TestImplementation](container, func(c *Container) *TestImplementation {
		return &TestImplementation{value: "singleton interface"}
	})
	if err != nil {
		t.Fatalf("Failed to register singleton interface: %v", err)
	}

	service1, err := container.Resolve((*TestInterface)(nil))
	if err != nil {
		t.Fatalf("Failed to resolve interface: %v", err)
	}

	service2, err := container.Resolve((*TestInterface)(nil))
	if err != nil {
		t.Fatalf("Failed to resolve interface: %v", err)
	}

	if service1 != service2 {
		t.Error("Singleton interface should return same instance")
	}
}

func TestRegisterTransientInterface(t *testing.T) {
	container := NewContainer()

	err := RegisterTransientInterface[TestInterface, *TestImplementation](container, func(c *Container) *TestImplementation {
		return &TestImplementation{value: "transient interface"}
	})
	if err != nil {
		t.Fatalf("Failed to register transient interface: %v", err)
	}

	service1, err := container.Resolve((*TestInterface)(nil))
	if err != nil {
		t.Fatalf("Failed to resolve interface: %v", err)
	}

	service2, err := container.Resolve((*TestInterface)(nil))
	if err != nil {
		t.Fatalf("Failed to resolve interface: %v", err)
	}

	if service1 == service2 {
		t.Error("Transient interface should return different instances")
	}
}

func TestRegisterType(t *testing.T) {
	container := NewContainer()

	err := RegisterType[*TestImplementation](container, func(c *Container) *TestImplementation {
		return &TestImplementation{value: "type helper"}
	}, Transient)
	if err != nil {
		t.Fatalf("Failed to register type: %v", err)
	}

	service := MustResolve[*TestImplementation](container)
	if service.GetValue() != "type helper" {
		t.Error("Type should be resolved correctly")
	}
}

func TestRegisterSingletonType(t *testing.T) {
	container := NewContainer()

	err := RegisterSingletonType[*TestImplementation](container, func(c *Container) *TestImplementation {
		return &TestImplementation{value: "singleton type"}
	})
	if err != nil {
		t.Fatalf("Failed to register singleton type: %v", err)
	}

	service1 := MustResolve[*TestImplementation](container)
	service2 := MustResolve[*TestImplementation](container)

	if service1 != service2 {
		t.Error("Singleton type should return same instance")
	}
}

func TestRegisterTransientType(t *testing.T) {
	container := NewContainer()

	err := RegisterTransientType[*TestImplementation](container, func(c *Container) *TestImplementation {
		return &TestImplementation{value: "transient type"}
	})
	if err != nil {
		t.Fatalf("Failed to register transient type: %v", err)
	}

	service1 := MustResolve[*TestImplementation](container)
	service2 := MustResolve[*TestImplementation](container)

	if service1 == service2 {
		t.Error("Transient type should return different instances")
	}
}

func TestRegisterValue(t *testing.T) {
	container := NewContainer()

	testValue := &TestImplementation{value: "registered value"}
	err := RegisterValue[*TestImplementation](container, testValue)
	if err != nil {
		t.Fatalf("Failed to register value: %v", err)
	}

	service1 := MustResolve[*TestImplementation](container)
	service2 := MustResolve[*TestImplementation](container)

	if service1 != testValue {
		t.Error("Registered value should be returned")
	}
	if service1 != service2 {
		t.Error("Registered value should return same instance")
	}
}

func TestRegisterFunc(t *testing.T) {
	container := NewContainer()

	err := container.RegisterFunc(func() *TestImplementation {
		return &TestImplementation{value: "register func"}
	}, Transient)
	if err != nil {
		t.Fatalf("Failed to register func: %v", err)
	}

	service := MustResolve[*TestImplementation](container)
	if service.GetValue() != "register func" {
		t.Error("RegisterFunc should work correctly")
	}
}

func TestRegisterFuncWithInterface(t *testing.T) {
	container := NewContainer()

	err := container.RegisterFunc(func() TestInterface {
		return &TestImplementation{value: "register func interface"}
	}, Singleton)
	if err != nil {
		t.Fatalf("Failed to register func with interface: %v", err)
	}

	service, err := container.Resolve((*TestInterface)(nil))
	if err != nil {
		t.Fatalf("Failed to resolve interface: %v", err)
	}

	impl := service.(TestInterface)
	if impl.GetValue() != "register func interface" {
		t.Error("RegisterFunc with interface should work correctly")
	}
}
