package providers

import (
	"fmt"
	"sort"
	"sync"
)

// Registry manages available providers for databases and caches.
// It provides thread-safe registration and retrieval of providers.
type Registry struct {
	databases map[string]DatabaseProvider
	caches    map[string]CacheProvider
	mu        sync.RWMutex
}

// NewRegistry creates a new provider registry
func NewRegistry() *Registry {
	return &Registry{
		databases: make(map[string]DatabaseProvider),
		caches:    make(map[string]CacheProvider),
	}
}

// RegisterDatabase registers a database provider
func (r *Registry) RegisterDatabase(name string, provider DatabaseProvider) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.databases[name] = provider
}

// GetDatabase retrieves a database provider by name
func (r *Registry) GetDatabase(name string) (DatabaseProvider, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	provider, ok := r.databases[name]
	if !ok {
		return nil, fmt.Errorf("database provider %q not found (available: %v)", name, r.listDatabasesUnsafe())
	}
	return provider, nil
}

// ListDatabases returns all registered database provider names
func (r *Registry) ListDatabases() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.listDatabasesUnsafe()
}

func (r *Registry) listDatabasesUnsafe() []string {
	names := make([]string, 0, len(r.databases))
	for name := range r.databases {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// RegisterCache registers a cache provider
func (r *Registry) RegisterCache(name string, provider CacheProvider) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.caches[name] = provider
}

// GetCache retrieves a cache provider by name
func (r *Registry) GetCache(name string) (CacheProvider, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	provider, ok := r.caches[name]
	if !ok {
		return nil, fmt.Errorf("cache provider %q not found (available: %v)", name, r.listCachesUnsafe())
	}
	return provider, nil
}

// ListCaches returns all registered cache provider names
func (r *Registry) ListCaches() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.listCachesUnsafe()
}

func (r *Registry) listCachesUnsafe() []string {
	names := make([]string, 0, len(r.caches))
	for name := range r.caches {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// DefaultRegistry is the global provider registry
var DefaultRegistry = NewRegistry()
