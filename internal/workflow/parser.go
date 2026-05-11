package workflow

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

func ParseYAML(data []byte) (*Workflow, error) {
	var wf Workflow
	if err := yaml.Unmarshal(data, &wf); err != nil {
		return nil, fmt.Errorf("workflow: parse yaml: %w", err)
	}
	if wf.Name == "" {
		return nil, fmt.Errorf("workflow: name is required")
	}
	if len(wf.Steps) == 0 {
		return nil, fmt.Errorf("workflow: at least one step is required")
	}
	for i, step := range wf.Steps {
		if step.ID == "" {
			return nil, fmt.Errorf("workflow: step %d: id is required", i)
		}
		if step.Agent == "" && len(step.Parallel) == 0 {
			return nil, fmt.Errorf("workflow: step %s: agent or parallel is required", step.ID)
		}
	}
	return &wf, nil
}

func LoadFromFile(path string) (*Workflow, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("workflow: read file: %w", err)
	}
	return ParseYAML(data)
}
