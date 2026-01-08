package registry

import (
	"os"
	"path/filepath"

	"go.yaml.in/yaml/v2"
)

/*
Load manifests function takes a directory path as input. Under this directory
it treats everything as potential guard definitions or sentinel manifests
The function traverses the directory including nested dirs, for each file encountered
it checks if that is a yaml file. For each yaml file it loads its contents to memory,
converts it into manifest struct, parse it validate it and add to a list.
Load and evaluation is not done here.
*/
func LoadManifests(dir string) ([]Manifest, error) {
	var manifests []Manifest
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		ext := filepath.Ext(path)
		if ext != ".yaml" && ext != ".yml" {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		var manifest Manifest
		err = yaml.Unmarshal(data, &manifest)
		if err != nil {
			return err
		}
		manifests = append(manifests, manifest)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return manifests, nil
}
