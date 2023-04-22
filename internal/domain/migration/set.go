package migration

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"sync"

	"github.com/webdevelop-pro/go-common/logger"
	"github.com/webdevelop-pro/migration-service/internal/adapters"

	"github.com/pkg/errors"
)

// Set is a set of migrations for all services.
type Set struct {
	data map[int]map[string]map[int][]Migration
	repo adapters.Repository
	log  logger.Logger
	sync.Mutex
}

// New returns new instance of Set.
func New(repo adapters.Repository) *Set {
	return &Set{
		data: make(map[int]map[string]map[int][]Migration),
		repo: repo,
		log:  logger.NewDefaultComponent("migration"),
	}
}

func (s *Set) ClearData() {
	s.data = make(map[int]map[string]map[int][]Migration)
}

// ServiceExists returns true if there are known migrations for service.
func (s *Set) ServiceExists(name string) bool {
	for priority := range s.data {
		if _, exists := s.data[priority][name]; exists {
			return true
		}
	}
	return false
}

// Add adds migration to the set.
func (s *Set) Add(service string, priority, version int, mig Migration) {
	s.Lock()

	priorityService, exists := s.data[priority]
	if !exists {
		priorityService = make(map[string]map[int][]Migration)
	}

	serviceMigrations, exists := priorityService[service]
	if !exists {
		serviceMigrations = make(map[int][]Migration)
	}

	// ToDo
	// Return error if version is already taken, no version duplications
	versionMigrations, exists := serviceMigrations[version]
	if !exists {
		versionMigrations = make([]Migration, 0)
	}

	versionMigrations = append(versionMigrations, mig)
	serviceMigrations[version] = versionMigrations
	priorityService[service] = serviceMigrations
	s.data[priority] = priorityService

	s.Unlock()
}

// Services returns list of services for given priority. If priority is -1, returns services for all priorities.
func (s *Set) services(priority int) []string {
	s.Lock()
	defer s.Unlock()

	services := make([]string, 0)

	switch priority {
	case -1:
		for priority := range s.data {
			for k := range s.data[priority] {
				services = append(services, k)
			}
		}
	default:
		if priorityServices, ok := s.data[priority]; ok {
			for k := range priorityServices {
				services = append(services, k)
			}
		}
	}

	return services
}

// priorities returns list of priorities.
func (s *Set) priorities() []int {
	s.Lock()

	priorities := make([]int, len(s.data))
	i := 0

	for priority := range s.data {
		priorities[i] = priority

		i++
	}

	s.Unlock()

	sort.Ints(priorities)
	return priorities

}

// serviceMigrations returns migrations for specified service with version > minVersion.
func (s *Set) serviceMigrations(name string, priority, minVersion int) map[int][]Migration {
	migrations := make(map[int][]Migration)
	var priorities []int

	switch priority {
	case -1:
		priorities = s.priorities()
	default:
		priorities = []int{priority}
	}

	s.Lock()
	defer s.Unlock()

	for _, priority := range priorities {
		priorityMigrations, exists := s.data[priority]
		if !exists {
			continue
		}

		serviceMigrations, exists := priorityMigrations[name]
		if !exists {
			continue
		}

		for ver, m := range serviceMigrations {
			if ver <= minVersion {
				continue
			}

			if _, exists := migrations[ver]; !exists {
				migrations[ver] = make([]Migration, 0)
			}

			migrations[ver] = append(migrations[ver], m...)
		}
	}
	return migrations
}

// Apply applies migrations for specified service with version > minVersion.
func (s *Set) Apply(name string, priority, minVersion int) (int, int, error) {
	migrations := s.serviceMigrations(name, priority, minVersion)

	var n, lastVersion int
	if len(migrations) == 0 {
		return n, lastVersion, nil
	}

	versions := make([]int, len(migrations))

	i := 0

	for ver := range migrations {
		versions[i] = ver

		i++
	}
	sort.Ints(versions)

	for _, ver := range versions {
		for _, mig := range migrations[ver] {

			for _, query := range mig.Queries {
				err := s.repo.Exec(context.Background(), query)

				if err != nil && !mig.AllowError {
					return n, lastVersion, errors.Wrapf(err, "migration(%d) query failed: %s, file: %s", ver, query, mig.Path)
				}

				if err := s.repo.UpdateServiceVersion(context.Background(), name, ver); err != nil {
					return n, lastVersion, errors.Wrapf(err, "cannot update migration_service, ver: %d, file: %s", ver, mig.Path)
				}

				s.log.Info().Msgf("executed query \n%s\n for %s, version: %d, file: %s", query, name, ver, mig.Path)
			}

			lastVersion = ver
			n++
		}
	}

	return n, lastVersion, nil
}

// GetSQL returns SQL statement for specified service with version > minVersion.
func (s *Set) GetSQL(name string, priority, minVersion int) (sql string, err error) {
	migrations := s.serviceMigrations(name, priority, minVersion)

	if len(migrations) == 0 {
		return
	}

	versions := make([]int, len(migrations))

	i := 0

	for ver := range migrations {
		versions[i] = ver

		i++
	}
	sort.Ints(versions)

	for _, ver := range versions {
		for _, mig := range migrations[ver] {
			for _, query := range mig.Queries {
				sql += "\n" + strings.TrimSpace(query)
				if sql[len(sql)-1:] != ";" {
					sql += ";"
				}
			}
		}
	}

	return sql, nil
}

// ApplyAll applies all migrations for all services.
func (s *Set) ApplyAll() (int, error) {
	var n int
	lastVersions := make(map[string]int)

	pariorities := s.priorities()
	for _, priority := range pariorities {
		services := s.services(priority)
		for _, service := range services {
			ver, err := s.repo.GetServiceVersion(context.Background(), service)

			if err != nil && priority > 0 && service != "migration" {
				s.log.Error().Err(err).Msgf("failed to get service version for %s", service)
				return n, fmt.Errorf("failed to get service version for %s", service)
			}

			num, lastVersion, err := s.Apply(service, priority, ver)
			if err != nil {
				s.log.Error().Err(err).Msgf("failed to apply migrations for %s", service)
				return n, fmt.Errorf("failed to apply migrations for %s", service)
			}

			n += num
			if lastVersion > ver {
				if curLastVersion, ok := lastVersions[service]; !ok || lastVersion > curLastVersion {
					lastVersions[service] = lastVersion
				}
			}
		}
	}

	return n, nil
}
