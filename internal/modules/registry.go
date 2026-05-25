package modules

import (
	"context"
	"fmt"

	"erp-system/pkg/logger"
)

// Registry holds all registered modules and manages their lifecycle.
type Registry struct {
	modules map[string]Module
	order   []string // topologically sorted load order
}

func NewRegistry() *Registry {
	return &Registry{modules: make(map[string]Module)}
}

// Register adds a module to the registry. Call before Initialize.
func (r *Registry) Register(m Module) {
	r.modules[m.Name()] = m
}

// Initialize resolves dependency order and calls Initialize on each module.
func (r *Registry) Initialize(deps Dependencies) error {
	order, err := topoSort(r.modules)
	if err != nil {
		return fmt.Errorf("module registry: %w", err)
	}
	r.order = order

	for _, name := range order {
		m := r.modules[name]
		logger.Info("modules: initializing", logger.String("module", name))
		if err := m.Initialize(deps); err != nil {
			return fmt.Errorf("module %s: initialize: %w", name, err)
		}
	}
	return nil
}

// RegisterEvents calls RegisterEvents on all modules in load order.
func (r *Registry) RegisterEvents(bus interface{ Subscribe(stream, group, consumer string, handler interface{}) }) {
	// Modules call bus.Subscribe themselves; this method is a hook for future use.
}

// Shutdown calls Shutdown on all modules in reverse load order.
func (r *Registry) Shutdown(ctx context.Context) {
	for i := len(r.order) - 1; i >= 0; i-- {
		name := r.order[i]
		if err := r.modules[name].Shutdown(ctx); err != nil {
			logger.Error("modules: shutdown error",
				logger.String("module", name),
				logger.Err(err),
			)
		}
	}
}

// Status returns a map of module name → loaded status.
func (r *Registry) Status() map[string]string {
	status := make(map[string]string, len(r.modules))
	loaded := make(map[string]bool, len(r.order))
	for _, n := range r.order {
		loaded[n] = true
	}
	for name := range r.modules {
		if loaded[name] {
			status[name] = "loaded"
		} else {
			status[name] = "registered"
		}
	}
	return status
}

// topoSort returns modules in dependency-first order.
func topoSort(modules map[string]Module) ([]string, error) {
	visited := make(map[string]bool)
	inStack := make(map[string]bool)
	var order []string

	var visit func(name string) error
	visit = func(name string) error {
		if inStack[name] {
			return fmt.Errorf("circular dependency detected at module %q", name)
		}
		if visited[name] {
			return nil
		}
		inStack[name] = true
		m, ok := modules[name]
		if !ok {
			return fmt.Errorf("module %q not registered", name)
		}
		for _, dep := range m.Dependencies() {
			if err := visit(dep); err != nil {
				return err
			}
		}
		inStack[name] = false
		visited[name] = true
		order = append(order, name)
		return nil
	}

	for name := range modules {
		if err := visit(name); err != nil {
			return nil, err
		}
	}
	return order, nil
}
