package inject

import (
	"fmt"
	"reflect"
	"sync"
)

type Lifecycle int

const (
	Transient Lifecycle = iota
	Singleton
)

type ServiceDescriptor struct {
	ServiceType reflect.Type
	Factory     interface{}
	Lifecycle   Lifecycle
	instance    interface{}
	mu          sync.RWMutex
}

type Container struct {
	services map[reflect.Type]*ServiceDescriptor
	mu       sync.RWMutex
}

func NewContainer() *Container {
	return &Container{
		services: make(map[reflect.Type]*ServiceDescriptor),
	}
}

func (c *Container) Register(serviceType interface{}, factory interface{}, lifecycle Lifecycle) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	sType := reflect.TypeOf(serviceType)
	if sType.Kind() == reflect.Ptr {
		sType = sType.Elem()
	}

	factoryType := reflect.TypeOf(factory)
	if factoryType.Kind() != reflect.Func {
		return fmt.Errorf("factory must be a function")
	}

	if factoryType.NumOut() != 1 && factoryType.NumOut() != 2 {
		return fmt.Errorf("factory function must return 1 or 2 values (service and optionally error)")
	}

	if factoryType.NumOut() == 2 {
		errorInterface := reflect.TypeOf((*error)(nil)).Elem()
		if !factoryType.Out(1).Implements(errorInterface) {
			return fmt.Errorf("second return value must be an error")
		}
	}

	returnType := factoryType.Out(0)

	// Check type compatibility
	if sType.Kind() == reflect.Interface {
		// Service type is an interface, check if return type implements it
		if !returnType.Implements(sType) {
			return fmt.Errorf("factory return type %s does not implement interface %s", returnType.String(), sType.String())
		}
	} else {
		// Service type is concrete, check for exact match or pointer compatibility
		if returnType != sType {
			if returnType.Kind() == reflect.Ptr && returnType.Elem() == sType {
				// Pointer to service type is acceptable
			} else if sType.Kind() == reflect.Ptr && sType.Elem() == returnType {
				// Service type is pointer, return type is value
			} else {
				return fmt.Errorf("factory return type %s does not match service type %s", returnType.String(), sType.String())
			}
		}
	}

	descriptor := &ServiceDescriptor{
		ServiceType: sType,
		Factory:     factory,
		Lifecycle:   lifecycle,
	}

	c.services[sType] = descriptor
	return nil
}

func (c *Container) RegisterSingleton(serviceType interface{}, factory interface{}) error {
	return c.Register(serviceType, factory, Singleton)
}

func (c *Container) RegisterTransient(serviceType interface{}, factory interface{}) error {
	return c.Register(serviceType, factory, Transient)
}

func (c *Container) Resolve(serviceType interface{}) (interface{}, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	sType := reflect.TypeOf(serviceType)
	if sType.Kind() == reflect.Ptr {
		sType = sType.Elem()
	}

	return c.resolveType(sType)
}

func (c *Container) resolveType(serviceType reflect.Type) (interface{}, error) {
	descriptor, exists := c.services[serviceType]
	if !exists {
		return nil, fmt.Errorf("service of type %s not registered", serviceType.String())
	}

	if descriptor.Lifecycle == Singleton {
		descriptor.mu.RLock()
		if descriptor.instance != nil {
			instance := descriptor.instance
			descriptor.mu.RUnlock()
			return instance, nil
		}
		descriptor.mu.RUnlock()

		descriptor.mu.Lock()
		defer descriptor.mu.Unlock()

		if descriptor.instance != nil {
			return descriptor.instance, nil
		}

		instance, err := c.createInstance(descriptor)
		if err != nil {
			return nil, err
		}
		descriptor.instance = instance
		return instance, nil
	}

	return c.createInstance(descriptor)
}

func (c *Container) createInstance(descriptor *ServiceDescriptor) (interface{}, error) {
	factoryValue := reflect.ValueOf(descriptor.Factory)
	factoryType := factoryValue.Type()

	args := make([]reflect.Value, factoryType.NumIn())
	for i := 0; i < factoryType.NumIn(); i++ {
		argType := factoryType.In(i)

		if argType == reflect.TypeOf((*Container)(nil)) {
			args[i] = reflect.ValueOf(c)
			continue
		}

		arg, err := c.resolveType(argType)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve dependency %s: %w", argType.String(), err)
		}
		args[i] = reflect.ValueOf(arg)
	}

	results := factoryValue.Call(args)

	if len(results) == 2 {
		if !results[1].IsNil() {
			return nil, results[1].Interface().(error)
		}
	}

	return results[0].Interface(), nil
}

func (c *Container) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.services = make(map[reflect.Type]*ServiceDescriptor)
}
