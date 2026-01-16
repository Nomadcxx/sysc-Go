package floydtools

import (
	"fmt"
	"sync"
)

// Registry manages dynamic tool registration and dispatch.
type Registry struct {
	mu    sync.RWMutex
	tools map[string]Runner
}

// global registry instance
var global = &Registry{
	tools: make(map[string]Runner),
}

// Register adds a tool to the global registry.
// This function is designed to be called from init() functions.
func Register(name string, runner Runner) {
	global.mu.Lock()
	defer global.mu.Unlock()
	global.tools[name] = runner
}

// Get retrieves a tool by name, returning an error if not found.
func Get(name string) (Tool, error) {
	global.mu.RLock()
	runner, ok := global.tools[name]
	global.mu.RUnlock()

	if !ok {
		return nil, &NotFoundError{Name: name}
	}

	return runner(), nil
}

// List returns all registered tool names.
func List() []string {
	global.mu.RLock()
	defer global.mu.RUnlock()

	names := make([]string, 0, len(global.tools))
	for name := range global.tools {
		names = append(names, name)
	}
	return names
}

// MustGet retrieves a tool by name, panicking if not found.
// Useful for initialization code where missing tools are critical errors.
func MustGet(name string) Tool {
	tool, err := Get(name)
	if err != nil {
		panic(fmt.Sprintf("tool not found: %s", name))
	}
	return tool
}

// NotFoundError is returned when a tool is not registered.
type NotFoundError struct {
	Name string
}

func (e *NotFoundError) Error() string {
	return fmt.Sprintf("tool not found: %s", e.Name)
}

// IsNotFound checks if an error is a NotFoundError.
func IsNotFound(err error) bool {
	_, ok := err.(*NotFoundError)
	return ok
}

// AllTools returns a slice of all registered tool instances.
func AllTools() []Tool {
	global.mu.RLock()
	defer global.mu.RUnlock()

	tools := make([]Tool, 0, len(global.tools))
	for _, runner := range global.tools {
		tools = append(tools, runner())
	}
	return tools
}
