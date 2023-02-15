package migration

import (
	"fmt"
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

		/*
			migratios/servce/<sql_index>_filename.sql
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
				"file %s does not have correct format <sql_index>_filename.sql, dont know how to parse",
				f.Name(),
			)
		}

		/*
			migratios/<service_index>_<service_name>/<sql_index>_filename.sql
			are looking for those  ^            ^
		*/
		if folders := strings.Split(dir, "/"); len(folders) > 1 {
			if parts := strings.Split(folders[1], "_"); len(parts) > 1 {
				if p, err := strconv.Atoi(parts[0]); err == nil {
					servicePriority = p
				}
				serviceName = folders[1][len(parts[0])+1 : len(folders[1])]
			} else {
				return fmt.Errorf(
					"file %s does not have correct format <service_index>_<service_name>, dont know how to parse",
					folders[1],
				)
			}
		}

		fullPath := filepath.Join(dir, f.Name())

		/* #nosec */
		file, err := ioutil.ReadFile(fullPath)
		if err != nil {
			return errors.Wrapf(err, "failed to open file %s", fullPath)
		}

		m := NewMigration([]string{string(file)})
		set.Add(serviceName, servicePriority, migrationPriority, m)
	}

	return nil
}