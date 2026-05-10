package harness

import "fmt"

type Artifact struct {
	SchemaName string
	Version    string
	Data       interface{}
	Raw        string
}

func ValidateArtifact(artifact *Artifact, expectedSchema string) error {
	if artifact.SchemaName != expectedSchema {
		return fmt.Errorf("harness: schema mismatch: expected %s, got %s",
			expectedSchema, artifact.SchemaName)
	}
	return nil
}
