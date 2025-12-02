package providers

import (
	"fmt"
	"sort"
	"sync"
)

// Registry manages available providers for databases, caches, compute, and CDN.
// It provides thread-safe registration and retrieval of providers.
type Registry struct {
	databases map[string]DatabaseProvider
	caches    map[string]CacheProvider
	computes  map[string]ComputeProvider
	cdns      map[string]CDNProvider
	mu        sync.RWMutex
}

// NewRegistry creates a new provider registry
func NewRegistry() *Registry {
	return &Registry{
		databases: make(map[string]DatabaseProvider),
		caches:    make(map[string]CacheProvider),
		computes:  make(map[string]ComputeProvider),
		cdns:      make(map[string]CDNProvider),
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

// RegisterCompute registers a compute provider
func (r *Registry) RegisterCompute(name string, provider ComputeProvider) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.computes[name] = provider
}

// GetCompute retrieves a compute provider by name
func (r *Registry) GetCompute(name string) (ComputeProvider, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	provider, ok := r.computes[name]
	if !ok {
		return nil, fmt.Errorf("compute provider %q not found (available: %v)", name, r.listComputesUnsafe())
	}
	return provider, nil
}

// ListComputes returns all registered compute provider names
func (r *Registry) ListComputes() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.listComputesUnsafe()
}

func (r *Registry) listComputesUnsafe() []string {
	names := make([]string, 0, len(r.computes))
	for name := range r.computes {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// RegisterCDN registers a CDN provider
func (r *Registry) RegisterCDN(name string, provider CDNProvider) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.cdns[name] = provider
}

// GetCDN retrieves a CDN provider by name
func (r *Registry) GetCDN(name string) (CDNProvider, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	provider, ok := r.cdns[name]
	if !ok {
		return nil, fmt.Errorf("CDN provider %q not found (available: %v)", name, r.listCDNsUnsafe())
	}
	return provider, nil
}

// ListCDNs returns all registered CDN provider names
func (r *Registry) ListCDNs() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.listCDNsUnsafe()
}

func (r *Registry) listCDNsUnsafe() []string {
	names := make([]string, 0, len(r.cdns))
	for name := range r.cdns {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// DefaultRegistry is the global provider registry
var DefaultRegistry = NewRegistry()
