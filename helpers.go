package inject

import (
	"fmt"
	"reflect"
)

func MustResolve[T any](container *Container) T {
	var zero T
	result, err := container.Resolve((*T)(nil))
	if err != nil {
		panic(fmt.Sprintf("failed to resolve service of type %T: %v", zero, err))
	}
	return result.(T)
}

func TryResolve[T any](container *Container) (T, bool) {
	var zero T
	result, err := container.Resolve((*T)(nil))
	if err != nil {
		return zero, false
	}
	return result.(T), true
}

func RegisterInterface[TInterface, TImplementation any](container *Container, factory func(*Container) TImplementation, lifecycle Lifecycle) error {
	return container.Register((*TInterface)(nil), func(c *Container) TInterface {
		impl := factory(c)
		return any(impl).(TInterface)
	}, lifecycle)
}

func RegisterSingletonInterface[TInterface, TImplementation any](container *Container, factory func(*Container) TImplementation) error {
	return RegisterInterface[TInterface, TImplementation](container, factory, Singleton)
}

func RegisterTransientInterface[TInterface, TImplementation any](container *Container, factory func(*Container) TImplementation) error {
	return RegisterInterface[TInterface, TImplementation](container, factory, Transient)
}

func RegisterType[T any](container *Container, factory func(*Container) T, lifecycle Lifecycle) error {
	return container.Register((*T)(nil), factory, lifecycle)
}

func RegisterSingletonType[T any](container *Container, factory func(*Container) T) error {
	return RegisterType[T](container, factory, Singleton)
}

func RegisterTransientType[T any](container *Container, factory func(*Container) T) error {
	return RegisterType[T](container, factory, Transient)
}

func RegisterValue[T any](container *Container, value T) error {
	return container.RegisterSingleton((*T)(nil), func() T {
		return value
	})
}

func (c *Container) RegisterFunc(factory interface{}, lifecycle Lifecycle) error {
	factoryType := reflect.TypeOf(factory)
	if factoryType.Kind() != reflect.Func {
		return fmt.Errorf("factory must be a function")
	}

	if factoryType.NumOut() == 0 {
		return fmt.Errorf("factory function must return at least one value")
	}

	returnType := factoryType.Out(0)

	if returnType.Kind() == reflect.Interface {
		return c.Register(reflect.New(returnType).Interface(), factory, lifecycle)
	}

	// Register with the exact return type
	return c.Register(reflect.New(returnType).Interface(), factory, lifecycle)
}

func (c *Container) Has(serviceType interface{}) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	sType := reflect.TypeOf(serviceType)
	if sType.Kind() == reflect.Ptr {
		sType = sType.Elem()
	}

	_, exists := c.services[sType]
	return exists
}

func (c *Container) GetServiceTypes() []reflect.Type {
	c.mu.RLock()
	defer c.mu.RUnlock()

	types := make([]reflect.Type, 0, len(c.services))
	for serviceType := range c.services {
		types = append(types, serviceType)
	}
	return types
}
