package migration

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

// yamlMigration is a Migration representation in yaml.
type yamlMigration struct {
	Version    int      `yaml:"version"`
	AllowError bool     `yaml:"allowError" default:"false"`
	NoAuto     bool     `yaml:"noAuto"`
	Queries    []string `yaml:"queries"`
}

// yamlServiceSet is a set of migrations for service.
type yamlServiceSet struct {
	ServiceName string          `yaml:"service"`
	Migrations  []yamlMigration `yaml:"migrations"`
}

func (m yamlMigration) ToMigration() Migration {
	return Migration{
		AllowError: m.AllowError,
		NoAuto:     m.NoAuto,
		Queries:    m.Queries,
	}
}

// ReadDir reads migrations from all yaml files in the dir.
func ReadDir(dir string, set *Set) error {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return errors.Wrap(err, "failed to read directory")
	}
	for _, f := range files {
		if f.IsDir() {
			if err := ReadDir(filepath.Join(dir, f.Name()), set); err != nil {
				return err
			}
			continue
		}
		if filepath.Ext(f.Name()) != ".yml" && filepath.Ext(f.Name()) != ".yaml" {
			continue
		}
		priority := 1
		if parts := strings.Split(f.Name(), "_"); len(parts) > 1 {
			if p, err := strconv.Atoi(parts[0]); err == nil {
				priority = p
			}
		}
		fullPath := filepath.Join(dir, f.Name())
		/* #nosec */
		file, err := os.Open(fullPath)
		if err != nil {
			return errors.Wrapf(err, "failed to open file %s", fullPath)
		}
		var serviceSet yamlServiceSet
		err = yaml.NewDecoder(file).Decode(&serviceSet)
		if err != nil {
			_ = file.Close()
			return errors.Wrapf(err, "failed to decode migrations from file %s", fullPath)
		}
		for _, m := range serviceSet.Migrations {
			set.Add(serviceSet.ServiceName, priority, m.Version, m.ToMigration())
		}
		_ = file.Close()
	}
	return nil
}
