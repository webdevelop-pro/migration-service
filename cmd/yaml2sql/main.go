package main

import (
	"fmt"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	"github.com/webdevelop-pro/go-common/configurator"
	"gopkg.in/yaml.v2"
)

// yamlMigration is a Migration representation in yaml.
type yamlMigration struct {
	Version    int      `yaml:"version"`
	AllowError bool     `yaml:"allow_error"`
	NoAuto     bool     `yaml:"no_auto"`
	Queries    []string `yaml:"queries"`
}

// yamlServiceSet is a set of migrations for service.
type yamlServiceSet struct {
	ServiceName string          `yaml:"service"`
	Migrations  []yamlMigration `yaml:"migrations"`
}

// ReadDir reads migrations from all yaml files in the dir.
func Migrate(inputDir string, outputDir string) error {
	files, err := os.ReadDir(inputDir)
	if err != nil {
		return errors.Wrap(err, "failed to read directory")
	}

	for _, f := range files {
		if f.IsDir() {
			newPath := fmt.Sprintf("%s/%s", outputDir, f.Name())
			err := os.Mkdir(newPath, 0755)
			if err != nil && err.Error() != fmt.Sprintf("mkdir %s: file exists", newPath) {
				return err
			}
			if err := Migrate(
				filepath.Join(inputDir, f.Name()),
				filepath.Join(outputDir, f.Name()),
			); err != nil {
				return err
			}

			continue
		}

		if filepath.Ext(f.Name()) != ".yml" && filepath.Ext(f.Name()) != ".yaml" {
			continue
		}
		priority := ""

		if parts := strings.Split(f.Name(), "_"); len(parts) > 1 {
			priority = parts[0]
			if len(priority) < 2 {
				priority = "0" + priority
			}
		}

		fullPath := filepath.Join(inputDir, f.Name())

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
			newPath := fmt.Sprintf("%s/%s_%s", outputDir, priority, serviceSet.ServiceName)
			err := os.Mkdir(newPath, 0755)
			if err != nil && err.Error() != fmt.Sprintf("mkdir %s: file exists", newPath) {
				log.Fatal(err)
			}

			f, err := os.Create(path.Join(newPath, fmt.Sprintf("%d_%s.sql", m.Version, "auto_generated")))
			if err != nil {
				log.Fatal(fmt.Errorf("failed to create file for migration(%d) for %s: %w", m.Version, serviceSet.ServiceName, err))
			}

			for _, query := range m.Queries {
				q := strings.TrimSpace(query)

				if !strings.HasSuffix(q, ";") {
					q = q + ";\n"
				} else {
					q = q + "\n"
				}

				_, err := f.WriteString(q)
				if err != nil {
					log.Fatal(fmt.Errorf("failed to write query to file for migration(%d) for %s: %w", m.Version, serviceSet.ServiceName, err))
				}
			}

			f.Close()
			// set.Add(serviceSet.ServiceName, priority, m.Version, m.ToMigration())
		}

		_ = file.Close()
	}

	return nil
}

func main() {
	config := configurator.New()
	cfg := config.New(pkgName, &Config{}, pkgName).(*Config)
	if err := Migrate(cfg.Yaml, cfg.Sql); err != nil {
		panic(err)
	}
}
