package migration

import (
	"fmt"
	"sort"
	"strings"
	"sync"

	"github.com/jackc/pgx"
	"github.com/pkg/errors"
)

// Migration is a single migration.
type Migration struct {
	AllowError bool
	NoAuto     bool
	Queries    []string
}

// Set is a set of migrations for all services.
type Set struct {
	data map[int]map[string]map[int][]Migration
	pg   *pgx.ConnPool
	sync.Mutex
}

// NewSet returns new instance of Set.
func NewSet(pg *pgx.ConnPool) *Set {
	s := &Set{
		data: make(map[int]map[string]map[int][]Migration),
		pg:   pg,
	}
	return s
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
func (s *Set) Add(service string, priority, version int, migration Migration) {
	s.Lock()
	priorityService, exists := s.data[priority]
	if !exists {
		priorityService = make(map[string]map[int][]Migration)
	}
	serviceMigrations, exists := priorityService[service]
	if !exists {
		serviceMigrations = make(map[int][]Migration)
	}
	versionMigrations, exists := serviceMigrations[version]
	if !exists {
		versionMigrations = make([]Migration, 0)
	}
	versionMigrations = append(versionMigrations, migration)
	serviceMigrations[version] = versionMigrations
	priorityService[service] = serviceMigrations
	s.data[priority] = priorityService
	s.Unlock()
}

// Services returns list of services for given priority. If priority is -1, returns services for all priorities.
func (s *Set) services(priority int) []string {
	s.Lock()
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
	s.Unlock()
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
func (s *Set) Apply(name string, priority, minVersion int, isForced, noAutoOnly bool) (int, int, error) {
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
		for _, migration := range migrations[ver] {
			if !isForced && migration.NoAuto && !noAutoOnly {
				continue
			}
			if !migration.NoAuto && noAutoOnly {
				continue
			}
			for _, query := range migration.Queries {
				// Add begin / commit automatically for every query
				if strings.ToLower(query[0:6]) != "begin;" {
					query = fmt.Sprintf("BEGIN;\n%s\nCOMMIT;", query)
				}
				_, err := s.pg.Exec(query)
				if err != nil && !migration.AllowError {
					return n, lastVersion, errors.Wrapf(err, "migration(%d) query failed: %s", ver, query)
				}
			}
			lastVersion = ver
			n++
		}
	}
	return n, lastVersion, nil
}

// ApplyAll applies all migrations for all services.
func (s *Set) ApplyAll(force bool) (int, error) {
	var n int
	lastVersions := make(map[string]int)
	for _, priority := range s.priorities() {
		for _, service := range s.services(priority) {
			ver, err := s.ServiceVersion(service)
			if err != nil && priority > 0 && service != "migration" {
				return n, errors.Wrapf(err, "failed to get service version for %s", service)
			}
			num, lastVersion, err := s.Apply(service, priority, ver, force, false)
			if err != nil {
				return n, errors.Wrapf(err, "failed to apply migrations for %s", service)
			}
			n += num
			if lastVersion > ver {
				if curLastVersion, ok := lastVersions[service]; !ok || lastVersion > curLastVersion {
					lastVersions[service] = lastVersion
				}
			}
		}
	}
	for service, ver := range lastVersions {
		if ver > 0 {
			err := s.BumpServiceVersion(service, ver)
			if err != nil {
				return n, errors.Wrapf(err, "failed to bump service version (%s)", service)
			}
		}
	}
	return n, nil
}

// BumpServiceVersion updates service version.
func (s *Set) BumpServiceVersion(name string, ver int) error {
	_, err := s.pg.Exec(
		`INSERT INTO migration_service (name, version) VALUES ($1, $2) ON CONFLICT(name) DO UPDATE SET version=$2`,
		name, ver,
	)
	if err != nil {
		return errors.Wrap(err, "query failed")
	}
	return nil
}

// ServiceVersion returns currently deployed version of the service.
func (s *Set) ServiceVersion(name string) (int, error) {
	var ver int
	err := s.pg.QueryRow(`SELECT version FROM migration_service WHERE name=$1`, name).Scan(&ver)
	if err != nil {
		if err == pgx.ErrNoRows {
			return 0, nil
		}
		return 0, errors.Wrap(err, "query failed")
	}
	return ver, nil
}

/*
	insert into user_user (id, email, username, first_name, last_name, date_joined, is_superuser, is_staff)
	values ('98914f21-a534-403f-8f7e-14792c2d3577', 'cachealot@gmail.com', 'cachealot_gmail.com','vlad', 'taras', now(), true, true);
*/
