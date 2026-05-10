package runtime

import (
	"fmt"
	"os"
	"path/filepath"
	"plugin"
)

type Loader struct {
	pluginsDir string
	plugins    map[string]AgentPlugin
}

func NewLoader(pluginsDir string) *Loader {
	return &Loader{
		pluginsDir: pluginsDir,
		plugins:    make(map[string]AgentPlugin),
	}
}

func (l *Loader) LoadAll() error {
	entries, err := os.ReadDir(l.pluginsDir)
	if err != nil {
		return fmt.Errorf("runtime: read plugins dir %s failed: %w", l.pluginsDir, err)
	}

	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".so" {
			continue
		}
		if err := l.loadFile(filepath.Join(l.pluginsDir, entry.Name())); err != nil {
			fmt.Printf("[Warning] load plugin %s failed: %v\n", entry.Name(), err)
		}
	}
	return nil
}

func (l *Loader) loadFile(soPath string) error {
	p, err := plugin.Open(soPath)
	if err != nil {
		return fmt.Errorf("plugin.Open: %w", err)
	}

	sym, err := p.Lookup(PluginSymbol)
	if err != nil {
		return fmt.Errorf("Lookup(%s): %w", PluginSymbol, err)
	}

	agent, ok := sym.(AgentPlugin)
	if !ok {
		return fmt.Errorf("plugin %s does not implement AgentPlugin interface", soPath)
	}

	info := agent.Info()
	if _, exists := l.plugins[info.Name]; exists {
		return fmt.Errorf("agent '%s' already registered", info.Name)
	}

	l.plugins[info.Name] = agent
	fmt.Printf("[Plugin] loaded: %s v%s\n", info.Name, info.Version)
	return nil
}

func (l *Loader) Get(name string) (AgentPlugin, bool) {
	p, ok := l.plugins[name]
	return p, ok
}

func (l *Loader) List() []PluginInfo {
	result := make([]PluginInfo, 0, len(l.plugins))
	for _, p := range l.plugins {
		result = append(result, p.Info())
	}
	return result
}

func (l *Loader) Count() int {
	return len(l.plugins)
}
