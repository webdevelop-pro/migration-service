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

type migrationStats struct {
	ServicePriority   int
	ServiceName       string
	MigrationPriority int
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

		fullPath := filepath.Join(rootDir, subDir, f.Name())
		stats, err := getMigrationInfo(fullPath)
		if err != nil {
			return err
		}

		/* #nosec */
		file, err := os.ReadFile(fullPath)
		if err != nil {
			return errors.Wrapf(err, "failed to open file %s", fullPath)
		}

		m := NewMigration([]string{string(file)}, fullPath)
		set.Add(stats.ServiceName, stats.ServicePriority, stats.MigrationPriority, m)
	}

	return nil
}

// ReadFile reads migrations from file
func ReadFile(path string, set *Set) error {
	if filepath.Ext(path) != ".sql" {
		return nil
	}

	stats, err := getMigrationInfo(path)
	if err != nil {
		return err
	}

	/* #nosec */
	file, err := os.ReadFile(path)
	if err != nil {
		return errors.Wrapf(err, "failed to open file %s", path)
	}

	m := NewMigration([]string{string(file)}, path)
	set.Add(stats.ServiceName, stats.ServicePriority, stats.MigrationPriority, m)

	return nil
}

func getMigrationInfo(path string) (migrationStats, error) {
	var stats migrationStats
	fileName := filepath.Base(path)

	if parts := strings.Split(fileName, "_"); len(parts) > 1 {
		if p, err := strconv.Atoi(parts[0]); err == nil {
			stats.MigrationPriority = p
		} else {
			return stats, errors.Wrapf(err, "cannot parse %s", fileName)
		}
	} else {
		return stats, fmt.Errorf(
			"file %s does not have correct format <sql_index>_filename.sql, dont know how to parse",
			path,
		)
	}

	pathParts := strings.Split(path, "/")
	serviceFolder := pathParts[len(pathParts)-2]

	if folders := strings.Split(serviceFolder, "/"); len(folders) > 0 {
		if parts := strings.Split(folders[0], "_"); len(parts) > 1 {
			if p, err := strconv.Atoi(parts[0]); err == nil {
				stats.ServicePriority = p
			}
			stats.ServiceName = folders[0][len(parts[0])+1 : len(folders[0])]
		} else {
			return stats, fmt.Errorf(
				"folder %s does not have correct format <service_index>_<service_name>, please update file name to have index and name",
				serviceFolder,
			)
		}
	}

	return stats, nil
}
