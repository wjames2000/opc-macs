package platform

import (
	"fmt"
	"strings"
	"sync"
)

type Registry struct {
	mu      sync.RWMutex
	clients map[string]PlatformClient
}

func NewRegistry() *Registry {
	return &Registry{
		clients: make(map[string]PlatformClient),
	}
}

func (r *Registry) Register(client PlatformClient) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	name := strings.ToLower(client.Name())
	if _, ok := r.clients[name]; ok {
		return fmt.Errorf("registry: client %s already registered", name)
	}
	r.clients[name] = client
	return nil
}

func (r *Registry) Get(name string) (PlatformClient, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	c, ok := r.clients[strings.ToLower(name)]
	return c, ok
}

func (r *Registry) List() []PlatformClient {
	r.mu.RLock()
	defer r.mu.RUnlock()
	list := make([]PlatformClient, 0, len(r.clients))
	for _, c := range r.clients {
		list = append(list, c)
	}
	return list
}

func (r *Registry) Names() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	names := make([]string, 0, len(r.clients))
	for n := range r.clients {
		names = append(names, n)
	}
	return names
}
