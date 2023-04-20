package migration

import (
	"fmt"
	"os"
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
func ReadDir(rootDir, subDir string, set *Set) error {
	files, err := os.ReadDir(filepath.Join(rootDir, subDir))
	if err != nil {
		return errors.Wrap(err, "failed to read directory")
	}

	for _, f := range files {
		if f.IsDir() {
			if err := ReadDir(rootDir, filepath.Join(subDir, f.Name()), set); err != nil {
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

		/*
			migratios/service/<sql_index>_filename.sql
			are looking for sql index ^
		*/
		if parts := strings.Split(f.Name(), "_"); len(parts) > 1 {
			if p, err := strconv.Atoi(parts[0]); err == nil {
				migrationPriority = p
			} else {
				return errors.Wrapf(err, "cannot parse %s", f.Name())
			}
		} else {
			return fmt.Errorf(
				"file %s/%s does not have correct format <sql_index>_filename.sql, dont know how to parse",
				filepath.Join(rootDir, subDir), f.Name(),
			)
		}

		/*
			migratios/<service_index>_<service_name>/<sql_index>_filename.sql
			are looking for those  ^            ^
		*/
		if folders := strings.Split(subDir, "/"); len(folders) > 0 {
			if parts := strings.Split(folders[0], "_"); len(parts) > 1 {
				if p, err := strconv.Atoi(parts[0]); err == nil {
					servicePriority = p
				}
				serviceName = folders[0][len(parts[0])+1 : len(folders[0])]
			} else {
				return fmt.Errorf(
					"file %s does not have correct format <service_index>_<service_name>, please update file name to have index and name",
					subDir,
				)
			}
		}

		fullPath := filepath.Join(rootDir, subDir, f.Name())

		/* #nosec */
		file, err := os.ReadFile(fullPath)
		if err != nil {
			return errors.Wrapf(err, "failed to open file %s", fullPath)
		}

		m := NewMigration([]string{string(file)}, fullPath)
		set.Add(serviceName, servicePriority, migrationPriority, m)
	}

	return nil
}
