package chain

import (
	"fmt"
	"sync"
)

// Router maps network names to chain adapters.
type Router struct {
	mu       sync.RWMutex
	adapters map[string]Adapter
}

// NewRouter constructs an empty chain router.
func NewRouter() *Router {
	return &Router{
		adapters: make(map[string]Adapter),
	}
}

// Register registers an adapter under its network name.
func (r *Router) Register(adapter Adapter) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.adapters[adapter.Name()] = adapter
}

// Get resolves the adapter for a given network name.
func (r *Router) Get(network string) (Adapter, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	adapter, ok := r.adapters[network]
	if !ok {
		return nil, fmt.Errorf("%w: %q", ErrNetworkUnregistered, network)
	}
	return adapter, nil
}

// SupportedNetworks lists all registered network identifiers.
func (r *Router) SupportedNetworks() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	networks := make([]string, 0, len(r.adapters))
	for net := range r.adapters {
		networks = append(networks, net)
	}
	return networks
}
