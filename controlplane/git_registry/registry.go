package registry

import (
	"context"
	"fmt"
)

/*
	@ayuspoudel
	At a high level:
		1. Registry interface loads guard definitions and expose them as Guard objects
		2. When load is called registry reads all manifests from disk, validates using validation rules
		3. Guard is a thin wrapper, to keep manifest intact with a name, making it searchable
		4. If a manifest is invalid loading stops and sentinel fails fast

*/

type Registry interface {
	Load(ctx context.Context) error
	Guards() []Guard
}

type Guard struct {
	Name     string
	Manifest Manifest
}

type FSRegistry struct {
	path   string
	guards []Guard
}

func NewFSRegistry(path string) *FSRegistry {
	return &FSRegistry{
		path: path,
	}
}

func (r *FSRegistry) Load(ctx context.Context) error {
	manifests, err := LoadManifests(r.path)
	if err != nil {
		return err
	}
	var guards []Guard
	for _, m := range manifests {
		err := ValidateManifest(m)
		if err != nil {
			return fmt.Errorf("invalid manifest %q: %w", m.Metadata.Name, err)
		}
		guards = append(guards, Guard{
			Name:     m.Metadata.Name,
			Manifest: m,
		})

	}
	r.guards = guards
	return nil
}

func (r *FSRegistry) Guards() []Guard {
	return r.guards
}
