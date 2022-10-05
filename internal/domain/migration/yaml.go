package migration

import (
	"io/ioutil"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

// yamlMigration is a Migration representation in yaml.
type yamlMigration struct {
	Version    int      `yaml:"version"`
	AllowError bool     `yaml:"allowError"`
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

		if filepath.Ext(f.Name()) != ".sql" {
			continue
		}
		servicePriority := 1
		serviceName := ""
		migrationPriority := 0

		if parts := strings.Split(f.Name(), "_"); len(parts) > 1 {
			if p, err := strconv.Atoi(parts[0]); err == nil {
				migrationPriority = p
			}
		}

		if folders := strings.Split(dir, "/"); len(folders) > 1 {
			if parts := strings.Split(folders[len(folders)-1], "_"); len(parts) > 1 {
				if p, err := strconv.Atoi(parts[0]); err == nil {
					servicePriority = p
				}
				serviceName = parts[1]
			}
		}

		fullPath := filepath.Join(dir, f.Name())

		/* #nosec */
		file, err := ioutil.ReadFile(fullPath)
		if err != nil {
			return errors.Wrapf(err, "failed to open file %s", fullPath)
		}

		m := Migration{
			AllowError: false, // ToDo, move to comment
			NoAuto:     false, // ToDo, mnove to comment
			Queries:    []string{string(file)},
		}
		set.Add(serviceName, servicePriority, migrationPriority, m)
	}

	return nil
}
